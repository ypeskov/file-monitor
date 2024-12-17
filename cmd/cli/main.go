package main

import (
	"os"
	"os/signal"
	"sync"
	"syscall"

	log "github.com/sirupsen/logrus"

	"ypeskov/file-monitor/internal/directories"
	"ypeskov/file-monitor/internal/monitor"
)

func main() {
	log.SetLevel(log.DebugLevel)

	paths := parseArgs()

	dirsToMonitor := directories.PrepareDirsForMonitoring(paths)
	if len(dirsToMonitor) == 0 {
		log.Fatal("No valid directories provided to monitor")
	}

	stopChan := make(chan struct{})
	var wg sync.WaitGroup

	handler := &monitor.LogHandler{}
	log.Info("Starting monitoring directories")
	monitor.Init(dirsToMonitor, stopChan, &wg, handler)

	waitForShutdown(stopChan, &wg)

	log.Info("The End")
}

/*
	parseArgs returns the list of directories to monitor.

If no directories are provided, it logs an error and exits.
*/
func parseArgs() []string {
	if len(os.Args) < 2 {
		log.Fatal("Please provide at least one directory to monitor")
	}
	return os.Args[1:]
}

/*
	waitForShutdown blocks until a system signal is received.

It then closes the stopChan and waits for all goroutines to stop.
*/
func waitForShutdown(stopChan chan struct{}, wg *sync.WaitGroup) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan
	close(stopChan)

	log.Info("Waiting for all goroutines to stop")
	wg.Wait()
}
