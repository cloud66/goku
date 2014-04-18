package models

import (
	"fmt"
)

type StatusTuple struct {
	Code			int
	Message	 string
}

type CtrlProcessSet struct {
	Name         string
	CallbackId   string
	Tags         []string
	Command      string
	Args         []string
	Directory    string
	UseEnv       bool
	Envs         []string
	AllowDrain   bool
	User         string
	Group        string
	UseStdPipe   bool
	Draining	   []CtrlProcess
	Active			 *CtrlProcess
}

func (c *CtrlProcessSet) Status() string {
	return fmt.Sprintf("%s %s", c.ActiveStatus(), c.DrainingStatus())
}

// returns the status of the child processes
func (c *CtrlProcessSet) DrainingStatus() string {
	if len(c.Draining) == 0 {
		return ""
	}

	return fmt.Sprintf("(%d draining)", len(c.Draining))
}

func (c *CtrlProcessSet) ActiveStatus() string {
	if c.Active == nil {
		return "unmonitored"
	}

	return c.Active.Status.Message
}
