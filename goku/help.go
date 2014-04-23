package main

import (
	"fmt"
	"html/template"
	"io"
	"log"
	"os"
	"sort"
	"text/tabwriter"
)

var helpEnviron = &Command{
	Usage:    "environ",
	Category: "goku",
	Short:    "environment variables used by goku",
	Long: `
Several environment variables affect goku's behavior.

GOKUDEBUG
	When this is set, goku prints the wire representation of each API
  request to stderr just before sending the request, and prints the
  response. This will most likely include your secret API key in
  the Authorization header field, so be careful with the output.
`,
}

var cmdVersion = &Command{
	Run:      runVersion,
	Usage:    "version",
	Category: "goku",
	Short:    "show goku version",
	Long:     `Version shows the goku client version string.`,
}

func runVersion(cmd *Command, args []string) {
	fmt.Println(VERSION)
	fmt.Printf("Server: %s\n", client.Version())
	if debugMode {
		fmt.Println("Running in debug mode")
		fmt.Printf("Build date: %s\n", BUILD_DATE)
	}
}

var cmdHelp = &Command{
	Usage:    "help [<topic>]",
	Category: "goku",
	Long:     `Help shows usage for a command or other topic.`,
}

var helpMore = &Command{
	Usage:    "more",
	Category: "goku",
	Short:    "additional commands, less frequently used",
	Long:     "(not displayed; see special case in runHelp)",
}

var helpCommands = &Command{
	Usage:    "commands",
	Category: "goku",
	Short:    "list all commands with usage",
	Long:     "(not displayed; see special case in runHelp)",
}

func init() {
	cmdHelp.Run = runHelp // break init loop
}

func runHelp(cmd *Command, args []string) {
	if len(args) == 0 {
		printUsageTo(os.Stdout)
		return // not os.Exit(2); success
	}
	if len(args) != 1 {
		printFatal("too many arguments")
	}
	switch args[0] {
	case helpMore.Name():
		printExtra()
		return
	case helpCommands.Name():
		printAllUsage()
		return
	}

	for _, cmd := range commands {
		if cmd.Name() == args[0] {
			cmd.printUsageTo(os.Stdout)
			return
		}
	}

	log.Printf("Unknown help topic: %q. Run 'goku help'.\n", args[0])
	os.Exit(2)
}

func maxStrLen(strs []string) (strlen int) {
	for i := range strs {
		if len(strs[i]) > strlen {
			strlen = len(strs[i])
		}
	}
	return
}

var usageTemplate = template.Must(template.New("usage").Parse(`
Usage: goku <command> [-p process|-t tag] [options] [arguments]


Commands:
{{range .Commands}}{{if .Runnable}}{{if .List}}
    {{.Name | printf (print "%-" $.MaxRunListName "s")}}  {{.Short}}{{end}}{{end}}{{end}}
{{range .Plugins}}
    {{.Name | printf (print "%-" $.MaxRunListName "s")}}  {{.Short}} (plugin){{end}}

Run 'goku help [command]' for details.


Additional help topics:
{{range .Commands}}{{if not .Runnable}}
    {{.Name | printf "%-8s"}}  {{.Short}}{{end}}{{end}}

{{if .Dev}}This dev build of goku cannot auto-update itself.
{{end}}`[1:]))

var extraTemplate = template.Must(template.New("usage").Parse(`
Additional commands:
{{range .Commands}}{{if .Runnable}}{{if .ListAsExtra}}
    {{.Name | printf (print "%-" $.MaxRunExtraName "s")}}  {{.ShortExtra}}{{end}}{{end}}{{end}}

Run 'goku help [command]' for details.

`[1:]))

func printUsageTo(w io.Writer) {
	var runListNames []string
	for i := range commands {
		if commands[i].Runnable() && commands[i].List() {
			runListNames = append(runListNames, commands[i].Name())
		}
	}

	usageTemplate.Execute(w, struct {
		Commands       []*Command
		Dev            bool
		MaxRunListName int
	}{
		commands,
		VERSION == "dev",
		maxStrLen(runListNames),
	})
}

func printExtra() {
	var runExtraNames []string
	for i := range commands {
		if commands[i].Runnable() && commands[i].ListAsExtra() {
			runExtraNames = append(runExtraNames, commands[i].Name())
		}
	}

	extraTemplate.Execute(os.Stdout, struct {
		Commands        []*Command
		MaxRunExtraName int
	}{
		commands,
		maxStrLen(runExtraNames),
	})
}

func printAllUsage() {
	w := tabwriter.NewWriter(os.Stdout, 1, 2, 2, ' ', 0)
	defer w.Flush()
	cl := commandList(commands)
	sort.Sort(cl)
	for i := range cl {
		if cl[i].Runnable() {
			listRec(w, "goku "+cl[i].FullUsage(), "# "+cl[i].Short)
		}
	}
}

type commandList []*Command

func (cl commandList) Len() int           { return len(cl) }
func (cl commandList) Swap(i, j int)      { cl[i], cl[j] = cl[j], cl[i] }
func (cl commandList) Less(i, j int) bool { return cl[i].Name() < cl[j].Name() }

type commandMap map[string]commandList
