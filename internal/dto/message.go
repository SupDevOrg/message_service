package dto

import "time"

type CreateMessageRequest struct {
	Content string `json:"content" binding:"required" example:"пипяо"`
}

type ChangeMessageRequest struct {
	Content string `json:"content" binding:"required" example:"новый пипяо"`
}

type MessagesResponse struct {
	Messages []MessageResponse `json:"messages"`
}

type MessageResponse struct {
	ID        uint      `json:"message_id" example:"1"`
	ChatID    uint      `json:"chat_id" example:"1"`
	SenderID  uint      `json:"sender_id" example:"1"`
	Content   string    `json:"content" example:"пипяо"`
	CreatedAt time.Time `json:"created_at" example:"2026-03-08T14:30:00Z"`
	UpdatedAt time.Time `json:"updated_at" example:"2026-03-08T14:30:00Z"`
}
