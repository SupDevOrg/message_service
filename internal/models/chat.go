package models

import "time"

type Chat struct {
	ID        uint         `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time    `json:"created_at"`
	ChatName  string       `gorm:"not null" json:"chatname"`
	TypeChat  bool         `grom:"index" json:"typeofchat"`
	Members   []ChatMember `gorm:"foreignKey:ChatID" json:"-"`
	Messages  []Message    `gorm:"foreignKey:ChatID" json:"-"`
}
