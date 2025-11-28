package models

import "time"

type User struct {
	ID        uint      `gorm:"primaryKey;autoIncrement:false" json:"id"`
	Username  string    `gorm:"unique;not null" json:"username"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
