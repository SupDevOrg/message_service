package repositories

import (
	"message_service/internal/models"

	"gorm.io/gorm"
)

type MessageRepository struct {
	db *gorm.DB
}

func NewMessageRepository(db *gorm.DB) *MessageRepository {
	return &MessageRepository{db: db}
}

func (r *MessageRepository) Create(chatID, senderID uint, content string) (*models.Message, error) {
	message := &models.Message{
		ChatID:   chatID,
		SenderID: senderID,
		Content:  content,
	}

	err := r.db.Create(message).Error
	if err != nil {
		return nil, err
	}

	return message, nil
}

func (r *MessageRepository) GetMsgsWithOffset(chatID uint, limit, offset int) ([]models.Message, error) {
	var messages []models.Message

	err := r.db.Where("chat_id = ?", chatID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&messages).Error

	if err != nil {
		return nil, err
	}

	return messages, nil
}

func (r *MessageRepository) GetByID(messageID uint) (*models.Message, error) {
	var message models.Message
	err := r.db.First(&message, messageID).Error

	if err != nil {
		return nil, err
	}

	return &message, nil
}

func (r *MessageRepository) Delete(messageID uint) error {
	return r.db.Delete(&models.Message{}, messageID).Error
}

func (r *MessageRepository) CountMsgsInChat(chatID uint) (int64, error) {
	var count int64

	err := r.db.Model(&models.Message{}).
		Where("chat_id = ?", chatID).
		Count(&count).Error

	return count, err
}
