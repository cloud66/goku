package main

import (
	"log"
	"net/rpc"

	"github.com/cloud66/goku/models"
)

var serverAddress = "127.0.0.1"

func main() {
	// connect to the server
	client, err := rpc.DialHTTP("tcp", serverAddress + ":1234")
	if err != nil {
		log.Fatal("dialing:", err)
	}

	var reply []*models.CtrlProcessSet

	err = client.Call("Control.List", 1, &reply)
	if err != nil {
		log.Fatal("control error:", err)
	}
}
