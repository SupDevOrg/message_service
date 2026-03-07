package handlers

import (
	"log"
	"net/http"
	"strconv"

	"message_service/internal/models"
	"message_service/internal/services"

	"github.com/gin-gonic/gin"
)

type MessageHandler struct {
	messageService *services.MessageService
}

type MessagesResponse struct {
	Messages []models.Message `json:"messages"`
}

func NewMessageHandler(messageService *services.MessageService) *MessageHandler {
	return &MessageHandler{
		messageService: messageService,
	}
}

func (h *MessageHandler) GetMessages(c *gin.Context) {
	chatIDStr := c.Param("chat_id")
	chatID, err := strconv.ParseUint(chatIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid chat_id"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))

	pageSize, err := strconv.Atoi(c.Query("page_size"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "page_size is required"})
		return
	}

	userIDStr := c.GetHeader("X-Auth-User-Id")
	userID, err := strconv.ParseUint(userIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id is required"})
		return
	}

	serviceReq := services.MessagesPaginationRequest{
		Chat:     uint(chatID),
		PageNum:  page,
		PageSize: pageSize,
	}

	messages, err := h.messageService.GetMessages(serviceReq, uint(userID))
	if err != nil {
		log.Printf("Error getting messages: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, MessagesResponse{Messages: *messages})
}

func (h *MessageHandler) ChangeMessage(c *gin.Context) {
	userIDStr := c.GetHeader("X-Auth-User-Id")
	userID64, err := strconv.ParseUint(userIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id is required"})
		return
	}
	userID := uint(userID64)

	messageIDStr := c.Param("message_id")
	messageID, err := strconv.ParseUint(messageIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid message_id"})
		return
	}

	var req struct {
		Content string `json:"content" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updatedMsg, err := h.messageService.ChangeMessage(uint(messageID), userID, req.Content)
	if err != nil {
		if err.Error() == "message not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		if err.Error() == "user cannot change message" {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, updatedMsg)
}
