package services

import (
	"errors"
	"message_service/internal/models"
	"message_service/internal/repositories"

	"gorm.io/gorm"
)

type ChatService struct {
	chatRepo      *repositories.ChatRepository
	chatMembeRepo *repositories.ChatMemberRepository
}

func NewChatService(chatRepo *repositories.ChatRepository, chatMembeRepo *repositories.ChatMemberRepository) *ChatService {
	return &ChatService{
		chatRepo:      chatRepo,
		chatMembeRepo: chatMembeRepo,
	}
}

func (s *ChatService) GetChat(chat uint) (*models.Chat, error) {
	if chat == 0 {
		return nil, errors.New("invalid chat ID")
	}

	return s.chatRepo.FindByID(chat)
}

func (s *ChatService) UpdateChat(chatID uint, userID uint, chatName string) (*models.Chat, error) {
	if chatID == 0 {
		return nil, errors.New("invalid chat ID")
	}
	if userID == 0 {
		return nil, errors.New("invalid user ID")
	}
	if chatName == "" {
		return nil, errors.New("chat name cannot be empty")
	}

	_, err := s.chatRepo.FindByID(chatID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("chat not found")
		}
		return nil, err
	}

	isMember, err := s.chatMembeRepo.IsUserInChat(chatID, userID)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, errors.New("user is not a member of this chat")
	}

	updatedChat, err := s.chatRepo.UpdateChatName(chatID, chatName)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("chat not found")
		}
		return nil, err
	}

	return updatedChat, nil
}

func (s *ChatService) CreateChat(user1, user2 uint) (*models.Chat, bool, error) {
	if user1 == 0 || user2 == 0 {
		return nil, false, errors.New("invalid user IDs")
	}

	if user1 == user2 {
		return nil, false, errors.New("cannot create private chat with yourself")
	}

	exst, err := s.chatMembeRepo.FindTwoUsersChat(user1, user2)
	if err != nil {
		return nil, false, err
	}

	if exst != 0 {
		chat, err := s.chatRepo.FindByID(exst)
		if err != nil {
			return nil, false, err
		}
		return chat, false, nil
	}

	chat, err := s.chatRepo.Create()
	if err != nil {
		return nil, false, err
	}

	_, err = s.chatMembeRepo.AddMember(chat.ID, user1)
	if err != nil {
		_ = s.chatRepo.Delete(chat.ID)
		return nil, false, err
	}

	_, err = s.chatMembeRepo.AddMember(chat.ID, user2)
	if err != nil {
		_ = s.chatMembeRepo.RemoveMember(chat.ID, user1)
		_ = s.chatRepo.Delete(chat.ID)
		return nil, false, err
	}

	return chat, true, nil
}

func (s *ChatService) CreateGroup(userID uint) (*models.Chat, bool, error) {
	if userID == 0 {
		return nil, false, errors.New("invalid owner ID")
	}

	chat, err := s.chatRepo.CreateGroup()
	if err != nil {
		return nil, false, err
	}

	_, err = s.chatMembeRepo.AddMember(chat.ID, userID)
	if err != nil {
		return nil, false, err
	}

	return chat, true, nil
}

func (s *ChatService) DeleteChat(chat uint, user uint) error {
	if chat == 0 {
		return errors.New("invalid chat ID")
	}

	isMember, err := s.chatMembeRepo.IsUserInChat(chat, user)
	if err != nil {
		return err
	}
	if !isMember {
		return errors.New("user is not a member of this chat")
	}

	return s.chatRepo.Delete(chat)
}
