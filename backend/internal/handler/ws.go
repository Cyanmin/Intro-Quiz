package handler

import (
	"log"
	"net/http"
	"os"

	"intro-quiz/backend/internal/service"
	"intro-quiz/backend/pkg/ws"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

var roomManager = service.NewRoomManager()
var youtubeService = service.NewYouTubeService(os.Getenv("YOUTUBE_API_KEY"))
var playlistID = os.Getenv("YOUTUBE_PLAYLIST_ID")

// WSHandler upgrades the HTTP request to a WebSocket connection.
// @Summary      WebSocket endpoint
// @Description  Upgrade the request and start echoing messages over WebSocket.
// @Tags         websocket
// @Success      101 {string} string "Switching Protocols"
// @Router       /ws [get]
func WSHandler(c *gin.Context) {
	roomID := c.Query("roomId")
	if roomID == "" {
		roomID = "default"
	}
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("upgrade: %v", err)
		return
	}
	client := ws.NewClient(conn, nil)
	svc := service.NewRoomService(roomManager, roomID, client, youtubeService, playlistID)
	client.Service = svc
	roomManager.Join(roomID, client)
	defer roomManager.Leave(roomID, client)
	log.Printf("client connected: %s room:%s", conn.RemoteAddr(), roomID)
	client.Listen()
	log.Printf("client disconnected: %s room:%s", conn.RemoteAddr(), roomID)
}
