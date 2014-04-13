package main

import (
	"testing"
	"os"
	"time"
	"flag"
	"syscall"
)

func init() {
	flag.Set("alsologtostderr", "true")

	os.RemoveAll(LogFolder)
}

func TestStart(t *testing.T) {
	p := Process{
		Name: "TestStart",
		Directory: "/bin",
		Command: "sleep",
		Args: []string{ "1" },
	}

	err := p.Start()
	if err != nil {
		t.Error(err)
	}

	if !p.IsRunning() {
		t.Error("Process not running")
	}

	time.Sleep(1100 * time.Millisecond)

	if p.IsRunning() {
		t.Error("Process is running")
	}
}

func TestSimpleStop(t *testing.T) {
	p := Process{
		Name: "TestSimpleStop",
		Directory: "tests",
		Command: "stops_with_quit.sh",
	}

	err := p.Start()
	if err != nil {
		t.Error(err)
	}

	if !p.IsRunning() {
		t.Error("Process not running")
	}

	err = p.Stop()
	if err != nil {
		t.Error(err)
	}

	if p.IsRunning() {
		t.Error("Process is still running")
	}
}

func TestTermToStop(t *testing.T) {
	p := Process{
		Name: "TestTermToStop",
		Directory: "tests",
		Command: "stops_with_term.sh",
		StopSequence: []Instruction{
			{ Signal: syscall.SIGQUIT, Wait: 1 },
			{ Signal: syscall.SIGTERM, Wait: 1 },
		},
	}

	err := p.Start()
	if err != nil {
		t.Error(err)
	}

	if !p.IsRunning() {
		t.Error("Process not running")
	}

	time.Sleep(100 * time.Millisecond)

	err = p.Stop()
	if err != nil {
		t.Error(err)
	}

	if p.IsRunning() {
		t.Error("Process is still running")
	}
}

func TestForceToStop(t *testing.T) {
	p := Process{
		Name: "TestForceToStop",
		Directory: "tests",
		Command: "stops_with_none.sh",
	}

	err := p.Start()
	if err != nil {
		t.Error(err)
	}

	if !p.IsRunning() {
		t.Error("Process not running")
	}

	err = p.Stop()
	if err != nil {
		t.Error(err)
	}

	if p.IsRunning() {
		t.Error("Process is still running")
	}
}
