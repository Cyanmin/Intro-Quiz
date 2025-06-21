package model

// ClientMessage represents a message received from the client.
type ClientMessage struct {
	Type       string `json:"type"`
	User       string `json:"user,omitempty"`
	PlaylistID string `json:"playlistId,omitempty"`
	Answer     string `json:"answer,omitempty"`
}

// ServerMessage represents a message sent to clients.
type ServerMessage struct {
	Type       string          `json:"type"`
	User       string          `json:"user,omitempty"`
	Timestamp  int64           `json:"timestamp"`
	ReadyUsers map[string]bool `json:"readyUsers,omitempty"`
	VideoID    string          `json:"videoId,omitempty"`
	BuzzOrder  []string        `json:"buzzOrder,omitempty"`
	VideoTitle string          `json:"videoTitle,omitempty"`
	Correct    bool            `json:"correct,omitempty"`
}
