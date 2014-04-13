package main

import (
	"testing"
	"os"
	"time"
	"flag"
)

func init() {
	os.RemoveAll(LogFolder)
}

func TestStart(t *testing.T) {
	flag.Set("alsologtostderr", "true")

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
