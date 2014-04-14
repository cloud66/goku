package main

import (
	"errors"
	"sync"
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

	sync.Mutex
}

// Starts a process in the set if possible.
func (p *ProcessSet) Start() error {
	p.Lock()
	defer p.Unlock()

	if !p.hasActive() {
		return errors.New("Process is already started")
	}

	// Start a new process and use as active
	proc := p.buildProcess()
	err := proc.Start()
	if err != nil {
		return err
	}

	p.Active = proc

	return nil
}

// Stops the active process in the set
func (p *ProcessSet) Stop() error {
	p.Lock()
	defer p.Unlock()

	if !p.hasActive() {
		return errors.New("No process is started")
	}

	err := p.Active.Stop()
	if err != nil {
		return err
	}

	p.Active = nil

	return nil
}

// stops all processes in the set
func (p *ProcessSet) StopAll() []error {
	p.Lock()
	defer p.Unlock()

	var res = []error{}

	var wg sync.WaitGroup
	for _, item := range p.allProcesses() {
		wg.Add(1)
		go func(proc *Process) {
			defer wg.Done()
			err := proc.Stop()
			if err != nil {
				res = append(res, err)
			}
		}(item)
	}

	wg.Wait()

	return res
}

// drains the active process and stops it in due course
func (p *ProcessSet) Drain() error {
	p.Lock()
	defer p.Unlock()

	if !p.hasActive() {
		return errors.New("No process is started")
	}

	err := p.Active.Drain(true)
	if err != nil {
		return err
	}

	p.Draining = append(p.Draining, p.Active)
	p.Active = nil

	return nil
}

// drain and start a new one
func (p *ProcessSet) Recycle() error {
	p.Lock()
	defer p.Unlock()

	if !p.hasActive() {
		return errors.New("No process is started")
	}

	err := p.Drain()
	if err != nil {
		return err
	}

	// TODO: Drain grace period

	err = p.Start()
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
	}
}
