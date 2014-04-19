package main

import (
	"io"
	"os"
	"sort"
	"text/tabwriter"

	"github.com/cloud66/goku/models"
)

var cmdList = &Command{
	Run:      runList,
	Usage:    "list [-v]",
	Category: "process",
	Short:    "lists all the processes under goku",
	Long:     `This returns a list of all processes managed by goku with their status
	-v	Verbose. Lists all draining processes as well`,
}

var flagVerbose bool

func init() {
	cmdList.Flag.BoolVar(&flagVerbose, "v", false, "verbose reporting")
}

func runList(cmd *Command, args []string) {
	w := tabwriter.NewWriter(os.Stdout, 1, 2, 2, ' ', 0)
	defer w.Flush()

	result, err := client.List()
	must(err)

	if err != nil {
		printFatal(err.Error())
	} else {
		printProcessList(w, result)
	}
}

func printProcessList(w io.Writer, servers *[]models.CtrlProcessSet) {
	sort.Sort(processesByName(*servers))
	for _, a := range *servers {
		if a.Name != "" {
			listProcess(w, a)
		}
	}
}

func listProcess(w io.Writer, a models.CtrlProcessSet) {
	pid := 0
	if a.Active != nil {
		pid = a.Active.Pid
	}
	listRec(w,
		a.Name,
		pid,
		a.Tags,
		a.Status(),
	)

	if flagVerbose {
		if a.Active != nil {
			listRec(w,
				a.Active.Uid,
				a.Active.Pid,
				prettyTime{a.Active.LastActionAt},
				a.Active.Status.Message)
		}
		for _, p := range a.Draining {
			lastActivity := prettyTime{p.LastActionAt}
			listRec(w,
				p.Uid,
				p.Pid,
				lastActivity,
				p.Status.Message)
		}
	}
}

type processesByName []models.CtrlProcessSet

func (a processesByName) Len() int           { return len(a) }
func (a processesByName) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a processesByName) Less(i, j int) bool { return a[i].Name < a[j].Name }
