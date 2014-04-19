package main

import (
)

var cmdReload = &Command{
	Run:      runReload,
	Usage:    "reload",
	NeedsProcess: true,
	Category: "process",
	Short:    "reloads configuration for a process",
	Long:     `This reloads the configuration file for a process set. It will try to do it
	on-the-fly but if the changes require a restart, it will kill the whole process set and start it
	again.

	Changes in Name, Command, Directory, Env, UseEnv, User and Group will cause a restart`,
}

func runReload(cmd *Command, args []string) {
	process := mustProcess()

	client.Reload(process)
}
