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
	userID1, err := strconv.ParseUint(c.Query("user_id_1"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id_1 is required"})
		return
	}

	userID2, err := strconv.ParseUint(c.Query("user_id_2"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id_2 is required"})
		return
	}

	chat, created, err := h.chatService.CreateChat(uint(userID1), uint(userID2))
	if err != nil {
		log.Printf("Failed to create/find chat: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create/find chat"})
		return
	}

	if created {
		log.Printf("Created new chat: %d between users %d and %d", chat.ID, userID1, userID2)
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
		UserID: uint(userID1),
		ChatID: chat.ID,
	}

	h.hub.Register <- client

	go client.Write()
	go client.Read()

	log.Printf("WebSocket connected: user=%d, chat=%d", userID1, chat.ID)
}
