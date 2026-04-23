package handlers

import (
	"log"
	"net/http"
	"strconv"

	"message_service/internal/dto"
	"message_service/internal/services"

	"github.com/gin-gonic/gin"
)

type MessageHandler struct {
	messageService *services.MessageService
}

func NewMessageHandler(messageService *services.MessageService) *MessageHandler {
	return &MessageHandler{
		messageService: messageService,
	}
}

// CreateMessage godoc
// @Summary Create message
// @Description Создаёт новое сообщение в чате
// @Tags messages
// @Accept json
// @Produce json
// @Param X-Auth-User-ID header string true "Authenticated user ID"
// @Param chat_id path int true "Chat ID"
// @Param request body dto.CreateMessageRequest true "Create message request"
// @Success 201 {object} dto.MessageResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /chats/{chat_id}/messages [post]
func (h *MessageHandler) CreateMessage(c *gin.Context) {
	chatIDStr := c.Param("chat_id")
	chatID, err := strconv.ParseUint(chatIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid chat_id"})
		return
	}

	userIDStr := c.GetHeader("X-Auth-User-ID")
	userID, err := strconv.ParseUint(userIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id is required"})
		return
	}

	var req dto.CreateMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	message, err := h.messageService.CreateMessage(uint(chatID), uint(userID), req.Content)
	if err != nil {
		switch err.Error() {
		case "chat not found":
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		case "sender is not a member of this chat":
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		case "invalid chat ID", "invalid sender ID", "message cannot be empty":
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusCreated, message)
}

// GetMessage godoc
// @Summary Get message
// @Description Возвращает одно сообщение по ID
// @Tags messages
// @Produce json
// @Param X-Auth-User-ID header string true "Authenticated user ID"
// @Param message_id path int true "Message ID"
// @Success 200 {object} dto.MessageResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /messages/{message_id} [get]
func (h *MessageHandler) GetMessage(c *gin.Context) {
	userIDStr := c.GetHeader("X-Auth-User-ID")
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

	message, err := h.messageService.GetMessage(uint(messageID), userID)
	if err != nil {
		switch err.Error() {
		case "message not found":
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		case "user is not a member of this chat":
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		case "invalid message ID", "invalid user ID":
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, message)
}

// GetMessages godoc
// @Summary Get messages by chat
// @Description Возвращает сообщения чата с пагинацией
// @Tags messages
// @Produce json
// @Param X-Auth-User-ID header string true "Authenticated user ID"
// @Param chat_id path int true "Chat ID"
// @Param page query int false "Page number" default(1)
// @Param page_size query int true "Page size"
// @Success 200 {object} dto.MessagesResponse
// @Failure 400 {object} dto.ErrorResponse
// @Router /messages/{chat_id} [get]
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

	userIDStr := c.GetHeader("X-Auth-User-ID")
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

	c.JSON(http.StatusOK, messages)
}

// ChangeMessage godoc
// @Summary Change message
// @Description Изменяет текст сообщения
// @Tags messages
// @Accept json
// @Produce json
// @Param X-Auth-User-ID header string true "Authenticated user ID"
// @Param message_id path int true "Message ID"
// @Param request body dto.ChangeMessageRequest true "Change message request"
// @Success 200 {object} dto.MessageResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /messages/{message_id} [put]
func (h *MessageHandler) ChangeMessage(c *gin.Context) {
	userIDStr := c.GetHeader("X-Auth-User-ID")
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

	var req dto.ChangeMessageRequest

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

// DeleteMessage godoc
// @Summary Delete message
// @Description Удаляет сообщение
// @Tags messages
// @Produce json
// @Param X-Auth-User-ID header string true "Authenticated user ID"
// @Param message_id path int true "Message ID"
// @Success 204 {string} string ""
// @Failure 400 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /messages/{message_id} [delete]
func (h *MessageHandler) DeleteMessage(c *gin.Context) {
	userIDStr := c.GetHeader("X-Auth-User-ID")
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

	err = h.messageService.DeleteMessage(uint(messageID), userID)
	if err != nil {
		switch err.Error() {
		case "message not found":
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		case "user cannot delete this message", "user is not a member of this chat":
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		case "invalid message ID", "invalid user ID":
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.Status(http.StatusNoContent)
}
