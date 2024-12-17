package monitor

import (
	"fmt"
	"os"
	"sync"

	"github.com/fsnotify/fsnotify"
	log "github.com/sirupsen/logrus"
)

var removeMutex sync.Mutex // Mutex to protect against duplicate REMOVE handling

// Init starts monitoring for all provided directories and manages goroutines lifecycle.
func Init(dirs map[string]struct{}, stopChan <-chan struct{}, wg *sync.WaitGroup) {
	for dir := range dirs {
		StartMonitoringSingleDir(dir, dirs, stopChan, wg)
	}
}

// StartMonitoringSingleDir initializes a new goroutine to monitor a single directory.
func StartMonitoringSingleDir(directory string, dirs map[string]struct{}, stopChan <-chan struct{}, wg *sync.WaitGroup) {
	wg.Add(1)
	go func(dir string) {
		defer wg.Done()
		if err := MonitorDirectory(dir, dirs, stopChan, wg); err != nil {
			log.Errorf("Error monitoring directory %s: %v", dir, err)
		}
	}(directory)
}

// MonitorDirectory initializes a watcher for a specific directory and processes events.
func MonitorDirectory(path string, dirs map[string]struct{}, stopChan <-chan struct{}, wg *sync.WaitGroup) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	defer watcher.Close()

	if err := watcher.Add(path); err != nil {
		return err
	}

	log.Infof("Started monitoring: %s", path)

	return handleEvents(watcher, path, dirs, stopChan, wg)
}

// handleEvents processes filesystem events for a given watcher.
func handleEvents(watcher *fsnotify.Watcher, path string, dirs map[string]struct{}, stopChan <-chan struct{}, wg *sync.WaitGroup) error {
	for {
		select {
		case <-stopChan:
			log.Infof("Stopping monitoring for %s", path)
			return nil

		case event, ok := <-watcher.Events:
			if !ok {
				return nil
			}
			processEvent(event, watcher, dirs, stopChan, wg)

		case err, ok := <-watcher.Errors:
			if !ok {
				return nil
			}
			log.Errorf("Watcher error for %s: %v", path, err)
		}
	}
}

/*
processEvent handles individual filesystem events.

- Removes a directory from monitoring when it is deleted.
- Adds new directories to monitoring when they are created.
*/
func processEvent(event fsnotify.Event, watcher *fsnotify.Watcher, dirs map[string]struct{}, stopChan <-chan struct{}, wg *sync.WaitGroup) {
	log.Infof("Event: %s on file %s", event.Op, event.Name)

	// Handle directory removal
	if event.Op&fsnotify.Remove == fsnotify.Remove {
		removeMutex.Lock() // Protect against duplicate REMOVE handling
		defer removeMutex.Unlock()

		// Check if the path exists in the map and is being watched
		if _, exists := dirs[event.Name]; exists {
			log.Warnf("Directory %s was removed", event.Name)

			// Safely remove from watcher
			if err := watcher.Remove(event.Name); err != nil {
				log.Debugf("Cannot remove non-existent watch: %s", event.Name)
			} else {
				log.Infof("Stopped monitoring: %s", event.Name)
			}

			// Remove from the map and decrement the WaitGroup
			delete(dirs, event.Name)
			wg.Done()
		} else {
			log.Debugf("Duplicate or invalid REMOVE event ignored for %s", event.Name)
		}
	}

	// Handle new directory creation
	if event.Op&fsnotify.Create == fsnotify.Create {
		// Check if the created path is a directory
		if stat, err := os.Stat(event.Name); err == nil && stat.IsDir() {
			log.Infof("New directory created: %s", event.Name)

			// Avoid adding duplicates
			if _, exists := dirs[event.Name]; !exists {
				dirs[event.Name] = struct{}{}
				StartMonitoringSingleDir(event.Name, dirs, stopChan, wg)
				log.Infof("Started monitoring new directory: %s", event.Name)
			}
		}
	}

	// logDirs(dirs)
}

// logDirs outputs the current list of monitored directories.
func logDirs(dirs map[string]struct{}) {
	fmt.Println("Directories:")
	fmt.Println("------------")
	for dir := range dirs {
		fmt.Println(dir)
	}
	fmt.Println("------------")
}
