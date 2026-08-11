package eventbus

import (
	"sync"

	"github.com/accessible-path/notification-service/internal/ports/out"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type WebSocketManager struct {
	mu       sync.RWMutex
	clients  map[uuid.UUID]*websocket.Conn
	broadcast chan map[string]interface{}
}

func NewWebSocketManager() *WebSocketManager {
	return &WebSocketManager{
		clients:   make(map[uuid.UUID]*websocket.Conn),
		broadcast: make(chan map[string]interface{}, 256),
	}
}

func (m *WebSocketManager) Register(userID uuid.UUID, conn interface{}) error {
	ws, ok := conn.(*websocket.Conn)
	if !ok {
		return nil
	}
	
	m.mu.Lock()
	m.clients[userID] = ws
	m.mu.Unlock()
	return nil
}

func (m *WebSocketManager) Unregister(userID uuid.UUID) error {
	m.mu.Lock()
	delete(m.clients, userID)
	m.mu.Unlock()
	return nil
}

func (m *WebSocketManager) Broadcast(message map[string]interface{}) error {
	m.broadcast <- message
	return nil
}

func (m *WebSocketManager) Start() {
	for msg := range m.broadcast {
		m.mu.RLock()
		for _, conn := range m.clients {
			if err := conn.WriteJSON(msg); err != nil {
				conn.Close()
			}
		}
		m.mu.RUnlock()
	}
}

var _ out.WebSocketManager = (*WebSocketManager)(nil)