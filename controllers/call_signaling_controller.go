package controllers

import (
	"log"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/raihansyahrin/backend_laundry_app.git/utils"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow cross-origin for dev & testing
	},
}

type CallMessage struct {
	Type         string      `json:"type"`           // "CALL_OFFER", "CALL_RINGING", "CALL_ANSWER", "CALL_REJECT", "CALL_END", "ICE_CANDIDATE"
	CallerUserID uint        `json:"caller_user_id"` // Sender ID
	TargetUserID uint        `json:"target_user_id"` // Receiver ID
	OrderID      uint        `json:"order_id"`
	CallerName   string      `json:"caller_name,omitempty"`
	SDP          string      `json:"sdp,omitempty"`
	Candidate    interface{} `json:"candidate,omitempty"`
	Reason       string      `json:"reason,omitempty"`
}

type Connection struct {
	UserID uint
	Conn   *websocket.Conn
	Send   chan CallMessage
}

type Hub struct {
	clients    map[uint]*Connection
	register   chan *Connection
	unregister chan *Connection
	broadcast  chan CallMessage
	mu         sync.RWMutex
}

var GlobalHub = Hub{
	clients:    make(map[uint]*Connection),
	register:   make(chan *Connection),
	unregister: make(chan *Connection),
	broadcast:  make(chan CallMessage),
}

func init() {
	go GlobalHub.run()
}

func (h *Hub) run() {
	for {
		select {
		case conn := <-h.register:
			h.mu.Lock()
			h.clients[conn.UserID] = conn
			h.mu.Unlock()
			log.Printf("[WebSocket Call Hub] User #%d connected", conn.UserID)

		case conn := <-h.unregister:
			h.mu.Lock()
			if existing, ok := h.clients[conn.UserID]; ok && existing == conn {
				delete(h.clients, conn.UserID)
				close(conn.Send)
				log.Printf("[WebSocket Call Hub] User #%d disconnected", conn.UserID)
			}
			h.mu.Unlock()

		case msg := <-h.broadcast:
			h.mu.RLock()
			targetConn, exists := h.clients[msg.TargetUserID]
			h.mu.RUnlock()

			if exists {
				select {
				case targetConn.Send <- msg:
				default:
					h.mu.Lock()
					delete(h.clients, targetConn.UserID)
					close(targetConn.Send)
					h.mu.Unlock()
				}
			} else {
				log.Printf("[WebSocket Call Hub] Target User #%d not connected for msg '%s'", msg.TargetUserID, msg.Type)
			}
		}
	}
}

// HandleCallSignaling - WebSocket endpoint: ws://localhost:8083/api/ws/call?token=<JWT>
func HandleCallSignaling(c *gin.Context) {
	tokenStr := c.Query("token")
	if tokenStr == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Parameter token wajib diisi"})
		return
	}

	claims, err := utils.ValidateJWT(tokenStr)
	if err != nil || claims.UserID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Token tidak valid atau kadaluarsa"})
		return
	}

	userID := claims.UserID

	ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("[WebSocket Upgrade Error]: %v", err)
		return
	}

	conn := &Connection{
		UserID: userID,
		Conn:   ws,
		Send:   make(chan CallMessage, 256),
	}

	GlobalHub.register <- conn

	// Writer goroutine
	go func() {
		defer ws.Close()
		for msg := range conn.Send {
			if err := ws.WriteJSON(msg); err != nil {
				break
			}
		}
	}()

	// Reader loop
	defer func() {
		GlobalHub.unregister <- conn
		ws.Close()
	}()

	for {
		var msg CallMessage
		if err := ws.ReadJSON(&msg); err != nil {
			break
		}

		// Set caller_user_id to authenticated UserID
		msg.CallerUserID = userID

		// Route message to target user if target_user_id > 0
		if msg.TargetUserID > 0 {
			GlobalHub.broadcast <- msg
		}
	}
}
