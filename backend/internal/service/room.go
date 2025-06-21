package service

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"

	"intro-quiz/backend/internal/config"
	"intro-quiz/backend/internal/model"
)

// RoomManager manages WebSocket connections grouped by room ID.
type RoomState struct {
	Fastest    string
	Active     bool
	Ready      map[string]bool
	Users      map[*websocket.Conn]string
	BuzzOrder  []string
	VideoTitle string
	PlaylistID string
}

// RoomManager manages WebSocket connections grouped by room ID and quiz state.
type RoomManager struct {
	rooms  map[string]map[*websocket.Conn]bool
	states map[string]*RoomState
	mu     sync.RWMutex
}

// ResetReady sets all ready states to false and returns the updated states.
func (m *RoomManager) ResetReady(roomID string) map[string]bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	st := m.states[roomID]
	if st == nil {
		return nil
	}
	for u := range st.Ready {
		st.Ready[u] = false
	}
	return copyReady(st.Ready)
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
		rooms:  make(map[string]map[*websocket.Conn]bool),
		states: make(map[string]*RoomState),
	}
}

// Join adds a connection to a room.
func (m *RoomManager) Join(roomID string, conn *websocket.Conn) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.rooms[roomID]; !ok {
		m.rooms[roomID] = make(map[*websocket.Conn]bool)
	}
	m.rooms[roomID][conn] = true
	if _, ok := m.states[roomID]; !ok {
		m.states[roomID] = &RoomState{Ready: make(map[string]bool), Users: make(map[*websocket.Conn]string)}
	}
}

// RegisterUser stores the user's name for a connection and returns current ready states.
func (m *RoomManager) RegisterUser(roomID string, conn *websocket.Conn, name string) map[string]bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	st := m.states[roomID]
	if st == nil {
		st = &RoomState{Ready: make(map[string]bool), Users: make(map[*websocket.Conn]string)}
		m.states[roomID] = st
	}
	st.Users[conn] = name
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
func (m *RoomManager) Leave(roomID string, conn *websocket.Conn) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if clients, ok := m.rooms[roomID]; ok {
		delete(clients, conn)
		if st, ok := m.states[roomID]; ok {
			if name, exists := st.Users[conn]; exists {
				delete(st.Users, conn)
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
func (m *RoomManager) Broadcast(roomID string, sender *websocket.Conn, mt int, msg []byte) {
	m.mu.RLock()
	clients := m.rooms[roomID]
	m.mu.RUnlock()

	for conn := range clients {
		if conn == sender {
			continue
		}
		conn.WriteMessage(mt, msg) // ignore errors for simplicity
	}
}

// StartQuestion marks the room as active and resets fastest user.
func (m *RoomManager) StartQuestion(roomID string) {
	m.mu.Lock()
	st, ok := m.states[roomID]
	if !ok {
		st = &RoomState{}
		m.states[roomID] = st
	}
	st.Active = true
	st.Fastest = ""
	st.BuzzOrder = nil
	m.mu.Unlock()

	go func() {
		time.Sleep(time.Duration(config.TimeLimit) * time.Second)
		m.mu.Lock()
		st := m.states[roomID]
		if st != nil && st.Active && st.Fastest == "" {
			st.Active = false
			m.mu.Unlock()
			resp, _ := json.Marshal(&model.ServerMessage{Type: "timeout", Timestamp: time.Now().UnixMilli()})
			m.Broadcast(roomID, nil, websocket.TextMessage, resp)
			states := m.ResetReady(roomID)
			readyMsg, _ := json.Marshal(&model.ServerMessage{Type: "ready_state", ReadyUsers: states, Timestamp: time.Now().UnixMilli()})
			m.Broadcast(roomID, nil, websocket.TextMessage, readyMsg)
			if vid, err := m.NextVideo(roomID); err == nil {
				videoMsg, _ := json.Marshal(&model.ServerMessage{Type: "video", VideoID: vid, Timestamp: time.Now().UnixMilli()})
				m.Broadcast(roomID, nil, websocket.TextMessage, videoMsg)
			}
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

// AddBuzz appends a user to the buzz order and returns if they were first and the current order.
func (m *RoomManager) AddBuzz(roomID, user string) (bool, []string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	st := m.states[roomID]
	if st == nil {
		return false, nil
	}
	for _, u := range st.BuzzOrder {
		if u == user {
			q := append([]string(nil), st.BuzzOrder...)
			return false, q
		}
	}
	st.BuzzOrder = append(st.BuzzOrder, user)
	first := false
	if st.Active && st.Fastest == "" {
		st.Fastest = user
		st.Active = false
		first = true
	}
	q := append([]string(nil), st.BuzzOrder...)
	return first, q
}

// SetVideoTitle stores the current video's title for answer checking.
func (m *RoomManager) SetVideoTitle(roomID, title string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	st := m.states[roomID]
	if st == nil {
		st = &RoomState{Ready: make(map[string]bool), Users: make(map[*websocket.Conn]string)}
		m.states[roomID] = st
	}
	st.VideoTitle = title
}

// SetPlaylist stores the playlist ID for the room.
func (m *RoomManager) SetPlaylist(roomID, playlistID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	st := m.states[roomID]
	if st == nil {
		st = &RoomState{Ready: make(map[string]bool), Users: make(map[*websocket.Conn]string)}
		m.states[roomID] = st
	}
	st.PlaylistID = playlistID
}

// GetVideoTitle retrieves the stored video title.
func (m *RoomManager) GetVideoTitle(roomID string) string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	st := m.states[roomID]
	if st == nil {
		return ""
	}
	return st.VideoTitle
}

// NextVideo retrieves a random video using the stored playlist ID.
func (m *RoomManager) NextVideo(roomID string) (string, error) {
	m.mu.RLock()
	playlist := ""
	if st, ok := m.states[roomID]; ok {
		playlist = st.PlaylistID
	}
	m.mu.RUnlock()
	if playlist == "" {
		return "", fmt.Errorf("playlist not set")
	}
	apiKey := os.Getenv("YOUTUBE_API_KEY")
	yt := NewYouTubeService(apiKey)
	videoID, title, err := yt.GetRandomVideo(playlist)
	if err != nil {
		return "", err
	}
	m.SetVideoTitle(roomID, title)
	return videoID, nil
}

// SubmitAnswer checks the user's answer and advances to the next if incorrect.
func (m *RoomManager) SubmitAnswer(roomID, user, answer string) (bool, string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	st := m.states[roomID]
	if st == nil {
		return false, ""
	}
	// 正解判定を「タイトルに含まれていれば正解」に変更
	title := strings.ToLower(strings.TrimSpace(st.VideoTitle))
	ans := strings.ToLower(strings.TrimSpace(answer))
	if title != "" && ans != "" && strings.Contains(title, ans) {
		st.Active = false
		st.Fastest = ""
		st.BuzzOrder = nil
		return true, ""
	}
	// remove user from buzz order
	if len(st.BuzzOrder) > 0 {
		if st.BuzzOrder[0] == user {
			st.BuzzOrder = st.BuzzOrder[1:]
		} else {
			for i, u := range st.BuzzOrder {
				if u == user {
					st.BuzzOrder = append(st.BuzzOrder[:i], st.BuzzOrder[i+1:]...)
					break
				}
			}
		}
	}
	if len(st.BuzzOrder) > 0 {
		st.Fastest = st.BuzzOrder[0]
		return false, st.Fastest
	}
	st.Fastest = ""
	return false, ""
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
	manager *RoomManager
	roomID  string
	conn    *websocket.Conn
}

// NewRoomService creates a RoomService for a specific connection and room.
func NewRoomService(m *RoomManager, roomID string, conn *websocket.Conn) *RoomService {
	return &RoomService{manager: m, roomID: roomID, conn: conn}
}

// ProcessMessage broadcasts the received message to the room.
func (r *RoomService) ProcessMessage(mt int, msg []byte) (int, []byte) {
	var req model.ClientMessage
	if err := json.Unmarshal(msg, &req); err != nil {
		return 0, nil
	}

	switch req.Type {
	case "join":
		states := r.manager.RegisterUser(r.roomID, r.conn, req.User)
		resp, _ := json.Marshal(&model.ServerMessage{Type: "ready_state", ReadyUsers: states, Timestamp: time.Now().UnixMilli()})
		r.conn.WriteMessage(websocket.TextMessage, resp)
		r.manager.Broadcast(r.roomID, r.conn, websocket.TextMessage, resp)
	case "playlist":
		r.manager.SetPlaylist(r.roomID, req.PlaylistID)
		videoID, err := r.manager.NextVideo(r.roomID)
		if err != nil {
			break
		}
		resp, _ := json.Marshal(&model.ServerMessage{Type: "video", VideoID: videoID, Timestamp: time.Now().UnixMilli()})
		r.conn.WriteMessage(websocket.TextMessage, resp)
		r.manager.Broadcast(r.roomID, r.conn, websocket.TextMessage, resp)
	case "ready":
		all, states := r.manager.SetReady(r.roomID, req.User)
		resp, _ := json.Marshal(&model.ServerMessage{Type: "ready_state", ReadyUsers: states, Timestamp: time.Now().UnixMilli()})
		r.conn.WriteMessage(websocket.TextMessage, resp)
		r.manager.Broadcast(r.roomID, r.conn, websocket.TextMessage, resp)
		if all {
			r.manager.StartQuestion(r.roomID)
			startMsg, _ := json.Marshal(&model.ServerMessage{Type: "start", Timestamp: time.Now().UnixMilli()})
			r.manager.Broadcast(r.roomID, nil, websocket.TextMessage, startMsg)
		}
	case "start":
		r.manager.StartQuestion(r.roomID)
		resp, _ := json.Marshal(&model.ServerMessage{Type: "start", Timestamp: time.Now().UnixMilli()})
		r.manager.Broadcast(r.roomID, nil, websocket.TextMessage, resp)
	case "buzz":
		// broadcast that someone pressed the answer button
		note, _ := json.Marshal(&model.ServerMessage{Type: "answer", User: req.User, Timestamp: time.Now().UnixMilli()})
		r.manager.Broadcast(r.roomID, r.conn, websocket.TextMessage, note)

		first, order := r.manager.AddBuzz(r.roomID, req.User)
		orderMsg, _ := json.Marshal(&model.ServerMessage{Type: "buzz_order", BuzzOrder: order, Timestamp: time.Now().UnixMilli()})
		r.manager.Broadcast(r.roomID, nil, websocket.TextMessage, orderMsg)

		if first {
			resp, _ := json.Marshal(&model.ServerMessage{Type: "buzz_result", User: req.User, Timestamp: time.Now().UnixMilli()})
			r.manager.Broadcast(r.roomID, nil, websocket.TextMessage, resp)
		}
	case "answer_text":
		correct, next := r.manager.SubmitAnswer(r.roomID, req.User, req.Answer)
		resultMsg, _ := json.Marshal(&model.ServerMessage{Type: "answer_result", User: req.User, Correct: correct, VideoTitle: r.manager.GetVideoTitle(r.roomID), Timestamp: time.Now().UnixMilli()})
		r.manager.Broadcast(r.roomID, nil, websocket.TextMessage, resultMsg)
		if !correct && next != "" {
			nextMsg, _ := json.Marshal(&model.ServerMessage{Type: "buzz_result", User: next, Timestamp: time.Now().UnixMilli()})
			r.manager.Broadcast(r.roomID, nil, websocket.TextMessage, nextMsg)
		}
		if correct || next == "" {
			states := r.manager.ResetReady(r.roomID)
			stateMsg, _ := json.Marshal(&model.ServerMessage{Type: "ready_state", ReadyUsers: states, Timestamp: time.Now().UnixMilli()})
			r.manager.Broadcast(r.roomID, nil, websocket.TextMessage, stateMsg)
			if vid, err := r.manager.NextVideo(r.roomID); err == nil {
				videoMsg, _ := json.Marshal(&model.ServerMessage{Type: "video", VideoID: vid, Timestamp: time.Now().UnixMilli()})
				r.manager.Broadcast(r.roomID, nil, websocket.TextMessage, videoMsg)
			}
		}
	}

	return 0, nil
}
