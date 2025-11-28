package handlers

import (
	"log"
	"net/http"

	"message_service/internal/services"

	"github.com/gin-gonic/gin"
)

type ChatHandler struct {
	chatMemberService *services.ChatMemberService
	chatService       *services.ChatService
}

func NewChatHandler(chatMemberService *services.ChatMemberService, chatService *services.ChatService) *ChatHandler {
	return &ChatHandler{
		chatMemberService: chatMemberService,
		chatService:       chatService,
	}
}

func (h *ChatHandler) AddUserToChat(c *gin.Context) {
	var req struct {
		ChatID        uint `json:"chat_id" binding:"required"`
		UserID        uint `json:"user_id" binding:"required"`
		CurrentUserID uint `json:"current_user_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	member, err := h.chatMemberService.AddUserToChat(req.ChatID, req.UserID, req.CurrentUserID)
	if err != nil {
		log.Printf("Error adding user to chat: %v", err)

		switch err.Error() {
		case "chat not found":
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		case "only chat members can add new users":
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		case "user is already a member of this chat":
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, member)
}

func (h *ChatHandler) GetUserChats(c *gin.Context) {
	var req struct {
		UserID uint `json:"user_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	chatIDs, err := h.chatMemberService.GetUserChats(req.UserID)
	if err != nil {
		log.Printf("Error getting user chats: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user_id": req.UserID,
		"chats":   chatIDs,
	})
}

func (h *ChatHandler) GetChatByTwoUsers(c *gin.Context) {
	var req struct {
		UserID1 uint `json:"user_id_1" binding:"required"`
		UserID2 uint `json:"user_id_2" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	chat, created, err := h.chatService.CreateChat(req.UserID1, req.UserID2)
	if err != nil {
		log.Printf("Error getting/creating chat: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"chat_id": chat.ID,
		"created": created,
	})
}
