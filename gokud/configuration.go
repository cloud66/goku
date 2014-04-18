package main

import (
	"errors"
	"io/ioutil"
	"syscall"
	"time"

	"github.com/BurntSushi/toml"
)

type duration struct {
	time.Duration
}

type signal struct {
	syscall.Signal
}

type inst struct {
	Signal signal
	Wait   duration
}

var signals = map[string]syscall.Signal{
	"abrt":   syscall.SIGABRT,
	"alrm":   syscall.SIGALRM,
	"bus":    syscall.SIGBUS,
	"chld":   syscall.SIGCHLD,
	"cont":   syscall.SIGCONT,
	"fpe":    syscall.SIGFPE,
	"hup":    syscall.SIGHUP,
	"ill":    syscall.SIGILL,
	"int":    syscall.SIGINT,
	"io":     syscall.SIGIO,
	"iot":    syscall.SIGIOT,
	"kill":   syscall.SIGKILL,
	"pipe":   syscall.SIGPIPE,
	"prof":   syscall.SIGPROF,
	"quit":   syscall.SIGQUIT,
	"segv":   syscall.SIGSEGV,
	"stop":   syscall.SIGSTOP,
	"sys":    syscall.SIGSYS,
	"term":   syscall.SIGTERM,
	"trap":   syscall.SIGTRAP,
	"tstp":   syscall.SIGTSTP,
	"ttin":   syscall.SIGTTIN,
	"ttou":   syscall.SIGTTOU,
	"urg":    syscall.SIGURG,
	"usr1":   syscall.SIGUSR1,
	"usr2":   syscall.SIGUSR2,
	"vtalrm": syscall.SIGVTALRM,
	"winch":  syscall.SIGWINCH,
	"xcpu":   syscall.SIGXCPU,
	"xfsz":   syscall.SIGXFSZ,
}

type Config struct {
	Name         string
	CallbackId   string
	Tags         []string
	Command      string
	Args         []string
	Directory    string
	StopSequence []*inst
	DrainSignal  *inst
	UseEnv       bool
	Envs         []string
	AllowDrain   bool
	User         string
	Group        string
	UseStdPipe   bool
}

func ReadConfiguration(file string) (*Config, error) {
	var config *Config
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}

	if _, err := toml.Decode(string(data), &config); err != nil {
		return nil, err
	}

	return config, nil
}

func (d *duration) UnmarshalText(text []byte) error {
	var err error
	d.Duration, err = time.ParseDuration(string(text))
	return err
}

func (d *signal) UnmarshalText(text []byte) error {
	value, ok := signals[string(text)]
	if !ok {
		return errors.New("invalid signal name")
	}
	d.Signal = value
	return nil
}

func (i *inst) ToInstruction() Instruction {
	var ins = Instruction{
		Signal: i.Signal.Signal,
		Wait:   i.Wait.Duration,
	}

	return ins
}
