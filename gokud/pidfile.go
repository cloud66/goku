package main

import (
	"io/ioutil"
	"os"
	"strconv"
)

type Pidfile string

func (f *Pidfile) read() int {
	data, err := ioutil.ReadFile(string(*f))
	if err != nil {
		return 0
	}
	pid, err := strconv.ParseInt(string(data), 0, 32)
	if err != nil {
		return 0
	}
	return int(pid)
}

func (f *Pidfile) write(data int) error {
	err := ioutil.WriteFile(string(*f), []byte(strconv.Itoa(data)), 0660)
	if err != nil {
		return err
	}
	return nil
}

func (f *Pidfile) delete() bool {
	_, err := os.Stat(string(*f))
	if err != nil {
		return true
	}
	err = os.Remove(string(*f))
	if err == nil {
		return true
	}
	return false
}
