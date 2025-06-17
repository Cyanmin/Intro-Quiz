package service

import (
	"sync"

	"github.com/gorilla/websocket"
)

// RoomManager manages WebSocket connections grouped by room ID.
type RoomManager struct {
	rooms map[string]map[*websocket.Conn]bool
	mu    sync.RWMutex
}

// NewRoomManager creates a new RoomManager.
func NewRoomManager() *RoomManager {
	return &RoomManager{rooms: make(map[string]map[*websocket.Conn]bool)}
}

// Join adds a connection to a room.
func (m *RoomManager) Join(roomID string, conn *websocket.Conn) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.rooms[roomID]; !ok {
		m.rooms[roomID] = make(map[*websocket.Conn]bool)
	}
	m.rooms[roomID][conn] = true
}

// Leave removes a connection from a room.
func (m *RoomManager) Leave(roomID string, conn *websocket.Conn) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if clients, ok := m.rooms[roomID]; ok {
		delete(clients, conn)
		if len(clients) == 0 {
			delete(m.rooms, roomID)
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
	r.manager.Broadcast(r.roomID, r.conn, mt, msg)
	return 0, nil
}
