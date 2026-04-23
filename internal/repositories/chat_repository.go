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

func (r *ChatRepository) UpdateChatName(chatID uint, chatName string) (*models.Chat, error) {
	var chat models.Chat

	err := r.db.First(&chat, chatID).Error
	if err != nil {
		return nil, err
	}

	chat.ChatName = chatName
	if err := r.db.Save(&chat).Error; err != nil {
		return nil, err
	}

	return &chat, nil
}

func (r *ChatRepository) Create() (*models.Chat, error) {
	chat := &models.Chat{}
	err := r.db.Create(chat).Error
	if err != nil {
		return nil, err
	}
	return chat, nil
}

func (r *ChatRepository) CreateGroup() (*models.Chat, error) {
	chat := &models.Chat{
		IsGroup: true,
	}

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

func (r *ChatRepository) FindByChatName(ChatName string) (*models.Chat, error) {
	var chat models.Chat
	err := r.db.Where("ChatName = ?", ChatName).First(&chat).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &chat, nil
}

func (r *ChatRepository) Delete(chat uint) error {
	return r.db.Delete(&models.Chat{}, chat).Error
}
