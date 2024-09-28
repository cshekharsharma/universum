// Package main is the entry point for the Universum NoSQL DB server.
// It initializes configuration, starts the server, and handles OS signals gracefully.
// The server supports customizable parameters, such as the configuration file, and manages
// internal statistics tracking and termination signals.
package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"universum/config"
	"universum/engine"
	"universum/server"
	"universum/utils"
)

var (
	AppVersion string
	BuildTime  string
	GitHash    string
	AppEnv     string

	configfile string
)

// main is the entry point of the Universum DB server application. It performs the following tasks:
//   - Loads the application configuration from a file.
//   - Initializes statistics and server metadata.
//   - Starts the TCP server for handling incoming client connections.
//   - Listens for OS signals (e.g., SIGTERM, SIGINT) and gracefully shuts down the server.
func main() {
	// Set the application metadata from build variables.
	config.AppVersion = AppVersion
	config.GitHash = GitHash
	config.AppEnv = AppEnv
	config.BuildTime, _ = strconv.ParseInt(BuildTime, 10, 64)

	// Configure command-line parameters and load the configuration file.
	configureCommandLineParams()
	configerr := config.Init(configfile)
	if configerr != nil {
		log.Fatalf("Config Error:: %v", configerr)
	}

	// Initialize server statistics.
	engine.InitInfoStatistics()
	engine.DatabaseInfoStats.Server.ConfigFile = configfile
	engine.DatabaseInfoStats.Server.StartedAt = utils.GetCurrentReadableTime()

	// Prepare to handle OS signals for graceful shutdown.
	var sigs chan os.Signal = make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGTERM, syscall.SIGINT, syscall.SIGHUP)

	// Use a wait group to manage the shutdown process.
	var wg sync.WaitGroup
	wg.Add(1)

	// Start the TCP server and wait for OS signals in separate goroutines.
	go server.StartTCPServer(&wg)
	go server.WaitForSignal(&wg, sigs)

	// Wait for all goroutines to finish before exiting.
	wg.Wait()
}

// configureCommandLineParams sets up the command-line flags for the application.
// It defines a flag for specifying the configuration file path.
//
// The configuration file path can be provided with the `-config` flag, and if not provided,
// it defaults to the path returned by config.GetDefaultConfigPath().
func configureCommandLineParams() {
	flag.StringVar(&configfile, "config", config.DefaultConfigFilePath, "db server config file name")
	flag.Parse()
}
