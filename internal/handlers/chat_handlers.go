package handlers

import (
	"log"
	"net/http"

	"message_service/internal/services"

	"github.com/gin-gonic/gin"
)

type ChatHandler struct {
	chatMemberService *services.ChatMemberService
}

func NewChatHandler(chatMemberService *services.ChatMemberService) *ChatHandler {
	return &ChatHandler{
		chatMemberService: chatMemberService,
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
