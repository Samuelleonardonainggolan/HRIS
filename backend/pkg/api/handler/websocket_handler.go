// pkg/api/handler/websocket_handler.go
// Handler untuk koneksi real-time menggunakan Server-Sent Events (SSE).
// SSE dipilih karena: (1) tidak perlu library tambahan, (2) native di HTTP/1.1,
// (3) otomatis reconnect, (4) cocok untuk push data satu arah server→client.
// Flutter menggunakan http package biasa untuk subscribe ke SSE stream.
package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/andikatampubolon10/hris-backend/internal/service"
	"github.com/andikatampubolon10/hris-backend/pkg/auth"
	"github.com/gin-gonic/gin"
)

// SSEHandler menangani koneksi Server-Sent Events.
type SSEHandler struct {
	hub       *service.WSHub
	jwtSecret string
}

// NewSSEHandler membuat handler SSE baru.
func NewSSEHandler(hub *service.WSHub, jwtSecret string) *SSEHandler {
	return &SSEHandler{hub: hub, jwtSecret: jwtSecret}
}

// Connect adalah endpoint SSE: GET /api/v1/realtime/connect
// Client Flutter subscribe ke endpoint ini dengan Authorization header atau ?token= query param.
func (h *SSEHandler) Connect(c *gin.Context) {
	// Ambil token dari Authorization header atau query param (untuk EventSource JS/Flutter)
	userID := ""
	token := ""

	authHeader := c.GetHeader("Authorization")
	if strings.HasPrefix(authHeader, "Bearer ") {
		token = strings.TrimPrefix(authHeader, "Bearer ")
	} else {
		token = c.Query("token")
	}

	if token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "token diperlukan"})
		return
	}

	claims, err := auth.ValidateToken(token, h.jwtSecret)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "token tidak valid"})
		return
	}
	userID = claims.UserID

	// Set headers untuk SSE
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no") // Untuk Nginx: nonaktifkan buffering
	c.Header("Access-Control-Allow-Origin", "*")

	// Buat client channel
	sendChan := make(chan []byte, 32)
	client := &service.WSClient{
		UserID: userID,
		Send:   sendChan,
	}

	h.hub.RegisterClient(client)
	defer h.hub.UnregisterClient(client)

	log.Printf("[SSE] Client connected: userID=%s, IP=%s", userID, c.ClientIP())

	// Kirim event "connected" pertama kali
	connectMsg, _ := json.Marshal(service.WSEvent{
		Type:    service.WSEventAttendanceUpdated,
		Payload: map[string]any{"status": "connected", "user_id": userID},
		Time:    time.Now(),
	})
	fmt.Fprintf(c.Writer, "data: %s\n\n", connectMsg)
	c.Writer.Flush()

	// Ping ticker setiap 30 detik agar koneksi tidak timeout
	pingTicker := time.NewTicker(30 * time.Second)
	defer pingTicker.Stop()

	// Loop utama: dengarkan event dan kirim ke client
	ctx := c.Request.Context()
	for {
		select {
		case <-ctx.Done():
			log.Printf("[SSE] Client disconnected: userID=%s", userID)
			return
		case msg, ok := <-sendChan:
			if !ok {
				return
			}
			fmt.Fprintf(c.Writer, "data: %s\n\n", msg)
			c.Writer.Flush()
		case <-pingTicker.C:
			pingMsg, _ := json.Marshal(service.WSEvent{
				Type: service.WSEventPing,
				Time: time.Now(),
			})
			fmt.Fprintf(c.Writer, "data: %s\n\n", pingMsg)
			c.Writer.Flush()
		}
	}
}
