package main

import (
)

var cmdStart = &Command{
	Run:      runStart,
	Usage:    "start",
	NeedsProcess: true,
	Category: "process",
	Short:    "starts a process",
	Long:     `This will try to start the process`,
}

func runStart(cmd *Command, args []string) {
	process := mustProcess()

	client.Start(process)
}
