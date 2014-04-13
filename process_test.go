package main

import (
	"testing"
	"os"
	"time"
	"flag"
)

func init() {
	flag.Set("alsologtostderr", "true")

	os.RemoveAll(LogFolder)
}

func TestStart(t *testing.T) {
	p := Process{
		Name: "test",
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
		Name: "test",
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
