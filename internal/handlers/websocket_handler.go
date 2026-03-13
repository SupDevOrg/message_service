package handlers

import (
	"log"
	"net/http"
	"strconv"

	"message_service/internal/services"
	ws "message_service/internal/websocket"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type WebSocketHandler struct {
	hub         *ws.Hub
	chatService *services.ChatService
}

func NewWebSocketHandler(hub *ws.Hub, chatService *services.ChatService) *WebSocketHandler {
	return &WebSocketHandler{
		hub:         hub,
		chatService: chatService,
	}
}

func (h *WebSocketHandler) HandleWebSocket(c *gin.Context) {
	userIDStr := c.GetHeader("X-Auth-User-ID")
	userID, err := strconv.ParseUint(userIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id is required"})
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}

	client := &ws.Client{
		Hub:    h.hub,
		Socket: conn,
		Recive: make(chan []byte, 256),
		UserID: uint(userID),
	}

	h.hub.Register <- client

	go client.Write()
	go client.Read()

	log.Printf("WebSocket connected: user=%d", userID)
}
