package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func wsHandler(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("upgrade: %v", err)
		return
	}
	log.Printf("client connected: %s", conn.RemoteAddr())
	defer func() {
		log.Printf("client disconnected: %s", conn.RemoteAddr())
		conn.Close()
	}()
	for {
		mt, message, err := conn.ReadMessage()
		if err != nil {
			log.Printf("read: %v", err)
			break
		}
		log.Printf("recv: %s", message)
		if err = conn.WriteMessage(mt, message); err != nil {
			log.Printf("write: %v", err)
			break
		}
	}
}

func main() {
	router := gin.Default()
	router.GET("/ws", wsHandler)
	log.Println("Listening on :8080")
	log.Fatal(router.Run(":8080"))
}
