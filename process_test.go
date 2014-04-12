package main

import (
	"testing"
	"os"
	"time"
)

func init() {
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

	proc, err := os.FindProcess(p.Pid)
	if err != nil {
		t.Error(err)
	}
	if proc.Pid == 0 {
		t.Error("No process found")
	}

	time.Sleep(15 * time.Second)

	proc, err = os.FindProcess(p.Pid)
	if err != nil {
		t.Error(err)
	}
	if proc.Pid == 0 {
		t.Error("No process found")
	}
}

func TestStartWithDirectory(t *testing.T) {
	p := Process{
		Name: "test1",
		Directory: "tests",
		Command: "test.sh",
		Args: []string{ "abc" },
	}

	err := p.Start()
	if err != nil {
		t.Error(err)
	}
}
