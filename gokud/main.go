package main

import (
	"flag"
	"os"
	"sync"
	"time"
	"path/filepath"

	"github.com/golang/glog"
)

const (
	Verbose = 5
	Detail  = 4
	Debug   = 3
)

var flagConfName string
var flagAutoStart bool
var loadWait sync.WaitGroup
var processes []*ProcessSet

func main() {
	args := os.Args[1:]

	flag.StringVar(&flagConfName, "d", "", "configuration file directory (toml format)")
	flag.BoolVar(&flagAutoStart, "s", false, "start the loaded configurations automatically")
	flag.Parse()

	if len(args) > 0 && args[0] == "help" {
		flag.PrintDefaults()
	}

	if flagConfName == "" {
		glog.Error("No configuration directory specified. Use the -d option")
		return
	}

	if _, err := os.Stat(flagConfName); os.IsNotExist(err) {
		glog.Errorf("Configuration directory not found: %s", flagConfName)
	}

	files, err := listConfigFiles(flagConfName)
	if err != nil {
		glog.Error(err)
	}

	glog.Infof("Loading configurations from %s", flagConfName)
	for _, file := range files {
		loadWait.Add(1)
		go loadConfiguration(file)
	}
	glog.Info("Waiting for all configurations to load")
	loadWait.Wait()

	registerServer(processes)

	glog.Info("Started. Control is now listening to tcp:1234")

	// sleep
	for {
		time.Sleep(1 * time.Second)
	}
}

func loadConfiguration(configFile string) {
	glog.Infof("Loading %s", configFile)

	conf, err := ReadConfiguration(configFile)
	if err != nil {
		glog.Error(err)
	}

	p := loadProcessSetFromConfig(conf)
	processes = append(processes, p)
	loadWait.Done()

	if flagAutoStart {
		err = p.start()
		if err != nil {
			glog.Error(err)
		}
	}
}

func listConfigFiles(directory string) ([]string, error) {
	files, err := filepath.Glob(filepath.Join(directory, "*.toml"))
	glog.Info(files)
	if err != nil {
		return nil, err
	}

	return files, nil
}
