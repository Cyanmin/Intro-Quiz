package ws

import (
	"log"

	"github.com/gorilla/websocket"
)

// OutgoingMessage holds a WebSocket message to be sent.
type OutgoingMessage struct {
	Type int
	Data []byte
}

// Client represents a single WebSocket connection.
type Client struct {
	Conn    *websocket.Conn
	Service MessageService
	Send    chan OutgoingMessage
}

// MessageService defines processing behavior for incoming messages.
type MessageService interface {
	ProcessMessage(messageType int, message []byte) (int, []byte)
}

// NewClient creates a Client bound to the given service.
func NewClient(conn *websocket.Conn, svc MessageService) *Client {
	return &Client{
		Conn:    conn,
		Service: svc,
		Send:    make(chan OutgoingMessage, 8),
	}
}

// Listen reads messages from the WebSocket and sends back the processed result.
func (c *Client) Listen() {
	defer c.Conn.Close()

	// send loop
	done := make(chan struct{})
	go func() {
		for msg := range c.Send {
			if err := c.Conn.WriteMessage(msg.Type, msg.Data); err != nil {
				log.Printf("write: %v", err)
				break
			}
		}
		close(done)
	}()

	for {
		mt, msg, err := c.Conn.ReadMessage()
		if err != nil {
			log.Printf("read: %v", err)
			break
		}
		log.Printf("recv: %s", msg)
		respType, respMsg := c.Service.ProcessMessage(mt, msg)
		if respMsg != nil {
			c.Send <- OutgoingMessage{Type: respType, Data: respMsg}
		}
	}

	close(c.Send)
	<-done
}
