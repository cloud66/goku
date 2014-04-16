package main

import (
	"reflect"
	"syscall"
	"testing"
)

func TestSimpleConfig(t *testing.T) {
	conf, err := ReadConfiguration("tests/simple.toml")
	if err != nil {
		t.Error(err)
	}

	if conf.Name != "Simple" {
		t.Errorf("Name not loaded: %s", conf.Name)
	}
	if conf.CallbackId != "some-callback" {
		t.Errorf("CallbackId not loaded: %s", conf.CallbackId)
	}
	if !reflect.DeepEqual(conf.Tags, []string{"web", "db"}) {
		t.Errorf("Tags not loaded: %s", conf.Tags)
	}
	if conf.Command != "sleep" {
		t.Errorf("Command not loaded: %s", conf.Command)
	}
	if !reflect.DeepEqual(conf.Args, []string{"1"}) {
		t.Errorf("Args not loaded: %s", conf.Args)
	}
	if conf.Directory != "/bin" {
		t.Errorf("Directory not loaded: %s", conf.Directory)
	}
	if !conf.UseEnv {
		t.Errorf("UseEnv not loaded: %s", conf.UseEnv)
	}
	if conf.AllowDrain {
		t.Errorf("AllowDrain not loaded: %s", conf.AllowDrain)
	}
	if conf.User != "user" {
		t.Errorf("User not loaded: %s", conf.User)
	}
	if !reflect.DeepEqual(conf.Envs, []string{"abc=123", "xyz=987"}) {
		t.Errorf("Evns not loaded: %s", conf.Envs)
	}
	if conf.Group != "group" {
		t.Errorf("Group not loaded: %s", conf.Group)
	}
}

func TestFullConfig(t *testing.T) {
	conf, err := ReadConfiguration("tests/full.toml")
	if err != nil {
		t.Error(err)
	}

	if conf.StopSequence[0].Signal.Signal != syscall.SIGKILL {
		t.Errorf("StopSequence signal not loaded %v", conf.StopSequence)
	}

	if conf.DrainSignal.Signal.Signal != syscall.SIGUSR2 {
		t.Errorf("DrainSignal signal not loaded %v", conf.DrainSignal)
	}
}
