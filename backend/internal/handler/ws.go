package handler

import (
	"log"
	"net/http"

	"intro-quiz/backend/internal/service"
	"intro-quiz/backend/pkg/ws"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

// WSHandler upgrades the HTTP request to a WebSocket connection.
// @Summary      WebSocket endpoint
// @Description  Upgrade the request and start echoing messages over WebSocket.
// @Tags         websocket
// @Success      101 {string} string "Switching Protocols"
// @Router       /ws [get]
func WSHandler(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("upgrade: %v", err)
		return
	}
	svc := service.NewEchoService()
	client := ws.NewClient(conn, svc)
	log.Printf("client connected: %s", conn.RemoteAddr())
	client.Listen()
	log.Printf("client disconnected: %s", conn.RemoteAddr())
}
