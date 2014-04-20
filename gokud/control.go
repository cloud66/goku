package main

import (
	"errors"
	"net"
	"net/rpc"
	"net/rpc/jsonrpc"
	"path/filepath"

	"github.com/cloud66/goku/models"
	"github.com/golang/glog"
)

type Control struct {
	processSets []*ProcessSet
}

func (c *Control) Stop(ctrlProcessSet *models.CtrlProcessSet, reply *int) error {
	procSet, err := c.findProcessSet(ctrlProcessSet)
	if err != nil {
		return err
	}

	go func() {
		err = procSet.stop()
		if err != nil {
			glog.Error(err)
		}
	}()

	return nil
}

// loads a new configuration into the daemon. the file should be located in the
// config directory
func (c *Control) Load(configName *string, reply *models.CtrlProcessSet) error {
	config, err := ReadConfiguration(filepath.Join(flagConfName, *configName))
	if err != nil {
		glog.Error(err)
		return err
	}

	// do we have this already?
	for _, item := range c.processSets {
		if item.config.Name == config.Name {
			return errors.New("configuration with the same name already loaded")
		}
	}

	procSet := loadProcessSetFromConfig(config)

	errs := procSet.verifyPids()
	if len(errs) != 0 {
		return errs[0]
	}

	// add it to the list
	c.processSets = append(c.processSets, procSet)

	*reply = procSet.toCtrlProcessSet()
	return nil
}

func (c *Control) Reload(ctrlProcessSet *models.CtrlProcessSet, reply *int) error {
	procSet, err := c.findProcessSet(ctrlProcessSet)
	if err != nil {
		glog.Error(err)
		return err
	}

	err = procSet.reload()
	if err != nil {
		glog.Error(err)
		return err
	}

	return nil
}

func (c *Control) Recycle(ctrlProcessSet *models.CtrlProcessSet, reply *int) error {
	procSet, err := c.findProcessSet(ctrlProcessSet)
	if err != nil {
		glog.Error(err)
		return err
	}

	err = procSet.recycle()
	if err != nil {
		glog.Error(err)
		return err
	}

	return nil
}

func (c *Control) Start(ctrlProcessSet *models.CtrlProcessSet, reply *int) error {
	procSet, err := c.findProcessSet(ctrlProcessSet)
	if err != nil {
		glog.Error(err)
		return err
	}

	err = procSet.start()
	if err != nil {
		glog.Error(err)
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
		glog.Error(err)
		return err
	}

	for {
		conn, err := l.Accept()
		if err != nil {
			glog.Error(err)
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
