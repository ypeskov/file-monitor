package main

import (
	"os"
	"os/signal"
	"sync"
	"syscall"

	log "github.com/sirupsen/logrus"

	"ypeskov/file-monitor/internal/directories"
	"ypeskov/file-monitor/internal/monitor"
	"ypeskov/file-monitor/internal/utils"
)

func main() {
	log.SetLevel(log.DebugLevel)

	paths := utils.ParseArgs()

	dirsToMonitor := directories.PrepareDirsForMonitoring(paths)

	stopChan := make(chan struct{})
	var wg sync.WaitGroup

	handler := &monitor.LogHandler{}
	log.Info("Starting monitoring directories")
	monitor.Init(dirsToMonitor, stopChan, &wg, handler)

	waitForShutdown(stopChan, &wg)

	log.Info("The End")
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
