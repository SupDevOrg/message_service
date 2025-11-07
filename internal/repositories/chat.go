package repositories

import (
	"message_service/internal/models"

	"gorm.io/gorm"
)

type ChatRepository struct {
	db *gorm.DB
}

func NewChatRepository(db *gorm.DB) *ChatRepository {
	return &ChatRepository{db: db}
}

func (r *ChatRepository) Create() (*models.Chat, error) {
	chat := &models.Chat{}
	err := r.db.Create(chat).Error
	if err != nil {
		return nil, err
	}
	return chat, nil
}

func (r *ChatRepository) FindByID(chatID uint) (*models.Chat, error) {
	var chat models.Chat
	err := r.db.First(&chat, chatID).Error
	if err != nil {
		return nil, err
	}
	return &chat, nil
}

func (r *ChatRepository) Delete(chat uint) error {
	return r.db.Delete(&models.Chat{}, chat).Error
}
