package main

import (
	"flag"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/golang/glog"
)

const (
	Verbose = 5
	Detail  = 4
	Debug   = 3
)

var (
	flagConfName    string
	flagAutoStart   bool
	flagAutoRecover bool
	loadWait        sync.WaitGroup
	processes       []*ProcessSet
	statusChange    chan *Process
	VERSION         string = "dev"
	BUILD_DATE      string = ""
	startLock			 sync.Mutex
)

func main() {
	args := os.Args[1:]

	flag.StringVar(&flagConfName, "d", "", "configuration file directory (toml format)")
	flag.BoolVar(&flagAutoStart, "autostart", false, "start the loaded configurations automatically")
	flag.BoolVar(&flagAutoRecover, "autorecover", false, "recover leftover processes")
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

	statusChange = make(chan *Process)

	glog.Infof("Loading configurations from %s", flagConfName)
	if len(files) == 0 {
		glog.Error("No configuration files found")
	}

	for _, file := range files {
		loadWait.Add(1)
		go loadConfiguration(file)
	}
	glog.Info("Waiting for all configurations to load")
	loadWait.Wait()

	go startNotifier(statusChange)
	registerServer(processes)

	glog.Info("Started. Control is now listening to tcp:9800")

	// sleep
	for {
		time.Sleep(1 * time.Second)
	}
}

func loadConfiguration(configFile string) {
	defer loadWait.Done()

	glog.Infof("Loading %s", configFile)

	conf, err := ReadConfiguration(configFile)
	if err != nil {
		glog.Error(err)
	}

	p := loadProcessSetFromConfig(conf)

	errs := p.verifyPids()
	if len(errs) != 0 {
		glog.Errorf("Process %s cannot be loaded", p.Name)
		for _, err := range errs {
			glog.Error(err.Error())
		}

		return
	}

	processes = append(processes, p)

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
