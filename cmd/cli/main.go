package main

import (
	"fmt"
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
	if len(os.Args) < 2 {
		log.Fatal("Please provide at least one directory to monitor")
	}

	paths := os.Args[1:]

	validDirs := filterValidDirs(paths)

	if len(validDirs) == 0 {
		log.Fatal("No valid directories provided to monitor")
	}

	dirs, err := directories.GetAllDirs(validDirs)
	if err != nil {
		log.Fatalf("Failed to get directories: %v", err)
	}

	stopChan := make(chan struct{})

	var wg sync.WaitGroup

	monitor.Init(dirs, stopChan, &wg)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan
	close(stopChan)

	fmt.Println("Waiting for goroutines to finish...")
	wg.Wait()

	fmt.Println("Shutting down gracefully")
}

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
			log.Warnf("Skipping invalid or non-existent directory or not a dir: %s", dir)
		}
	}
	return validDirs
}
