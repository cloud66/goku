package main

import (
	"os"
	"text/tabwriter"
)

var cmdLoad = &Command{
	Run:      runLoad,
	Usage:    "load -c <configuration file>",
	Category: "process",
	Short:    "loads a new configuration without starting it",
	Long:     `This loads a new process configuration into the daemon. The file should exist in the
	configuration directory of the daemon. To reload an existing configuration use the reload command.

	Use the filename with the .toml extension only. Not the full path`,
}

var flagConfigFile string

func init() {
	cmdLoad.Flag.StringVar(&flagConfigFile, "c", "", "configuration file name")
}

func runLoad(cmd *Command, args []string) {
	procSet, err := client.Load(flagConfigFile)
	if err != nil {
		must(err)
	}

	w := tabwriter.NewWriter(os.Stdout, 1, 2, 2, ' ', 0)
	defer w.Flush()

	listProcess(w, *procSet)
}
