package main

import (
	"log"
	"net"
	"net/rpc"
	"net/rpc/jsonrpc"
	"strconv"

	"github.com/cloud66/goku/models"
)

type Client struct {
	client *rpc.Client
}

func (c *Client) initializeRpc(serverAddress string) error {
	// connect to the server
	conn, err := net.Dial("tcp", serverAddress+":"+strconv.Itoa(config.ServerPort))
	if err != nil {
		log.Fatal("dialing:", err)
	}
	c.client = jsonrpc.NewClient(conn)

	return nil
}

func (c *Client) Version() string {
	var reply *models.CtrlVersion

	err := c.client.Call("Control.Version", 1, &reply)
	if err != nil {
		return err.Error()
	}

	if reply.BuildDate != "" {
		return reply.Version + ", " + reply.BuildDate
	} else {
		return reply.Version
	}
}

func (c *Client) List() (*[]models.CtrlProcessSet, error) {
	var reply *[]models.CtrlProcessSet

	err := c.client.Call("Control.List", 1, &reply)
	if err != nil {
		return nil, err
	}

	return reply, nil
}

func (c *Client) Stop(process *models.CtrlProcessSet) error {
	var reply *int
	err := c.client.Call("Control.Stop", process, &reply)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) Recycle(process *models.CtrlProcessSet) error {
	var reply *int
	err := c.client.Call("Control.Recycle", process, &reply)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) Reload(process *models.CtrlProcessSet) error {
	var reply *int
	err := c.client.Call("Control.Reload", process, &reply)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) Load(file string) (*models.CtrlProcessSet, error) {
	var reply *models.CtrlProcessSet
	err := c.client.Call("Control.Load", file, &reply)
	if err != nil {
		return nil, err
	}

	return reply, nil
}

func (c *Client) Start(process *models.CtrlProcessSet) error {
	var reply *int
	err := c.client.Call("Control.Start", process, &reply)
	if err != nil {
		return err
	}

	return nil
}
