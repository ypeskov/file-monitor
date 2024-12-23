package server

import (
	"fmt"
	"net/http"
	"sync"

	"ypeskov/file-monitor/internal/config"
	"ypeskov/file-monitor/internal/monitor"
	"ypeskov/file-monitor/internal/render"
	"ypeskov/file-monitor/web"
	"ypeskov/file-monitor/web/templates"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	log "github.com/sirupsen/logrus"
)

var (
	clients   = make(map[*websocket.Conn]bool)
	clientsMu sync.Mutex
)

func New(cfg *config.Config, eventsChan <-chan monitor.EventMessage) *http.Server {
	e := echo.New()

	fileServer := http.FileServer(http.FS(web.Files))
	e.GET("/public/*", echo.WrapHandler(http.StripPrefix("/public/", fileServer)))

	e.GET("/", HomeHandler)

	upgrader := websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}
	e.GET("/ws", func(c echo.Context) error {
		conn, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
		if err != nil {
			log.Errorf("WebSocket upgrade failed: %v", err)
			return err
		}
		defer conn.Close()

		clientsMu.Lock()
		clients[conn] = true
		clientsMu.Unlock()
		log.Info("New WebSocket client connected")

		defer func() {
			clientsMu.Lock()
			delete(clients, conn)
			clientsMu.Unlock()
			log.Info("WebSocket client disconnected")
		}()

		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				break
			}
		}
		return nil
	})

	go RunWebSocketServer(eventsChan)

	return &http.Server{
		Addr:    fmt.Sprintf(":%s", cfg.Port),
		Handler: e,
	}
}

func HomeHandler(c echo.Context) error {
	log.Info("Home page accessed")
	component := templates.HomePage()
	return render.Render(c, http.StatusOK, component)
}

func RunWebSocketServer(eventsChan <-chan monitor.EventMessage) {
	log.Info("RunWebSocketServer started, waiting for events...")

	for event := range eventsChan {
		log.Infof("Sending event: %s - %s", event.Event, event.Path)

		clientsMu.Lock()
		for client := range clients {
			if err := client.WriteJSON(event); err != nil {
				log.Warnf("Error sending message to client: %v", err)
				client.Close()
				delete(clients, client)
			}
		}
		clientsMu.Unlock()
	}
}
