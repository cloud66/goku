package models

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
}
