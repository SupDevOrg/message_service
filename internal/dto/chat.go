package dto

import (
	"time"
)

type AddUserToChatRequest struct {
	UserID uint `json:"user_id" example:"12"`
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

type CreateChatRequest struct {
	UserID uint `json:"user_id" example:"12"`
}

type CreateChatResponse struct {
	Chat    ChatDTO `json:"chat_id"`
	Created bool    `json:"created" example:"true"`
}

type CreateGroupChatRequest struct {
	Users []UserDTO `json:"user_id"`
}

type CreateGroupChatResponse struct {
	Chat    ChatDTO `json:"chat_id"`
	IsGroup bool    `json:"is_group" example:"true"`
}
