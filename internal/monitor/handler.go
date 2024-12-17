package monitor

import (
	"fmt"

	log "github.com/sirupsen/logrus"
)

type EventMessage struct {
	Event string `json:"event"`
	Path  string `json:"path"`
}

type EventHandler interface {
	Handle(event string, path string)
}

type LogHandler struct{}

func (l *LogHandler) Handle(event string, path string) {
	log.Infof("Event: %s, Path: %s", event, path)
}

type WebSocketHandler struct {
	EventsChan chan EventMessage
}

func (w *WebSocketHandler) Handle(event string, path string) {
	fmt.Println("Event: ", event, "Path***: ", path)
	w.EventsChan <- EventMessage{Event: event, Path: path}
}
