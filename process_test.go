package main

import (
	"testing"
)

func TestStartProcess(t *testing.T) {
	p := Process{
		Name: "test",
		Directory: "/bin",
		Command: "sleep",
		Args: []string{ "0" },
	}

	err := p.Start()
	if err != nil {
		t.Error(err)
	}
}
