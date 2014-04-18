package main

import (
	"flag"
	"os"
	"syscall"
	"testing"
	"time"
)

func init() {
	flag.Set("alsologtostderr", "true")

	os.RemoveAll(LogFolder)
	os.RemoveAll(PidFolder)
}

func TestStart(t *testing.T) {
	p := Process{
		Name:      "TestStart",
		Directory: "/bin",
		Command:   "sleep",
		Args:      []string{"1"},
	}

	err := p.start()
	if err != nil {
		t.Error(err)
	}

	if !p.isRunning() {
		t.Error("Process not running")
	}

	time.Sleep(1100 * time.Millisecond)

	if p.isRunning() {
		t.Error("Process is running")
	}
}

func TestSimpleStop(t *testing.T) {
	p := Process{
		Name:      "TestSimpleStop",
		Directory: "tests",
		Command:   "stops_with_quit.sh",
	}

	err := p.start()
	if err != nil {
		t.Error(err)
	}

	if !p.isRunning() {
		t.Error("Process not running")
	}

	err = p.stop()
	if err != nil {
		t.Error(err)
	}

	if p.isRunning() {
		t.Error("Process is still running")
	}
}

func TestTermToStop(t *testing.T) {
	p := Process{
		Name:      "TestTermToStop",
		Directory: "tests",
		Command:   "stops_with_term.sh",
		StopSequence: []Instruction{
			{Signal: syscall.SIGQUIT, Wait: 1},
			{Signal: syscall.SIGTERM, Wait: 1},
		},
	}

	err := p.start()
	if err != nil {
		t.Error(err)
	}

	if !p.isRunning() {
		t.Error("Process not running")
	}

	time.Sleep(100 * time.Millisecond)

	err = p.stop()
	if err != nil {
		t.Error(err)
	}

	if p.isRunning() {
		t.Error("Process is still running")
	}
}

func TestForceToStop(t *testing.T) {
	p := Process{
		Name:      "TestForceToStop",
		Directory: "tests",
		Command:   "stops_with_none.sh",
	}

	err := p.start()
	if err != nil {
		t.Error(err)
	}

	if !p.isRunning() {
		t.Error("Process not running")
	}

	err = p.stop()
	if err != nil {
		t.Error(err)
	}

	if p.isRunning() {
		t.Error("Process is still running")
	}
}

func TestStatus(t *testing.T) {
	p := Process{
		Name:      "TestStatus",
		Directory: "tests",
		Command:   "stops_with_term.sh",
	}

	if p.statusCode != PS_UNMONITORED {
		t.Errorf("Status is not unmonitored (%s)", p.status())
	}

	err := p.start()
	if err != nil {
		t.Error("Failed to start")
	}

	if p.statusCode != PS_UP {
		t.Errorf("Status is not up (%s)", p.status())
	}

	go p.stop()

	time.Sleep(100 * time.Millisecond)

	if p.statusCode != PS_STOPPING {
		t.Errorf("Status is not stopping (%s)", p.status())
	}

	time.Sleep(100 * time.Millisecond)

	if p.statusCode != PS_UNMONITORED {
		t.Errorf("Status is not unmonitored (%s)", p.status())
	}
}
