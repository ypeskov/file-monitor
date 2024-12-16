package monitor

import (
	"sync"

	"github.com/fsnotify/fsnotify"
	log "github.com/sirupsen/logrus"
)

func Init(dirs []string, stopChan <-chan struct{}, wg *sync.WaitGroup) {
	for _, dir := range dirs {
		wg.Add(1)
		go MonitorDirectory(dir, stopChan, wg)
	}
}

func MonitorDirectory(path string, stopChan <-chan struct{}, wg *sync.WaitGroup) {
	defer wg.Done()

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Errorf("Failed to initialize watcher: %v", err)
		return
	}
	defer watcher.Close()

	err = watcher.Add(path)
	if err != nil {
		log.Errorf("Failed to watch directory %s: %v", path, err)
		return
	}

	log.Infof("Started monitoring: %s", path)

	for {
		select {
		case <-stopChan:
			log.Infof("Stopping monitoring for %s", path)
			return

		case event := <-watcher.Events:
			log.Infof("Event: %s on file %s", event.Op, event.Name)

			// Дополнительная обработка событий
			if event.Op&fsnotify.Remove == fsnotify.Remove {
				log.Warnf("Directory %s was removed", path)
				return
			}

		case err := <-watcher.Errors:
			log.Errorf("Error watching %s: %v", path, err)
			return
		}
	}
}
