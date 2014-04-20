package main

import (
	"errors"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/cloud66/goku/models"
	"github.com/golang/glog"
	"github.com/nu7hatch/gouuid"
)

const (
	LogFolder        = "/tmp/goku/logs"
	PidFolder        = "/tmp/goku/pids"
	MAX_START_COUNTS = 5

	// not running or we are unaware of the status
	PS_UNMONITORED = 0
	// unknown status. an error for example during a transition
	PS_UNKNOWN = 1
	// process is starting. it is not completely functional yet
	PS_STARTING = 2
	// process is up. all good
	PS_UP = 3
	// process is due to be stopped intentinally
	PS_STOPPING = 4
	// process is stopped unintentionally
	PS_STOPPED = 5
	// process is draining and will stop eventually
	PS_DRAINING = 6
)

var statusMap = map[int]string{
	PS_UNMONITORED: "unmonitored",
	PS_UNKNOWN:     "unknown",
	PS_STARTING:    "starting",
	PS_UP:          "up",
	PS_STOPPING:    "stopping",
	PS_DRAINING:    "draining",
}

type Instruction struct {
	Signal os.Signal
	Wait   time.Duration
}

var (
	defaultStopSequence = []Instruction{
		{Signal: syscall.SIGQUIT, Wait: 5},
		{Signal: syscall.SIGKILL, Wait: 0},
	}
	startLock sync.Mutex
)

type Process struct {
	Name       string
	Uid        string
	CallbackId string
	Tags       []string
	Command    string
	Args       []string
	Directory  string
	// signals to send and wait expecting the process to die
	StopSequence []Instruction
	// signal to start the drain.
	// we then wait until we start the stop sequence
	DrainSignal Instruction
	UseEnv      bool
	Envs        []string
	AllowDrain  bool
	Pid         int
	User        string
	Group       string
	UseStdPipe  bool

	LastActionAt time.Time

	x           *os.Process
	timestamp   int64
	cmd         *exec.Cmd
	pidfile     Pidfile
	userId      int
	groupId     int
	statusCode  int
	processSet  *ProcessSet
	startCount  int
	dontRecover bool
}

func (p *Process) status() (int, string) {
	return p.statusCode, statusMap[p.statusCode]
}

func (p *Process) setStatus(newStatus int) {
	glog.V(Detail).Infof("Process %s (%s) status is chaning %s -> %s", p.Name, p.Uid, statusMap[p.statusCode], statusMap[newStatus])
	p.statusCode = newStatus
	p.LastActionAt = time.Now()
}

func (p *Process) start() error {
	if p.Pid != 0 && p.isRunning() {
		glog.Infof("Process %s (%s) is already running", p.Name, p.Uid)
		return errors.New("Process is already running")
	}

	p.setStatus(PS_STARTING)

	// make sure the needed folders are there
	os.MkdirAll(LogFolder, 0777)
	os.MkdirAll(PidFolder, 0777)

	// mark the start
	p.timestamp = time.Now().Unix()
	uid, err := uuid.NewV4()
	if err != nil {
		p.setStatus(PS_UNKNOWN)
		return err
	}

	p.Uid = uid.String()
	glog.Infof("Starting process '%s' Timestamp: %d Uid:%s", p.Name, p.timestamp, p.Uid)

	if len(p.StopSequence) == 0 {
		glog.Info("Will use the default StopSequence for stop")
		p.StopSequence = defaultStopSequence
	}

	// now start it
	go func() {
		err := p.startProcessByExec()
		if err != nil {
			p.setStatus(PS_UNKNOWN)
			glog.Errorf("Failed to start %s (%s) due to %v", p.Name, p.Uid, err)
			return
		}
		p.setStatus(PS_UP)
		go p.waitForProcess()
	}()

	return nil
}

func (p *Process) stop() error {
	if !p.isRunning() {
		glog.Infof("Process %s (%s) is not running. Can't stop it.", p.Name, p.Uid)

		return errors.New("Process is not running")
	}
	p.setStatus(PS_STOPPING)

	for _, item := range p.StopSequence {
		glog.Infof("Sending %s to %d", item.Signal, p.Pid)
		err := p.sendSignalAndWait(item)
		if err != nil {
			p.setStatus(PS_UNKNOWN)
			return err
		}

		// is it running?
		if !p.isRunning() {
			p.setStatus(PS_UNMONITORED)

			glog.Infof("Process '%s' (%s) stopped", p.Name, p.Uid)
			return nil
		}
	}

	// still running? use force
	if p.isRunning() {
		glog.Infof("Process '%s' (%s) still running trying force", p.Name, p.Uid)
		syscall.Kill(p.Pid, syscall.SIGKILL)
		time.Sleep(100 * time.Millisecond)
		p.x.Release()
	}

	// still running?
	if p.isRunning() {
		p.setStatus(PS_UNKNOWN)

		return errors.New("cannot stop the process")
	}

	p.setStatus(PS_UNMONITORED)

	return nil
}

// sends a drain signal to the process.
// it can stop the process in due course (DrainSignal.Wait) if needed
// it waits for the drain and then stops the process by calling stop
func (p *Process) drain(stop bool) error {
	if !p.isRunning() {
		glog.Infof("Process %s (%s) is not running. Can't drain it.", p.Name, p.Uid)
		return errors.New("Process is not running")
	}

	p.setStatus(PS_DRAINING)

	// move the pid file to a timestamped one
	newPidfile := filepath.Join(PidFolder, p.Name+"_"+strconv.FormatInt(p.timestamp, 10)+".pid")
	err := os.Rename(string(p.pidfile), newPidfile)
	if err != nil {
		return err
	}
	p.pidfile = Pidfile(newPidfile)

	err = p.sendDrainSignal()
	if err != nil {
		p.setStatus(PS_UNKNOWN)
		return err
	}

	if stop {
		time.Sleep(p.DrainSignal.Wait)

		err := p.stop()
		if err != nil {
			p.setStatus(PS_UNKNOWN)
			return err
		}
	}

	return nil
}

func (p *Process) isRunning() bool {
	if err := syscall.Kill(p.Pid, 0); err != nil {
		glog.V(Debug).Infof("Looking for process with pid %d. %s", p.Pid, err.Error())
		return false
	} else {
		p.setStatus(PS_UNMONITORED)
		return true
	}
}

// send the drain signal
func (p *Process) sendDrainSignal() error {
	err := p.sendSignal(p.DrainSignal.Signal)
	if err != nil {
		p.setStatus(PS_UNKNOWN)

		return err
	}

	return nil
}

func (p *Process) sendSignalAndWait(instruction Instruction) error {
	// send
	err := p.sendSignal(instruction.Signal)
	if err != nil {
		return err
	}

	// wait
	time.Sleep(instruction.Wait * time.Second)

	return nil
}

func (p *Process) sendSignal(signal os.Signal) error {
	glog.V(Detail).Infof("Sending %s to %s (%s)", signal, p.Name, p.Uid)

	err := p.x.Signal(signal)
	if err != nil {
		return err
	}

	return nil
}

func (p *Process) startProcessByExec() error {
	var envs []string
	if p.UseEnv {
		envs = os.Environ()
	} else {
		envs = p.Envs
	}

	// find the executable
	fullPath, err := exec.LookPath(p.Command)
	if err != nil {
		fullPath = p.Command
	}
	glog.V(Detail).Infof("Found '%s' here '%s'", p.Command, fullPath)

	// find the user/group
	if p.User != "" {
		uuid, err := user.Lookup(p.User)
		if err != nil {
			return err
		}
		uid, err := strconv.Atoi(uuid.Uid)
		if err != nil {
			return err
		}
		p.userId = uid
	}

	if p.Group != "" {
		gid, err := lookupGroupId(p.Group)
		if err != nil {
			return err
		}
		p.groupId = gid
	}

	outLog := filepath.Join(LogFolder, p.Name+"_"+strconv.FormatInt(p.timestamp, 10)+"_stdout.log")
	errLog := filepath.Join(LogFolder, p.Name+"_"+strconv.FormatInt(p.timestamp, 10)+"_stderr.log")
	outLogFile, err := getLogfile(outLog)
	if err != nil {
		return err
	}
	errLogFile, err := getLogfile(errLog)
	if err != nil {
		return err
	}
	// this is the active process, so the pid will be the same
	p.pidfile = Pidfile(filepath.Join(PidFolder, p.Name+".pid"))

	if len(p.Args) == 0 {
		p.Args = []string{}
	}

	cmd := exec.Cmd{
		Path:  fullPath,
		Args:  append([]string{p.Command}, p.Args...),
		Dir:   p.Directory,
		Env:   envs,
		Stdin: os.Stdin,
	}

	if !p.UseStdPipe {
		cmd.Stdout = outLogFile
		cmd.Stderr = errLogFile
	} else {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}

	if p.userId != 0 || p.groupId != 0 {
		startLock.Lock()
		defer startLock.Unlock()
		// switch the user
		cuid := syscall.Getuid()
		cgid := syscall.Getgid()
		defer func() {
			syscall.Setuid(cuid)
			syscall.Setgid(cgid)
		}()
	}

	if p.userId != 0 {
		if err = syscall.Setuid(p.userId); err != nil {
			return err
		}
	}
	if p.groupId != 0 {
		if err = syscall.Setgid(p.groupId); err != nil {
			return err
		}
	}

	glog.V(Verbose).Infof("Calling exec command for %s", cmd)
	glog.V(Detail).Infof("Calling exec for command '%s' in '%s' with %s", cmd.Path, cmd.Dir, cmd.Args)
	err = cmd.Start()
	if err != nil {
		return err
	}
	glog.V(Detail).Infof("Process '%s' (%s) started", cmd.Path, p.Uid)

	p.Pid = cmd.Process.Pid
	p.pidfile.write(p.Pid)
	p.x = cmd.Process
	p.cmd = &cmd

	glog.Infof("Process '%s' (%s) started. Pid: %d", p.Name, p.Uid, p.Pid)

	return nil
}

func (p *Process) waitForProcess() {
	glog.Infof("Watching close of process '%s' (%s)", p.Name, p.Uid)

	p.cmd.Process.Wait()

	// proces is closed. was it an accident? have we not tried enough?
	if p.statusCode != PS_STOPPING && p.startCount < MAX_START_COUNTS && !p.dontRecover {
		p.setStatus(PS_STOPPED)
		glog.Infof("Unintentional stop detected for %s (%s). Trying to recover attempt %d", p.Name, p.Uid, p.startCount)

		// start it again
		p.startCount++
		err := p.start()
		if err != nil {
			glog.Errorf("Failed to recover process %s (%s) from an unintentional stop, due to %s", p.Name, p.Uid, err.Error())
		}

		return
	}

	p.cmd.Process.Kill()
	p.cmd.Process.Release()

	p.pidfile.delete()

	p.setStatus(PS_UNMONITORED)

	if p.processSet != nil {
		p.processSet.removeDrained(p)
	}

	p.startCount = 0
	glog.Infof("Process '%s' (%s) closed.", p.Name, p.Uid)
}

func getLogfile(path string) (*os.File, error) {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0660)
	if err != nil {
		return nil, err
	}

	return file, nil
}

// tries to kill a stuck process gracefully.
// only
func (p *Process) sunset() error {
	glog.Infof("[SUNSET] Sunsetting process %s with pid %d", p.Name, p.Pid)
	if p.AllowDrain {
		glog.Infof("[SUNSET] Process %s allows draining. Trying draining first", p.Name)
		p.sendSignal(p.DrainSignal.Signal)
		time.Sleep(p.DrainSignal.Wait)
	}

	if p.isRunning() {
		glog.Infof("[SUNSET] Process %s is still running after drainig. Trying the stop sequence", p.Name)
		for _, item := range p.StopSequence {
			glog.V(Detail).Infof("[SUNSET] Sending %s to pid %d", item.Signal, p.Pid)
			err := p.sendSignalAndWait(item)
			if err != nil {
				return err
			}

			// is it running?
			if !p.isRunning() {
				return nil
			}
		}

		// still running? use force
		if p.isRunning() {
			glog.Infof("[SUNSET] Process %s still running. Trying force", p.Name)
			syscall.Kill(p.Pid, syscall.SIGKILL)
			time.Sleep(100 * time.Millisecond)
			p.x.Release()
		}
	}

	return nil
}

func (p *Process) toCtrlProcess() models.CtrlProcess {
	ctrlProcess := models.CtrlProcess{
		Uid:          p.Uid,
		Pid:          p.Pid,
		LastActionAt: p.LastActionAt,
		TimeStamp:    p.timestamp,
	}

	ctrlProcess.Status.Code, ctrlProcess.Status.Message = p.status()

	return ctrlProcess
}
