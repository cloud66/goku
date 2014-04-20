package main

import (
	"errors"
	"reflect"
	"sync"
	"time"

	"github.com/cloud66/goku/models"
	"github.com/golang/glog"
)

const (
	PSET_UNMONITORED = 0
	PSET_STARTING    = 1
	PSET_UP          = 2
	PSET_STOPPING    = 3
)

type ProcessSet struct {
	Active       *Process
	Draining     []*Process
	Name         string
	CallbackId   string
	Tags         []string
	Command      string
	Args         []string
	Directory    string
	StopSequence []Instruction
	DrainSignal  Instruction
	UseEnv       bool
	Envs         []string
	AllowDrain   bool
	User         string
	Group        string
	UseStdPipe   bool

	sync.Mutex
	config *Config
}

// creates a new ProcessSet from a Config
func loadProcessSetFromConfig(config *Config) *ProcessSet {
	p := ProcessSet{}
	p.Name = config.Name
	p.CallbackId = config.CallbackId
	p.Tags = config.Tags
	p.Command = config.Command
	p.Directory = config.Directory
	p.UseEnv = config.UseEnv
	p.Envs = config.Envs
	p.AllowDrain = config.AllowDrain
	p.User = config.User
	p.Group = config.Group
	p.UseStdPipe = config.UseStdPipe
	p.config = config

	if config.DrainSignal != nil {
		p.DrainSignal = config.DrainSignal.ToInstruction()
	}

	if len(config.Args) != 0 {
		p.Args = config.Args
	}

	if len(config.StopSequence) != 0 {
		var stopSequences []Instruction
		for _, stopSequence := range config.StopSequence {
			stopSequences = append(stopSequences, stopSequence.ToInstruction())
		}
		p.StopSequence = stopSequences
	}

	return &p
}

func (p *ProcessSet) reload() error {
	glog.Infof("Reloading configuration for %s", p.Name)

	err := p.config.reload()
	if err != nil {
		return err
	}

	config := p.config

	p.CallbackId = config.CallbackId
	p.Tags = config.Tags
	p.AllowDrain = config.AllowDrain
	p.UseStdPipe = config.UseStdPipe

	if config.DrainSignal != nil {
		p.DrainSignal = config.DrainSignal.ToInstruction()
	}

	if len(config.Args) != 0 {
		p.Args = config.Args
	}

	if len(config.StopSequence) != 0 {
		var stopSequences []Instruction
		for _, stopSequence := range config.StopSequence {
			stopSequences = append(stopSequences, stopSequence.ToInstruction())
		}
		p.StopSequence = stopSequences
	}

	// any of these changes means a restart
	if (p.Command != config.Command) ||
		(p.Name != config.Name) ||
		(p.Directory != config.Directory) ||
		(p.UseEnv != config.UseEnv) ||
		(!reflect.DeepEqual(p.Envs, config.Envs)) ||
		(p.User != config.User) ||
		(p.Group != config.Group) {
		glog.Infof("Restarting configuration change detected. Restarting %s", p.Name)
		p.Command = config.Command
		p.Directory = config.Directory
		p.UseEnv = config.UseEnv
		p.Envs = config.Envs
		p.User = config.User
		p.Group = config.Group
		p.Name = config.Name

		return p.restart()
	}

	return nil
}

// Starts a process in the set if possible.
func (p *ProcessSet) start() error {
	p.Lock()
	defer p.Unlock()

	return p.doStart()
}

// Starts a process in the set if possible.
// not thread safe. Always call start instead
func (p *ProcessSet) doStart() error {
	if p.hasActive() {
		glog.Errorf("Process %s is already started", p.Name)
		return errors.New("Process is already started")
	}

	glog.Infof("Starting %s", p.Name)

	// Start a new process and use as active
	proc := p.buildProcess()
	err := proc.start()
	if err != nil {
		glog.Errorf("Starting %s failed due to %v", p.Name, err)
		return err
	}

	p.Active = proc

	glog.Infof("Process %s started", p.Name)

	return nil
}

// Stops the active process in the set
func (p *ProcessSet) stop() error {
	p.Lock()
	defer p.Unlock()

	if !p.hasActive() {
		glog.Errorf("Process %s is not started", p.Name)
		return errors.New("No process is started")
	}

	glog.Infof("Stopping %s", p.Name)

	err := p.Active.stop()
	if err != nil {
		glog.Errorf("Failed to stop process %s due to %v", p.Name, err)
		return err
	}

	p.Active = nil

	glog.Infof("Process %s stopped", p.Name)

	return nil
}

// kills the old set and starts a new one
func (p *ProcessSet) restart() error {
	glog.Infof("Stopping the process set %s", p.Name)
	errs := p.stopAll()
	if len(errs) != 0 {
		glog.Errorf("Failed to stop all members of the set %s due to %v", p.Name, errs)
		return errors.New("Failed to stop all members of the set")
	}

	return p.start()
}

// stops all processes in the set
func (p *ProcessSet) stopAll() []error {
	var res = []error{}

	glog.Infof("Stopping all processes under %s", p.Name)

	var wg sync.WaitGroup
	for _, item := range p.allProcesses() {
		wg.Add(1)
		go func(proc *Process) {
			defer wg.Done()
			err := proc.stop()
			if proc == p.Active {
				p.Active = nil
			}
			if err != nil {
				glog.Errorf("Failed to stop process %s (%s) due to %v", proc.Name, proc.Uid, err)
				res = append(res, err)
			}
		}(item)
	}
	wg.Wait()

	glog.Infof("All processes under %s stopped", p.Name)

	return res
}

// drain and start a new one
func (p *ProcessSet) recycle() error {
	p.Lock()
	defer p.Unlock()

	if !p.hasActive() {
		glog.Errorf("Process %s is not started", p.Name)
		return errors.New("No process is started")
	}

	glog.Infof("Recycling process %s", p.Name)

	// this part is async since we need to have the new one start immediately
	go func(proc *ProcessSet, active *Process) {
		glog.Infof("Waiting for %s (%s) to drain", proc.Name, active.Uid)
		err := active.drain(true)
		if err != nil {
			glog.Errorf("Failed to drain %s (%s) due to %v", proc.Name, active.Uid, err)
		}
	}(p, p.Active)

	p.Draining = append(p.Draining, p.Active)
	p.Active = nil

	// wait for drain signals
	time.Sleep(1 * time.Second)

	err := p.doStart()
	if err != nil {
		return err
	}

	glog.Infof("Process %s recycled. New active UID is %s", p.Name, p.Active.Uid)

	return nil
}

// should not be called outside of a p.lock
func (p *ProcessSet) hasActive() bool {
	return p.Active != nil
}

// should not be called outside of a p.lock
func (p *ProcessSet) allProcesses() []*Process {
	return append(p.Draining, p.Active)
}

func (p *ProcessSet) buildProcess() *Process {
	return &Process{
		Name:         p.Name,
		CallbackId:   p.CallbackId,
		Tags:         p.Tags,
		Command:      p.Command,
		Args:         p.Args,
		Directory:    p.Directory,
		StopSequence: p.StopSequence,
		DrainSignal:  p.DrainSignal,
		UseEnv:       p.UseEnv,
		Envs:         p.Envs,
		AllowDrain:   p.AllowDrain,
		User:         p.User,
		Group:        p.Group,
		UseStdPipe:   p.UseStdPipe,
		processSet:   p,
	}
}

func (c *ProcessSet) toCtrlProcessSet() models.CtrlProcessSet {
	ctrlProcessSet := models.CtrlProcessSet{
		Name:       c.Name,
		CallbackId: c.CallbackId,
		Tags:       c.Tags,
		Command:    c.Command,
		Args:       c.Args,
		Directory:  c.Directory,
		UseEnv:     c.UseEnv,
		Envs:       c.Envs,
		AllowDrain: c.AllowDrain,
		User:       c.User,
		Group:      c.Group,
		UseStdPipe: c.UseStdPipe,
	}

	for _, process := range c.Draining {
		ctrlProcessSet.Draining = append(ctrlProcessSet.Draining, process.toCtrlProcess())
	}

	if c.hasActive() {
		ctrlProcSet := c.Active.toCtrlProcess()
		ctrlProcessSet.Active = &ctrlProcSet
	}

	return ctrlProcessSet
}

func (c *ProcessSet) listenToProcessEvents(events chan *Process) {
	for {
		select {
		case r := <-events:
			glog.V(Debug).Infof("Event received from %s (%s)", r.Name, r.Uid)

			if r.statusCode == PS_UNMONITORED {
				glog.V(Detail).Infof("Process %s stopped. Removing from the draining list", r.Uid)
				// if the process is stopped, then take it of the drained list
				c.removeDrained(r)
			}
		case <-time.After(50 * time.Millisecond):
			// move along
		}
	}
}

func (c *ProcessSet) removeDrained(toRemove *Process) {
	c.Lock()
	defer c.Unlock()

	newList := []*Process{}
	for _, p := range c.Draining {
		glog.V(Verbose).Infof("Comparing '%s' with '%s' to remove", p.Uid, toRemove.Uid)
		if p.Uid != toRemove.Uid {
			newList = append(newList, p)
		}
	}

	c.Draining = newList
}
