package main

import (
	"flag"
	"os"
	"time"

	"github.com/golang/glog"
)

const (
	Verbose = 5
	Detail  = 4
	Debug   = 3
)

var flagConfName string

func main() {
	args := os.Args[1:]

	flag.StringVar(&flagConfName, "c", "", "configuration file (toml format)")
	flag.Parse()

	if len(args) > 0 && args[0] == "help" {
		flag.PrintDefaults()
	}

	if flagConfName == "" {
		glog.Error("No configuration file specified. Use the -c option")
		return
	}

	if _, err := os.Stat(flagConfName); os.IsNotExist(err) {
		glog.Errorf("Configuration file not found: %s", flagConfName)
	}

	conf, err := ReadConfiguration(flagConfName)
	if err != nil {
		glog.Error(err)
	}
	glog.Infof("Starting Goku with configuration %s", flagConfName)

	var p = LoadFromConfig(conf)
	err = p.Start()
	if err != nil {
		glog.Error(err)
	}

	registerServer([]*ProcessSet{p})

	glog.Info("Started. Control is now listening to tcp:1234")

	// sleep
	for {
		time.Sleep(1 * time.Second)
	}
}
