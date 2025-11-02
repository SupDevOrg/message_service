package models

import "time"

type User struct {
	Id        uint      `gorm:"primary key" json:"id"`
	Username  string    `gorm:"unique" json:"username"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
