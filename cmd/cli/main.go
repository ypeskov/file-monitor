package main

import (
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"

	log "github.com/sirupsen/logrus"

	"ypeskov/file-monitor/internal/directories"
	"ypeskov/file-monitor/internal/monitor"
)

func main() {
	paths := parseArgs()

	dirsToMonitor := prepareDirsForMonitoring(paths)
	if len(dirsToMonitor) == 0 {
		log.Fatal("No valid directories provided to monitor")
	}

	stopChan := make(chan struct{})
	var wg sync.WaitGroup

	log.Info("Starting monitoring directories")
	monitor.Init(dirsToMonitor, stopChan, &wg)

	waitForShutdown(stopChan, &wg)

	log.Info("Shutting down gracefully")
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
	prepareDirsForMonitoring prepares the list of directories for monitoring.

It filters out invalid directories and returns a list of all recursive directories.
*/
func prepareDirsForMonitoring(paths []string) []string {
	validDirs := filterValidDirs(paths)

	allRecursiveDirs, err := directories.GetAllDirs(validDirs)
	if err != nil {
		log.Fatalf("Failed to get directories: %v", err)
	}
	return allRecursiveDirs
}

/*
	waitForShutdown blocks until a system signal is received.

It then closes the stopChan and waits for all goroutines to stop.
*/
func waitForShutdown(stopChan chan struct{}, wg *sync.WaitGroup) {
	// Канал для получения системных сигналов
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan // Блокируем выполнение до получения сигнала
	close(stopChan)

	log.Info("Waiting for all goroutines to stop")
	wg.Wait()
}

/*
	filterValidDirs filters out invalid directories from the list of paths.

It returns a list of valid directories.
*/
func filterValidDirs(paths []string) []string {
	validDirs := []string{}
	for _, dir := range paths {
		absPath, err := filepath.Abs(dir)
		if err != nil {
			log.Errorf("Invalid path %s: %v", dir, err)
			continue
		}
		if stat, err := os.Stat(absPath); err == nil && stat.IsDir() {
			validDirs = append(validDirs, absPath)
		} else {
			log.Warnf("Skipping invalid or non-existent directory: %s", dir)
		}
	}
	return validDirs
}
