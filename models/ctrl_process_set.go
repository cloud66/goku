package models

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
	Status			 StatusTuple
	Draining	   []CtrlProcess
	Active			 CtrlProcess
}
