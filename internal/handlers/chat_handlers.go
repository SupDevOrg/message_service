package handlers

import (
	"log"
	"message_service/internal/services"
	"net/http"
	"strconv"

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
	userIDStr := c.GetHeader("X-Auth-User-ID")
	userID, err := strconv.ParseUint(userIDStr, 10, 64)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id is required"})
		return
	}

	chatIDStr := c.Param("chat_id")
	chatID, err := strconv.ParseUint(chatIDStr, 10, 64)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid chat id"})
		return
	}

	var req struct {
		UserID uint `json:"user_id" binding:"required"` // Кого добавляем
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	member, err := h.chatMemberService.AddUserToChat(uint(chatID), req.UserID, uint(userID))
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
	userIDStr := c.GetHeader("X-Auth-User-ID")
	userID, err := strconv.ParseUint(userIDStr, 10, 64)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id is required"})
		return
	}

	chatIDs, err := h.chatMemberService.GetUserChats(uint(userID))
	if err != nil {
		log.Printf("Error getting user chats: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user_id": userID,
		"chats":   chatIDs,
	})
}

func (h *ChatHandler) GetChatMembers(c *gin.Context) {
	userIDStr := c.GetHeader("X-Auth-User-ID")
	userID64, err := strconv.ParseUint(userIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}
	userID := uint(userID64)

	chatIDStr := c.Param("chat_id")
	chatID64, err := strconv.ParseUint(chatIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid chat id"})
		return
	}
	chatID := uint(chatID64)

	members, err := h.chatMemberService.GetChatMembers(chatID, userID)
	if err != nil {
		if err.Error() == "only chat members can view member list" {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get chat members"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"chat_id": chatID,
		"members": members,
	})
}

func (h *ChatHandler) CreateChat(c *gin.Context) {
	userIDStr := c.GetHeader("X-Auth-User-ID")
	userID, err := strconv.ParseUint(userIDStr, 10, 64)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id is required"})
		return
	}

	var req struct {
		PartnerUserID uint `json:"user_id" binding:"required"`
	}

	chat, created, err := h.chatService.CreateChat(uint(userID), req.PartnerUserID)
	if err != nil {
		log.Printf("Failed to create/find chat: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"chat_id": chat.ID,
		"created": created,
	})
}

func (h *ChatHandler) CreateGroupChat(c *gin.Context) {
	userIDStr := c.GetHeader("X-Auth-User-ID")
	userID64, err := strconv.ParseUint(userIDStr, 10, 64)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id is required"})
		return
	}

	userID := uint(userID64)

	var req struct {
		GroupMates []uint `json:"user_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	chat, created, err := h.chatService.CreateGroup(userID, req.GroupMates)
	if err != nil {
		log.Printf("Failed to create/find chat: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create/find chat"})
		return
	}

	if created {
		log.Printf("Created new chat: %d between users %d and %d", chat.ID, userID, req.GroupMates)
	}

	c.JSON(http.StatusOK, gin.H{
		"chat_id":  chat.ID,
		"is_group": chat.IsGroup,
	})
}
