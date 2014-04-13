package main

import (
	"os"
	"os/exec"
	"time"
	"strconv"
	"syscall"
	"path/filepath"

	"github.com/golang/glog"
	"github.com/nu7hatch/gouuid"
)

const (
	LogFolder = "/tmp/goku/logs"
	PidFolder = "/tmp/goku/pids"
)

type Instruction struct {
	Signal		os.Signal
	Wait			time.Duration
}

type Process struct {
	Name					string
	Uid			 		string
	CallbackId		string
	Tags					[]string
	Command		 	string
	Args					[]string
	Directory		 string
	StopSequence	[]Instruction
	DrainSequence []Instruction
	UseEnv				bool
	Envs					[]string
	AllowDrain		bool
	Pid					 int

	x						 *os.Process
	timestamp		 int64
	cmd					 *exec.Cmd
}

func (p *Process) Start() error {
	// make sure the needed folders are there
	os.MkdirAll(LogFolder, 0777)

	// mark the start
	p.timestamp = time.Now().Unix()
	uid, err := uuid.NewV4()
	if err != nil {
		return err
	}

	p.Uid = uid.String()
	glog.Infof("Starting process '%s' Timestamp: %d Uid:%s", p.Name, p.timestamp, p.Uid)

	// now start it
	err = p.startProcessByExec()
	if err != nil {
		return err
	}

	go p.waitForProcess()

	return nil
}

func (p *Process) startProcessByExec() error {
	var envs []string
	if p.UseEnv {
		envs = os.Environ()
	} else {
		envs = p.Envs
	}

	outLog := filepath.Join(LogFolder, p.Name + "_stdout_" + strconv.FormatInt(p.timestamp, 10) + ".log")
	errLog := filepath.Join(LogFolder, p.Name + "_stderr_" + strconv.FormatInt(p.timestamp, 10) + ".log")
	outLogFile, err := getLogfile(outLog)
	if err != nil {
		return err
	}
	errLogFile, err := getLogfile(errLog)
	if err != nil {
		return err
	}

	cmd := exec.Cmd{
		Path: p.Command,
		Args: append([]string{p.Command}, p.Args...),
		Dir: p.Directory,
		Stdin: os.Stdin,
		Stdout: outLogFile,
		Stderr: errLogFile,
		Env: envs,
	}
	err = cmd.Start()
	if err != nil {
		return err
	}

	p.Pid = cmd.Process.Pid
	p.x = cmd.Process
	p.cmd = &cmd

	glog.Infof("Process '%s' started. Pid: %d", p.Name, p.Pid)

	return nil
}

func (p *Process) waitForProcess() {
	glog.Infof("Watching close of process '%s'", p.Name)

	p.cmd.Process.Wait()
	p.cmd.Process.Kill()
	p.cmd.Process.Release()

	glog.Infof("Process '%s' closed.", p.Name)
}

func (p *Process) IsRunning() bool {
	if err := syscall.Kill(p.Pid, 0); err != nil {
		return false
	} else {
		return true
	}
}

func getLogfile(path string) (*os.File, error) {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0660)
	if err != nil {
		return nil, err
	}

	return file, nil
}
