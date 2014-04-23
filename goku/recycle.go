package main

import ()

var cmdRecycle = &Command{
	Run:          runRecycle,
	Usage:        "recycle",
	NeedsProcess: true,
	Category:     "process",
	Short:        "recycles a process",
	Long: `This works by sending the drain signal to the active process of a set
	it then starts a new process in the same set and marks it as active.
	The drained process will be killed after a predetermined time`,
}

func runRecycle(cmd *Command, args []string) {
	processes := mustProcess()

	for _, process := range *processes {
		err := client.Recycle(&process)
		if err != nil {
			printFatal(err.Error())
		}
	}
}
