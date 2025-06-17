package service

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/gorilla/websocket"

	"intro-quiz/backend/internal/model"
	"intro-quiz/backend/pkg/ws"
)

// RoomManager manages WebSocket connections grouped by room ID.
type RoomState struct {
	Fastest string
	Active  bool
	Ready   map[string]bool
	Users   map[*ws.Client]string
	VideoID string
}

// RoomManager manages WebSocket connections grouped by room ID and quiz state.
type RoomManager struct {
	rooms  map[string]map[*ws.Client]bool
	states map[string]*RoomState
	mu     sync.RWMutex
}

// copyReady returns a copy of ready state map.
func copyReady(src map[string]bool) map[string]bool {
	dst := make(map[string]bool)
	for k, v := range src {
		dst[k] = v
	}
	return dst
}

// NewRoomManager creates a new RoomManager.
func NewRoomManager() *RoomManager {
	return &RoomManager{
		rooms:  make(map[string]map[*ws.Client]bool),
		states: make(map[string]*RoomState),
	}
}

// Join adds a connection to a room.
func (m *RoomManager) Join(roomID string, client *ws.Client) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.rooms[roomID]; !ok {
		m.rooms[roomID] = make(map[*ws.Client]bool)
	}
	m.rooms[roomID][client] = true
	if _, ok := m.states[roomID]; !ok {
		m.states[roomID] = &RoomState{Ready: make(map[string]bool), Users: make(map[*ws.Client]string)}
	}
}

// RegisterUser stores the user's name for a connection and returns current ready states.
func (m *RoomManager) RegisterUser(roomID string, client *ws.Client, name string) map[string]bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	st := m.states[roomID]
	if st == nil {
		st = &RoomState{Ready: make(map[string]bool), Users: make(map[*ws.Client]string)}
		m.states[roomID] = st
	}
	st.Users[client] = name
	if _, ok := st.Ready[name]; !ok {
		st.Ready[name] = false
	}
	return copyReady(st.Ready)
}

// SetReady marks a user as ready and returns if all are ready and current state map.
func (m *RoomManager) SetReady(roomID, name string) (bool, map[string]bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	st := m.states[roomID]
	if st == nil {
		return false, nil
	}
	st.Ready[name] = true
	all := true
	for _, v := range st.Ready {
		if !v {
			all = false
			break
		}
	}
	return all, copyReady(st.Ready)
}

// Leave removes a connection from a room.
func (m *RoomManager) Leave(roomID string, client *ws.Client) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if clients, ok := m.rooms[roomID]; ok {
		delete(clients, client)
		if st, ok := m.states[roomID]; ok {
			if name, exists := st.Users[client]; exists {
				delete(st.Users, client)
				delete(st.Ready, name)
			}
		}
		if len(clients) == 0 {
			delete(m.rooms, roomID)
			delete(m.states, roomID)
		}
	}
}

// Broadcast sends a message to all clients in the room except the sender.
func (m *RoomManager) Broadcast(roomID string, sender *ws.Client, mt int, msg []byte) {
	m.mu.RLock()
	clients := m.rooms[roomID]
	m.mu.RUnlock()

	for c := range clients {
		if c == sender {
			continue
		}
		c.Send <- ws.OutgoingMessage{Type: mt, Data: msg}
	}
}

// StartQuestion marks the room as active and resets fastest user.
func (m *RoomManager) StartQuestion(roomID, videoID string) {
	m.mu.Lock()
	st, ok := m.states[roomID]
	if !ok {
		st = &RoomState{}
		m.states[roomID] = st
	}
	st.Active = true
	st.Fastest = ""
	st.VideoID = videoID
	m.mu.Unlock()

	go func() {
		time.Sleep(10 * time.Second)
		m.mu.Lock()
		st := m.states[roomID]
		if st != nil && st.Active && st.Fastest == "" {
			st.Active = false
			m.mu.Unlock()
			resp, _ := json.Marshal(&model.ServerMessage{Type: "timeout", Timestamp: time.Now().UnixMilli()})
			m.Broadcast(roomID, nil, websocket.TextMessage, resp)
			return
		}
		m.mu.Unlock()
	}()
}

// SetFastest records the fastest user if not already set.
func (m *RoomManager) SetFastest(roomID, user string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	st := m.states[roomID]
	if st == nil || !st.Active || st.Fastest != "" {
		return false
	}
	st.Fastest = user
	st.Active = false
	return true
}

// IsActive returns whether a question is active.
func (m *RoomManager) IsActive(roomID string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	st := m.states[roomID]
	if st == nil {
		return false
	}
	return st.Active
}

// RoomService uses RoomManager to broadcast messages within a room.
type RoomService struct {
	manager    *RoomManager
	roomID     string
	client     *ws.Client
	yt         *YouTubeService
	playlistID string
}

// NewRoomService creates a RoomService for a specific connection and room.
func NewRoomService(m *RoomManager, roomID string, client *ws.Client, yt *YouTubeService, playlistID string) *RoomService {
	return &RoomService{manager: m, roomID: roomID, client: client, yt: yt, playlistID: playlistID}
}

// ProcessMessage broadcasts the received message to the room.
func (r *RoomService) ProcessMessage(mt int, msg []byte) (int, []byte) {
	var req model.ClientMessage
	if err := json.Unmarshal(msg, &req); err != nil {
		return 0, nil
	}

	switch req.Type {
	case "join":
		states := r.manager.RegisterUser(r.roomID, r.client, req.User)
		resp, _ := json.Marshal(&model.ServerMessage{Type: "ready_state", ReadyUsers: states, Timestamp: time.Now().UnixMilli()})
		r.client.Send <- ws.OutgoingMessage{Type: websocket.TextMessage, Data: resp}
		r.manager.Broadcast(r.roomID, r.client, websocket.TextMessage, resp)
	case "ready":
		all, states := r.manager.SetReady(r.roomID, req.User)
		resp, _ := json.Marshal(&model.ServerMessage{Type: "ready_state", ReadyUsers: states, Timestamp: time.Now().UnixMilli()})
		r.client.Send <- ws.OutgoingMessage{Type: websocket.TextMessage, Data: resp}
		r.manager.Broadcast(r.roomID, r.client, websocket.TextMessage, resp)
		if all {
			videoID, err := r.yt.GetRandomVideoID(r.playlistID)
			if err != nil {
				return 0, nil
			}
			r.manager.StartQuestion(r.roomID, videoID)
			startMsg, _ := json.Marshal(&model.ServerMessage{Type: "start", VideoID: videoID, Timestamp: time.Now().UnixMilli()})
			r.manager.Broadcast(r.roomID, nil, websocket.TextMessage, startMsg)
		}
	case "start":
		videoID, err := r.yt.GetRandomVideoID(r.playlistID)
		if err != nil {
			return 0, nil
		}
		r.manager.StartQuestion(r.roomID, videoID)
		resp, _ := json.Marshal(&model.ServerMessage{Type: "start", VideoID: videoID, Timestamp: time.Now().UnixMilli()})
		r.manager.Broadcast(r.roomID, nil, websocket.TextMessage, resp)
	case "buzz":
		// broadcast that someone pressed the answer button
		note, _ := json.Marshal(&model.ServerMessage{Type: "answer", User: req.User, Timestamp: time.Now().UnixMilli()})
		r.manager.Broadcast(r.roomID, r.client, websocket.TextMessage, note)

		if r.manager.SetFastest(r.roomID, req.User) {
			resp, _ := json.Marshal(&model.ServerMessage{Type: "buzz_result", User: req.User, Timestamp: time.Now().UnixMilli()})
			r.manager.Broadcast(r.roomID, nil, websocket.TextMessage, resp)
		}
	}

	return 0, nil
}
