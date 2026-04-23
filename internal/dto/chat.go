package dto

import (
	"time"
)

type AddUsersToChatRequest struct {
	UserIDs []uint `json:"user_ids" binding:"required"`
}

type AddUserToChatRequest struct {
	UserID uint `json:"user_id" binding:"required" example:"12"`
}

type UpdateChatRequest struct {
	ChatName string `json:"chat_name" binding:"required" example:"hello there"`
}

type ChatDTO struct {
	ID        uint      `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	ChatName  string    `json:"chat_name"`
	IsGroup   bool      `json:"is_group"`
}

type GetUserChatsResponse struct {
	UserID uint      `json:"user_id" example:"12"`
	Chats  []ChatDTO `json:"chats"`
}

type UserDTO struct {
	ID uint `json:"id" example:"12"`
}

type GetChatMembersResponse struct {
	ChatID  uint      `json:"chat_id" example:"12"`
	Members []UserDTO `json:"members"`
}

/*
type CreateChatRequest struct {
	UserID uint `json:"user_id" binding:"required" example:"12"`
}
*/

type CreateChatRequest struct {
	Type     string `json:"type" binding:"required,oneof=private group"`
	ChatName string `json:"chat_name"`
	UserIDs  []uint `json:"users" binding:"required"`
}

type CreateChatResponse struct {
	Chat    ChatDTO `json:"chat_id"`
	Created bool    `json:"created" example:"true"`
}

type CreateGroupChatRequest struct {
	Users []UserDTO `json:"user_id" binding:"required"`
}

type CreateGroupChatResponse struct {
	Chat ChatDTO `json:"chat_id"`
}
