package monitor

import (
	"sync"

	"github.com/fsnotify/fsnotify"
	log "github.com/sirupsen/logrus"
)

func Init(dirs map[string]struct{}, stopChan <-chan struct{}, wg *sync.WaitGroup) {
	for dir, _ := range dirs {
		wg.Add(1)
		go func(directory string) {
			defer wg.Done()
			if err := MonitorDirectory(directory, stopChan); err != nil {
				log.Errorf("Error monitoring directory %s: %v", directory, err)
			}
		}(dir)
	}
}

func MonitorDirectory(path string, stopChan <-chan struct{}) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	defer watcher.Close()

	if err := watcher.Add(path); err != nil {
		return err
	}

	log.Infof("Started monitoring: %s", path)

	return handleEvents(watcher, path, stopChan)
}

func handleEvents(watcher *fsnotify.Watcher, path string, stopChan <-chan struct{}) error {
	for {
		select {
		case <-stopChan:
			log.Infof("Stopping monitoring for %s", path)
			return nil

		case event, ok := <-watcher.Events:
			if !ok {
				return nil
			}
			processEvent(event, path)

		case err, ok := <-watcher.Errors:
			if !ok {
				return nil
			}
			log.Errorf("Watcher error for %s: %v", path, err)
		}
	}
}

func processEvent(event fsnotify.Event, path string) {
	log.Infof("Event: %s on file %s", event.Op, event.Name)

	if event.Op&fsnotify.Remove == fsnotify.Remove {
		log.Warnf("Directory %s was removed", path)
	}
}
