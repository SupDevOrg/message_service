package repositories

import (
	"message_service/internal/models"

	"gorm.io/gorm"
)

type ChatMemberRepository struct {
	db *gorm.DB
}

func NewChatMemberRepository(db *gorm.DB) *ChatMemberRepository {
	return &ChatMemberRepository{db: db}
}

func (r *ChatMemberRepository) AddMember(chat, user uint) (*models.ChatMember, error) {
	member := &models.ChatMember{
		ChatID: chat,
		UserID: user,
	}

	err := r.db.Create(member).Error

	if err != nil {
		return nil, err
	}
	return member, nil
}

func (r *ChatMemberRepository) RemoveMember(chat, user uint) error {
	return r.db.Where("chat_id = ? AND user_id = ?", chat, user).
		Delete(&models.ChatMember{}).Error
}

func (r *ChatMemberRepository) GetUserChats(user uint) ([]uint, error) {
	var chatMembers []uint

	err := r.db.Where("user_id = ?", user).Find(&chatMembers).Error

	if err != nil {
		return nil, err
	}
	return chatMembers, nil
}

func (r *ChatMemberRepository) GetChatMembers(chat uint) ([]uint, error) {
	var users []uint

	err := r.db.Model(&models.ChatMember{}).
		Where("chat_id = ?", chat).
		Pluck("user_id", &users).Error

	if err != nil {
		return nil, err
	}
	return users, nil
}

func (r *ChatMemberRepository) CountChatMembers(chat uint) (int64, error) {
	var count int64
	err := r.db.Model(&models.ChatMember{}).
		Where("chat_id = ?", chat).
		Count(&count).Error
	return count, err
}

func (r *ChatMemberRepository) IsUserInChat(chat, user uint) (bool, error) {
	var member models.ChatMember

	err := r.db.Where("chat_id = ? AND user_id = ?", chat, user).First(&member).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

func (r *ChatMemberRepository) FindTwoUsersChat(user1, user2 uint) (uint, error) {
	var chat uint

	havingClause := `
		COUNT(DISTINCT user_id) = ? 
		AND SUM(CASE WHEN user_id = ? THEN 1 ELSE 0 END) = 1 
		AND SUM(CASE WHEN user_id = ? THEN 1 ELSE 0 END) = 1
	`
	err := r.db.Model(&models.ChatMember{}).
		Select("chat_id").
		Group("chat_id").
		Having(havingClause, 2, user1, user2).
		Limit(1).
		Scan(&chat).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return 0, nil
		}
		return 0, err
	}

	return chat, nil
}
