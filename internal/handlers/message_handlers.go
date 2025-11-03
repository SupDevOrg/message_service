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

	userID, err := strconv.ParseUint(c.Query("user_id"), 10, 64)
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
