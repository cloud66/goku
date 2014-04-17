package main

import (
	"log"
	"net"
	"net/rpc/jsonrpc"
	"fmt"

	"github.com/cloud66/goku/models"
)

var serverAddress = "127.0.0.1"

func main() {
	// connect to the server
	conn, err := net.Dial("tcp", serverAddress + ":1234")
	if err != nil {
		log.Fatal("dialing:", err)
	}
	client := jsonrpc.NewClient(conn)

	var reply *[]models.CtrlProcessSet

	err = client.Call("Control.List", 1, &reply)
	if err != nil {
		log.Fatal("control error:", err)
	}

	fmt.Print(reply)
}
