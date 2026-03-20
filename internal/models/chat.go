package models

import "time"

type Chat struct {
	ID        uint         `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time    `json:"created_at"`
	ChatName  string       `gorm:"not null" json:"chatname"`
	IsGroup   bool         `grom:"index" json:"isgroupchat"`
	Members   []ChatMember `gorm:"foreignKey:ChatID" json:"-"`
	Messages  []Message    `gorm:"foreignKey:ChatID" json:"-"`
}
