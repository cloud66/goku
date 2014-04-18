package main

import (
)

var cmdStop = &Command{
	Run:      runStop,
	Usage:    "stop",
	NeedsProcess: true,
	Category: "process",
	Short:    "stops a process",
	Long:     `This will try to stop the active process in a process set by
	sending it the stop sequence.
	If that fails to stop the process, it will try to force kill it`,
}

func runStop(cmd *Command, args []string) {
	process := mustProcess()

	client.Stop(process)
}
