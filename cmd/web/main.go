package main

import (
	"sync"
	"ypeskov/file-monitor/internal/config"
	"ypeskov/file-monitor/internal/directories"
	"ypeskov/file-monitor/internal/monitor"
	"ypeskov/file-monitor/internal/server"
	"ypeskov/file-monitor/internal/utils"

	log "github.com/sirupsen/logrus"
)

func main() {
	cfg := config.New()
	log.SetReportCaller(true)

	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
	})
	log.Info("Starting server on port ", cfg.Port)

	paths := utils.ParseArgs()

	dirsToMonitor := directories.PrepareDirsForMonitoring(paths)

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
