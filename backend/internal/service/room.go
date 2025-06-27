package service

import (
	"encoding/json"
	"fmt"
	"math/rand"
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
	Fastest         string
	Active          bool
	Ready           map[string]bool
	Users           map[*websocket.Conn]string
	AnswerRights    map[string]bool
	VideoTitle      string
	PlaylistID      string
	RemainingVideos []VideoItem
	TimeoutCancel   chan struct{}
	TimeLeft        time.Duration
	TimerStarted    time.Time
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
		st = &RoomState{Ready: make(map[string]bool), Users: make(map[*websocket.Conn]string)}
		m.states[roomID] = st
	}
	if st.TimeoutCancel != nil {
		close(st.TimeoutCancel)
		st.TimeoutCancel = nil
	}
	st.Active = true
	st.Fastest = ""
	st.AnswerRights = make(map[string]bool)
	for _, name := range st.Users {
		st.AnswerRights[name] = true
	}
	st.TimeLeft = time.Duration(config.TimeLimit) * time.Second
	st.TimerStarted = time.Now()
	st.TimeoutCancel = make(chan struct{})
	cancel := st.TimeoutCancel
	m.mu.Unlock()

	go m.runTimer(roomID, cancel)
}

// runTimer handles the countdown for a question.
func (m *RoomManager) runTimer(roomID string, cancel chan struct{}) {
	m.mu.RLock()
	st := m.states[roomID]
	duration := st.TimeLeft
	m.mu.RUnlock()
	timer := time.NewTimer(duration)
	start := time.Now()
	select {
	case <-timer.C:
		m.mu.Lock()
		st := m.states[roomID]
		if st != nil && st.Active && st.Fastest == "" {
			st.Active = false
			for k := range st.AnswerRights {
				st.AnswerRights[k] = false
			}
			m.mu.Unlock()
			resp, _ := json.Marshal(&model.ServerMessage{Type: "timeout", Timestamp: time.Now().UnixMilli()})
			m.Broadcast(roomID, nil, websocket.TextMessage, resp)
			m.prepareNext(roomID)
			return
		}
		m.mu.Unlock()
	case <-cancel:
		if !timer.Stop() {
			<-timer.C
		}
		m.mu.Lock()
		if st := m.states[roomID]; st != nil {
			st.TimeLeft -= time.Since(start)
		}
		m.mu.Unlock()
	}
}

// resumeTimer restarts the countdown with remaining time.
func (m *RoomManager) resumeTimer(roomID string) {
	m.mu.Lock()
	st := m.states[roomID]
	if st == nil || st.TimeLeft <= 0 {
		m.mu.Unlock()
		return
	}
	st.Active = true
	st.TimerStarted = time.Now()
	st.TimeoutCancel = make(chan struct{})
	cancel := st.TimeoutCancel
	m.mu.Unlock()
	go m.runTimer(roomID, cancel)
}

// AddBuzz registers a user's intent to answer.
// It returns true if the user successfully obtained the answer right.
func (m *RoomManager) AddBuzz(roomID, user string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	st := m.states[roomID]
	if st == nil || !st.Active || st.Fastest != "" || !st.AnswerRights[user] {
		return false
	}
	st.Fastest = user
	st.Active = false
	if st.TimeoutCancel != nil {
		close(st.TimeoutCancel)
		st.TimeoutCancel = nil
		st.TimeLeft -= time.Since(st.TimerStarted)
	}
	return true
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
func (m *RoomManager) SetPlaylist(roomID, playlistID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	st := m.states[roomID]
	if st == nil {
		st = &RoomState{Ready: make(map[string]bool), Users: make(map[*websocket.Conn]string)}
		m.states[roomID] = st
	}
	st.PlaylistID = playlistID
	apiKey := os.Getenv("YOUTUBE_API_KEY")
	yt := NewYouTubeService(apiKey)
	videos, err := yt.ListPlaylistVideos(playlistID)
	if err != nil {
		return err
	}
	st.RemainingVideos = videos
	return nil
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
	m.mu.Lock()
	defer m.mu.Unlock()
	st := m.states[roomID]
	if st == nil || st.PlaylistID == "" {
		return "", fmt.Errorf("playlist not set")
	}
	if len(st.RemainingVideos) == 0 {
		apiKey := os.Getenv("YOUTUBE_API_KEY")
		yt := NewYouTubeService(apiKey)
		vids, err := yt.ListPlaylistVideos(st.PlaylistID)
		if err != nil {
			return "", err
		}
		st.RemainingVideos = vids
	}
	if len(st.RemainingVideos) == 0 {
		return "", fmt.Errorf("no videos available")
	}
	// Go 1.20以降はrand.Seedでの初期化は不要です
	for len(st.RemainingVideos) > 0 {
		idx := rand.Intn(len(st.RemainingVideos))
		item := st.RemainingVideos[idx]
		st.RemainingVideos = append(st.RemainingVideos[:idx], st.RemainingVideos[idx+1:]...)
		emb, err := CheckEmbeddable(item.ID)
		if err != nil || !emb {
			continue
		}
		st.VideoTitle = item.Title
		return item.ID, nil
	}
	return "", fmt.Errorf("no embeddable videos found")
}

// advanceQuestion moves the room to the next quiz item immediately.
func (m *RoomManager) advanceQuestion(roomID string) {
	vid, err := m.NextVideo(roomID)
	if err != nil {
		return
	}
	videoMsg, _ := json.Marshal(&model.ServerMessage{Type: "video", VideoID: vid, Timestamp: time.Now().UnixMilli()})
	m.Broadcast(roomID, nil, websocket.TextMessage, videoMsg)
	m.StartQuestion(roomID)
	startMsg, _ := json.Marshal(&model.ServerMessage{Type: "start", Timestamp: time.Now().UnixMilli()})
	m.Broadcast(roomID, nil, websocket.TextMessage, startMsg)
}

// prepareNext resets ready states and notifies users.
func (m *RoomManager) prepareNext(roomID string) {
	states := m.ResetReady(roomID)
	readyMsg, _ := json.Marshal(&model.ServerMessage{Type: "ready_state", ReadyUsers: states, Timestamp: time.Now().UnixMilli()})
	m.Broadcast(roomID, nil, websocket.TextMessage, readyMsg)
}

// SubmitAnswer checks the user's answer and advances to the next if incorrect.
func (m *RoomManager) SubmitAnswer(roomID, user, answer string) (bool, bool) {
	m.mu.Lock()
	st := m.states[roomID]
	if st == nil || st.Fastest != user {
		m.mu.Unlock()
		return false, false
	}
	title := strings.ToLower(strings.TrimSpace(st.VideoTitle))
	ans := strings.ToLower(strings.TrimSpace(answer))
	if title != "" && ans != "" && strings.Contains(title, ans) {
		st.Active = false
		st.Fastest = ""
		for k := range st.AnswerRights {
			st.AnswerRights[k] = false
		}
		m.mu.Unlock()
		m.prepareNext(roomID)
		return true, false
	}
	st.AnswerRights[user] = false
	st.Fastest = ""
	remaining := false
	for _, v := range st.AnswerRights {
		if v {
			remaining = true
			break
		}
	}
	m.mu.Unlock()
	if remaining {
		m.resumeTimer(roomID)
	} else {
		m.prepareNext(roomID)
	}
	return false, remaining
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
		if err := r.manager.SetPlaylist(r.roomID, req.PlaylistID); err != nil {
			break
		}
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
			r.manager.advanceQuestion(r.roomID)
		}
	case "start":
		r.manager.StartQuestion(r.roomID)
		resp, _ := json.Marshal(&model.ServerMessage{Type: "start", Timestamp: time.Now().UnixMilli()})
		r.manager.Broadcast(r.roomID, nil, websocket.TextMessage, resp)
	case "buzz":
		note, _ := json.Marshal(&model.ServerMessage{Type: "answer", User: req.User, Timestamp: time.Now().UnixMilli()})
		r.manager.Broadcast(r.roomID, r.conn, websocket.TextMessage, note)

		if r.manager.AddBuzz(r.roomID, req.User) {
			resp, _ := json.Marshal(&model.ServerMessage{Type: "buzz_result", User: req.User, Timestamp: time.Now().UnixMilli()})
			r.manager.Broadcast(r.roomID, nil, websocket.TextMessage, resp)
		}
	case "answer_text":
		correct, remain := r.manager.SubmitAnswer(r.roomID, req.User, req.Answer)
		resultMsg, _ := json.Marshal(&model.ServerMessage{Type: "answer_result", User: req.User, Correct: correct, VideoTitle: r.manager.GetVideoTitle(r.roomID), Timestamp: time.Now().UnixMilli()})
		r.manager.Broadcast(r.roomID, nil, websocket.TextMessage, resultMsg)
		if remain {
			resumeMsg, _ := json.Marshal(&model.ServerMessage{Type: "resume", Timestamp: time.Now().UnixMilli()})
			r.manager.Broadcast(r.roomID, nil, websocket.TextMessage, resumeMsg)
		}
	}

	return 0, nil
}
