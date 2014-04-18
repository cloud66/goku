package main

import (
	"errors"
	"net"
	"net/rpc"
	"net/rpc/jsonrpc"

	"github.com/cloud66/goku/models"
	//	"github.com/golang/glog"
)

type Control struct {
	processSets []*ProcessSet
}

func (c *Control) Stop(ctrlProcessSet *models.CtrlProcessSet, reply *int) error {
	procSet, err := c.findProcessSet(ctrlProcessSet)
	if err != nil {
		return err
	}

	err = procSet.stop()
	if err != nil {
		return err
	}

	return nil
}

func (c *Control) List(_ *int, reply *[]models.CtrlProcessSet) error {
	// map real process sets to control process sets
	var ctrlProcesses []models.CtrlProcessSet
	for _, processSet := range c.processSets {
		ctrlProcesses = append(ctrlProcesses, processSet.toCtrlProcessSet())
	}

	*reply = ctrlProcesses
	return nil
}

func registerServer(processSets []*ProcessSet) error {
	control := new(Control)
	control.processSets = processSets

	server := rpc.NewServer()
	server.Register(control)
	server.HandleHTTP(rpc.DefaultRPCPath, rpc.DefaultDebugPath)
	l, err := net.Listen("tcp", "127.0.0.1:1234")
	if err != nil {
		return err
	}

	for {
		conn, err := l.Accept()
		if err != nil {
			return err
		}

		go server.ServeCodec(jsonrpc.NewServerCodec(conn))
	}

	return nil
}

func (c *Control) findProcessSet(ctrlProcessSet *models.CtrlProcessSet) (*ProcessSet, error) {
	for _, processSet := range c.processSets {
		if processSet.Name == ctrlProcessSet.Name {
			return processSet, nil
		}
	}

	return nil, errors.New("process not found")
}
