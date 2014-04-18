package main

import (
	"errors"
	"sync"

	"github.com/cloud66/goku/models"
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
}

func loadProcessSetFromConfig(config *Config) *ProcessSet {
	p := ProcessSet {}
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

// Starts a process in the set if possible.
func (p *ProcessSet) start() error {
	p.Lock()
	defer p.Unlock()

	if p.hasActive() {
		return errors.New("Process is already started")
	}

	// Start a new process and use as active
	proc := p.buildProcess()
	err := proc.start()
	if err != nil {
		return err
	}

	p.Active = proc

	return nil
}

// Stops the active process in the set
func (p *ProcessSet) stop() error {
	p.Lock()
	defer p.Unlock()

	if !p.hasActive() {
		return errors.New("No process is started")
	}

	err := p.Active.stop()
	if err != nil {
		return err
	}

	p.Active = nil

	return nil
}

// stops all processes in the set
func (p *ProcessSet) stopAll() []error {
	p.Lock()
	defer p.Unlock()

	var res = []error{}

	var wg sync.WaitGroup
	for _, item := range p.allProcesses() {
		wg.Add(1)
		go func(proc *Process) {
			defer wg.Done()
			err := proc.stop()
			if err != nil {
				res = append(res, err)
			}
		}(item)
	}

	wg.Wait()

	return res
}

// drains the active process and stops it in due course
func (p *ProcessSet) drain() error {
	p.Lock()
	defer p.Unlock()

	if !p.hasActive() {
		return errors.New("No process is started")
	}

	err := p.Active.drain(true)
	if err != nil {
		return err
	}

	p.Draining = append(p.Draining, p.Active)
	p.Active = nil

	return nil
}

// drain and start a new one
func (p *ProcessSet) recycle() error {
	p.Lock()
	defer p.Unlock()

	if !p.hasActive() {
		return errors.New("No process is started")
	}

	err := p.drain()
	if err != nil {
		return err
	}

	// TODO: Drain grace period

	err = p.start()
	if err != nil {
		return err
	}

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
