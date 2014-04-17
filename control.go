package main

import (
	"net/rpc"
	"net"
	"net/http"

	"github.com/cloud66/goku/models"
)

type Control struct {
	processSets []*ProcessSet
}

func (c *Control) List(_ *int, reply []*models.CtrlProcessSet) error {
	// map real process sets to control process sets
	var ctrlProcesses []*models.CtrlProcessSet
	for _, processSet := range c.processSets {
		ctrlProcesses = append(ctrlProcesses, processSet.ToCtrlProcessSet())
	}

	reply = ctrlProcesses
	return nil
}

func registerServer(processSets []*ProcessSet) error {
	control := new(Control)
	control.processSets = processSets
	rpc.Register(control)
	rpc.HandleHTTP()
	l, err := net.Listen("tcp", "127.0.0.1:1234")
	if err != nil {
		return err
	}
	go http.Serve(l, nil)

	return nil
}
