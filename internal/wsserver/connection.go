package wsserver

import (
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
)

func HandleConnection(conn *websocket.Conn) {
	log.Info("WebSocket connection established")
	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			log.Warnf("WebSocket error: %v", err)
			break
		}
		log.Infof("Received: %s", msg)

		if err := conn.WriteMessage(websocket.TextMessage, []byte("Message received: "+string(msg))); err != nil {
			log.Errorf("Error writing WebSocket message: %v", err)
			break
		}
	}
	log.Info("WebSocket connection closed")
}
