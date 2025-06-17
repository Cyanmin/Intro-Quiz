package ws

import (
	"log"

	"github.com/gorilla/websocket"
)

// Client represents a single WebSocket connection.
type Client struct {
	Conn    *websocket.Conn
	Service MessageService
}

// MessageService defines processing behavior for incoming messages.
type MessageService interface {
	ProcessMessage(messageType int, message []byte) (int, []byte)
}

// NewClient creates a Client bound to the given service.
func NewClient(conn *websocket.Conn, svc MessageService) *Client {
	return &Client{Conn: conn, Service: svc}
}

// Listen reads messages from the WebSocket and sends back the processed result.
func (c *Client) Listen() {
	defer c.Conn.Close()
	for {
		mt, msg, err := c.Conn.ReadMessage()
		if err != nil {
			log.Printf("read: %v", err)
			break
		}
		log.Printf("recv: %s", msg)
		respType, respMsg := c.Service.ProcessMessage(mt, msg)
		if respMsg != nil {
			if err := c.Conn.WriteMessage(respType, respMsg); err != nil {
				log.Printf("write: %v", err)
				break
			}
		}
	}
}
