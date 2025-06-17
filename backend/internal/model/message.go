package model

// Message represents a WebSocket message.
type Message struct {
	Type int
	Body []byte
}
