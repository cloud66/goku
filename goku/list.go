package main

import (
	"io"
	"os"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/cloud66/goku/models"
)

var cmdList = &Command{
	Run:      runList,
	Usage:    "list",
	Category: "process",
	Short:    "lists all the processes under goku",
	Long:     `This returns a list of all processes managed by goku with their status`,
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
	listRec(w,
		strings.ToLower(a.Name),
	)
}

type processesByName []models.CtrlProcessSet

func (a processesByName) Len() int           { return len(a) }
func (a processesByName) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a processesByName) Less(i, j int) bool { return a[i].Name < a[j].Name }
