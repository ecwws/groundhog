package main

import (
	"flag"
	"fmt"
	"github.com/BurntSushi/toml"
	"io/ioutil"
	"os"
	"os/signal"
	"syscall"
)

type GHConfig struct {
	Workers  int    `toml:"workers"`
	Endpoint string `toml:"endpoint"`
}

var Logger *MultiLogger
var GlobalConfig *GHConfig

func main() {
	var err error

	confFile := flag.String("conf", "/etc/groundhog/groundhog.toml",
		"Conf files, you know, conf files")
	logLevel := flag.String("loglevel", "info", "Log level")

	flag.Parse()

	Logger, err = NewMultiLogger(os.Stdout, *logLevel)

	if err != nil {
		fmt.Fprintln(os.Stderr, "Unable to initialize logger:", err)
		os.Exit(1)
	}

	confRaw, err := ioutil.ReadFile(*confFile)

	GlobalConfig = &GHConfig{Workers: 4, Endpoint: "test"}
	if err != nil {
		Logger.Info.Println("Unable to read config file, use default values",
			err)
	} else {
		Logger.Debug.Println("Config file read")
		err := toml.Unmarshal(confRaw, GlobalConfig)
		if err != nil {
			Logger.Warn.Println("Unable to decode config file:", err)
		}
	}

	Logger.Debug.Println("GlobalConfig:", *GlobalConfig)

	if GlobalConfig.Workers < 1 {
		Logger.Error.Fatal("Must have at least 1 worker, current:",
			GlobalConfig.Workers)
	}

	trap := make(chan os.Signal, 1)
	signal.Notify(trap, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	Logger.Debug.Println("Signals trapped")

	termination := make(chan bool)
	waitGroup := make(chan bool)

	for i := 0; i < GlobalConfig.Workers; i++ {
		Logger.Debug.Println("Spawn worker:", i)
		go worker(termination, waitGroup)
	}

	<-trap

	Logger.Info.Println("Received termination request, terminating")

	close(termination)

	Logger.Debug.Println("Waiting for worker to terminate")

	for i := 0; i < GlobalConfig.Workers; i++ {
		<-waitGroup
	}
}
