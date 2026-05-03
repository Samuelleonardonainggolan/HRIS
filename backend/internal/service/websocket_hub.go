// internal/service/websocket_hub.go
// WebSocket Hub: mengelola koneksi per user dan broadcast event real-time.
package service

import (
	"encoding/json"
	"log"
	"sync"
	"time"
)

// WSEventType adalah tipe event yang dikirim via WebSocket.
type WSEventType string

const (
	WSEventAttendanceUpdated WSEventType = "attendance_updated"
	WSEventLeaveUpdated      WSEventType = "leave_updated"
	WSEventStatsUpdated      WSEventType = "stats_updated"
	WSEventPing              WSEventType = "ping"
)

// WSEvent adalah struktur pesan WebSocket.
type WSEvent struct {
	Type    WSEventType    `json:"type"`
	Payload map[string]any `json:"payload,omitempty"`
	Time    time.Time      `json:"time"`
}

// WSClient mewakili satu koneksi SSE/WebSocket user (exported untuk digunakan di handler).
type WSClient struct {
	UserID string
	Send   chan []byte
	hub    *WSHub
}



// WSHub mengelola semua koneksi WebSocket yang aktif.
type WSHub struct {
	mu      sync.RWMutex
	clients map[string][]*WSClient // userID → list of connections
}

// NewWSHub membuat hub baru.
func NewWSHub() *WSHub {
	return &WSHub{
		clients: make(map[string][]*WSClient),
	}
}

// RegisterClient mendaftarkan client baru ke hub.
func (h *WSHub) RegisterClient(c *WSClient) {
	c.hub = h
	h.mu.Lock()
	defer h.mu.Unlock()
	h.clients[c.UserID] = append(h.clients[c.UserID], c)
	log.Printf("[WS Hub] Client registered: userID=%s, total=%d", c.UserID, len(h.clients[c.UserID]))
}

// UnregisterClient menghapus client dari hub.
func (h *WSHub) UnregisterClient(c *WSClient) {
	h.mu.Lock()
	defer h.mu.Unlock()
	list := h.clients[c.UserID]
	newList := make([]*WSClient, 0, len(list))
	for _, cl := range list {
		if cl != c {
			newList = append(newList, cl)
		}
	}
	if len(newList) == 0 {
		delete(h.clients, c.UserID)
	} else {
		h.clients[c.UserID] = newList
	}
	close(c.Send)
	log.Printf("[WS Hub] Client unregistered: userID=%s", c.UserID)
}



// BroadcastToUser mengirim event ke semua koneksi milik userID tertentu.
func (h *WSHub) BroadcastToUser(userID string, eventType WSEventType, payload map[string]any) {
	event := WSEvent{
		Type:    eventType,
		Payload: payload,
		Time:    time.Now(),
	}
	data, err := json.Marshal(event)
	if err != nil {
		log.Printf("[WS Hub] Marshal error: %v", err)
		return
	}

	h.mu.RLock()
	clients := h.clients[userID]
	h.mu.RUnlock()

	for _, c := range clients {
		select {
		case c.Send <- data:
		default:
			// Buffer penuh, skip
			log.Printf("[WS Hub] Buffer penuh untuk userID=%s, pesan dilewati", userID)
		}
	}
}

// BroadcastToAll mengirim event ke semua koneksi yang aktif.
func (h *WSHub) BroadcastToAll(eventType WSEventType, payload map[string]any) {
	event := WSEvent{
		Type:    eventType,
		Payload: payload,
		Time:    time.Now(),
	}
	data, err := json.Marshal(event)
	if err != nil {
		return
	}

	h.mu.RLock()
	defer h.mu.RUnlock()
	for _, list := range h.clients {
		for _, c := range list {
			select {
			case c.Send <- data:
			default:
			}
		}
	}
}

// ActiveConnectionCount mengembalikan jumlah koneksi aktif.
func (h *WSHub) ActiveConnectionCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	count := 0
	for _, list := range h.clients {
		count += len(list)
	}
	return count
}
