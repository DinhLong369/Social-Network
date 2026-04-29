package app

import (
	"sync"

	fiberws "github.com/gofiber/contrib/v3/websocket"
)

// Hub quản lý tất cả kết nối WebSocket đang active
type Hub struct {
	clients map[string]*fiberws.Conn
	mu      sync.RWMutex
}

// GlobalHub là instance dùng chung cho toàn app
var GlobalHub = newHub()

func newHub() *Hub {
	return &Hub{
		clients: make(map[string]*fiberws.Conn),
	}
}

// Register đăng ký kết nối mới cho user
func (h *Hub) Register(userID string, conn *fiberws.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	// Đóng kết nối cũ nếu user đã kết nối từ thiết bị khác
	if old, ok := h.clients[userID]; ok {
		old.Close()
	}
	h.clients[userID] = conn
}

// Unregister huỷ đăng ký khi user ngắt kết nối
func (h *Hub) Unregister(userID string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.clients, userID)
}

// Send gửi message tới một user cụ thể, trả về true nếu user đang online
func (h *Hub) Send(userID string, msg []byte) bool {
	h.mu.RLock()
	conn, ok := h.clients[userID]
	h.mu.RUnlock()
	if !ok {
		return false
	}
	conn.WriteMessage(fiberws.TextMessage, msg)
	return true
}

// IsOnline kiểm tra user có đang kết nối không
func (h *Hub) IsOnline(userID string) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	_, ok := h.clients[userID]
	return ok
}
