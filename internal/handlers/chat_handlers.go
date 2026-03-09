package handlers

import (
	"log"
	"message_service/internal/dto"
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

	var req dto.AddUserToChatRequest

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

	chats, err := h.chatMemberService.GetUserChats(uint(userID))
	if err != nil {
		log.Printf("Error getting user chats: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp := dto.GetUserChatsResponse{
		UserID: uint(userID),
		Chats:  chats,
	}

	c.JSON(http.StatusOK, resp)
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

	resp := dto.GetChatMembersResponse{
		ChatID:  chatID,
		Members: make([]dto.UserDTO, len(members)),
	}

	for i, mmbrs := range members {
		resp.Members[i] = dto.UserDTO{ID: mmbrs}
	}

	c.JSON(http.StatusOK, resp)
}

func (h *ChatHandler) CreateChat(c *gin.Context) {
	userIDStr := c.GetHeader("X-Auth-User-ID")
	userID, err := strconv.ParseUint(userIDStr, 10, 64)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id is required"})
		return
	}

	var req dto.CreateChatRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	chat, created, err := h.chatService.CreateChat(uint(userID), req.UserID)
	if err != nil {
		log.Printf("Failed to create/find chat: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.CreateChatResponse{
		Chat: dto.ChatDTO{ID: chat.ID,
			CreatedAt: chat.CreatedAt,
			ChatName:  chat.ChatName,
			IsGroup:   chat.IsGroup},
		Created: created,
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

	var req dto.CreateGroupChatRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	chat, created, err := h.chatService.CreateGroup(userID)
	if err != nil {
		log.Printf("Failed to create/find chat: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create/find chat"})
		return
	}

	if created {
		log.Printf("Created new group chat %d, owner: %d", chat.ID, userID)
	}

	usersIDs := make([]uint, 0, len(req.Users))
	for _, u := range req.Users {
		usersIDs = append(usersIDs, u.ID)
	}

	if err := h.chatMemberService.AddUsersToChat(chat.ID, usersIDs, userID); err != nil {
		log.Printf("Failed to add users to group chat: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to add users to group chat"})
		return
	}

	c.JSON(http.StatusOK, dto.CreateGroupChatResponse{
		Chat: dto.ChatDTO{ID: chat.ID,
			CreatedAt: chat.CreatedAt,
			ChatName:  chat.ChatName,
			IsGroup:   chat.IsGroup},
	})
}
