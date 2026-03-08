package models

import "time"

type Message struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	ChatID    uint      `gorm:"not null;index" json:"chat_id"`
	SenderID  uint      `gorm:"not null" json:"sender_id"`
	Content   string    `gorm:"type:text;not null" json:"content"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	Chat Chat `gorm:"foreignKey:ChatID" json:"-"`
}
