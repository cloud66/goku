package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"

	"github.com/cloud66/goku/models"
	"github.com/jcoene/honeybadger"
	"github.com/mgutz/ansi"
)

type Command struct {
	Run          func(cmd *Command, args []string)
	Flag         flag.FlagSet
	NeedsProcess bool

	Usage    string
	Category string
	Short    string
	Long     string
}

var (
	client          Client
	debugMode       bool   = false
	VERSION         string = "dev"
	BUILD_DATE      string = ""
	flagProcessName string
	flagProcess     *models.CtrlProcessSet
)

func (c *Command) printUsage() {
	c.printUsageTo(os.Stderr)
}

func (c *Command) printUsageTo(w io.Writer) {
	if c.Runnable() {
		fmt.Fprintf(w, "Usage: cx %s\n\n", c.FullUsage())
	}
	fmt.Fprintln(w, strings.Trim(c.Long, "\n"))
}

func (c *Command) FullUsage() string {
	if c.NeedsProcess {
		return c.Name() + " [-p <process>]" + strings.TrimPrefix(c.Usage, c.Name())
	}
	return c.Usage
}

func (c *Command) Name() string {
	name := c.Usage
	i := strings.Index(name, " ")
	if i >= 0 {
		name = name[:i]
	}
	return name
}

func (c *Command) Runnable() bool {
	return c.Run != nil
}

const extra = " (extra)"

func (c *Command) List() bool {
	return c.Short != "" && !strings.HasSuffix(c.Short, extra)
}

func (c *Command) ListAsExtra() bool {
	return c.Short != "" && strings.HasSuffix(c.Short, extra)
}

func (c *Command) ShortExtra() string {
	return c.Short[:len(c.Short)-len(extra)]
}

var commands = []*Command{
	cmdList,
	cmdStop,
	cmdStart,
	cmdRecycle,
	cmdReload,
	cmdLoad,

	/*	helpCommands,
		helpEnviron,
		helpMore,*/
}

var serverAddress = "127.0.0.1"

func main() {
	honeybadger.ApiKey = "2188ca35"

	// make sure command is specified, disallow global args
	args := os.Args[1:]
	if len(args) < 1 || strings.IndexRune(args[0], '-') == 0 {
		printUsageTo(os.Stderr)
		os.Exit(2)
	}

	/*	if args[0] == cmdUpdate.Name() {
			cmdUpdate.Run(cmdUpdate, args[1:])
			return
		} else if VERSION != "dev" {
			defer backgroundRun()
		}*/

	if !IsANSI(os.Stdout) {
		ansi.DisableColors(true)
	}

	initClients()

	for _, cmd := range commands {
		if cmd.Name() == args[0] && cmd.Run != nil {
			defer recoverPanic()

			cmd.Flag.Usage = func() {
				cmd.printUsage()
			}
			if cmd.NeedsProcess {
				cmd.Flag.StringVar(&flagProcessName, "p", "", "process name")
			}
			if err := cmd.Flag.Parse(args[1:]); err != nil {
				os.Exit(2)
			}
			if cmd.NeedsProcess {
				s, err := process()
				switch {
				case err == nil && s == nil:
					msg := "no process specified"
					if err != nil {
						msg = err.Error()
					}
					printError(msg)
					cmd.printUsage()
					os.Exit(2)
				case err != nil:
					printFatal(err.Error())
				}
			}
			cmd.Run(cmd, cmd.Flag.Args())
			return
		}
	}

	// invalid command
	fmt.Fprintf(os.Stderr, "Unknown command: %s\n", args[0])
	/*	if g := suggest(args[0]); len(g) > 0 {
		fmt.Fprintf(os.Stderr, "Possible alternatives: %v\n", strings.Join(g, " "))
	}*/
	fmt.Fprintf(os.Stderr, "Run 'goku help' for usage.\n")
	os.Exit(2)
}

func initClients() {
	client.initializeRpc("127.0.0.1")
}

func recoverPanic() {
	if VERSION != "dev" {
		if rec := recover(); rec != nil {
			report, err := honeybadger.NewReport(rec)
			if err != nil {
				printError("reporting crash failed: %s", err.Error())
				panic(rec)
			}
			report.AddContext("Version", VERSION)
			report.AddContext("Platform", runtime.GOOS)
			report.AddContext("Architecture", runtime.GOARCH)
			report.AddContext("DebugMode", debugMode)
			result := report.Send()
			if result != nil {
				printError("reporting crash failed: %s", result.Error())
				panic(rec)
			}
			printFatal("goku encountered and reported an internal client error")
		}
	}
}

func process() (*models.CtrlProcessSet, error) {
	if flagProcess != nil {
		return flagProcess, nil
	}

	if flagProcessName != "" {
		processes, err := client.List()
		if err != nil {
			return nil, err
		}
		var processNames []string
		for _, process := range *processes {
			processNames = append(processNames, process.Name)
		}
		idx, err := fuzzyFind(processNames, flagProcessName)
		if err != nil {
			return nil, err
		}

		flagProcess = &(*processes)[idx]
		fmt.Printf("Process: %s\n", flagProcess.Name)
		return flagProcess, err
	}

	return nil, errors.New("No process found")
}

func mustProcess() *models.CtrlProcessSet {
	process, err := process()
	if err != nil {
		printFatal(err.Error())
	}
	return process
}
