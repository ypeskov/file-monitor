package main

import (
	"sync"
	"ypeskov/file-monitor/internal/config"
	"ypeskov/file-monitor/internal/monitor"
	"ypeskov/file-monitor/internal/server"

	log "github.com/sirupsen/logrus"
)

func main() {
	cfg := config.New()
	log.SetReportCaller(true)

	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
	})
	log.Info("Starting server on port ", cfg.Port)

	paths := []string{"./"}
	dirsToMonitor := map[string]struct{}{paths[0]: {}}

	stopChan := make(chan struct{})

	eventsChan := make(chan monitor.EventMessage, 100)
	handler := &monitor.WebSocketHandler{EventsChan: eventsChan}

	var wg sync.WaitGroup

	server := server.New(cfg, eventsChan)

	monitor.Init(dirsToMonitor, stopChan, &wg, handler)

	err := server.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}
