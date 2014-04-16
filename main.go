package main

import (
	"flag"
	"os"
//	"fmt"

	"github.com/golang/glog"
)

var flagConfName string

func main() {
	args := os.Args[1:]

	flag.StringVar(&flagConfName, "c", "", "configuration file (toml format)")
	flag.Parse()

	if len(args) > 0 && args[0] == "help" {
		flag.PrintDefaults()
	}

	conf, err := ReadConfiguration(flagConfName)
	if err != nil {
		glog.Error(err)
	}
	glog.Infof("Starting Goku with configuration %s", flagConfName)

	var p = LoadFromConfig(conf)
	glog.Infof("%v", p)
}
