package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"universum/config"
	"universum/engine"
	"universum/server"
	"universum/utils"
)

var configfile string

func main() {
	configureCommandLineParams()
	configerr := config.Load(configfile)

	if configerr != nil {
		log.Fatalf("Cannot proceed without config: %v", configerr)
	}

	engine.InitInfoStatistics()
	engine.DatabaseInfoStats.Server.ConfigFile = configfile
	engine.DatabaseInfoStats.Server.StartedAt = utils.GetCurrentReadableTime()

	var sigs chan os.Signal = make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGTERM, syscall.SIGINT, syscall.SIGHUP)

	var wg sync.WaitGroup
	wg.Add(1)

	go server.StartTCPServer(&wg)
	go server.WaitForSignal(&wg, sigs)

	wg.Wait()
}

func configureCommandLineParams() {
	flag.StringVar(&configfile, "config", config.GetDefaultConfigPath(), "db server config file name")

	flag.Parse()
}
