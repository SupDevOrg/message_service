package models

import "time"

type ChatMember struct {
	ID       uint      `gorm:"primaryKey" json:"id"`
	ChatID   uint      `gorm:"not null;index" json:"chat_id"`
	UserID   uint      `gorm:"not null;index" json:"user_id"`
	JoinedAt time.Time `json:"joined_at"`

	Chat Chat `gorm:"foreignKey:ChatID" json:"-"`
}
