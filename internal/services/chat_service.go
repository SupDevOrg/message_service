package services

import (
	"errors"
	"message_service/internal/dto"
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

// Единая точка входа под POST /chats
func (s *ChatService) CreateChat(ownerID uint, req dto.CreateChatRequest) (*models.Chat, bool, error) {
	if ownerID == 0 {
		return nil, false, errors.New("invalid owner ID")
	}

	switch req.Type {
	case "private":
		return s.createPrivateChat(ownerID, req.UserIDs)

	case "group":
		return s.createGroupChat(ownerID, req.ChatName, req.UserIDs)

	default:
		return nil, false, errors.New("invalid chat type")
	}
}

func (s *ChatService) createPrivateChat(ownerID uint, userIDs []uint) (*models.Chat, bool, error) {
	if len(userIDs) != 1 {
		return nil, false, errors.New("private chat must contain exactly one user")
	}

	targetUserID := userIDs[0]
	if targetUserID == 0 {
		return nil, false, errors.New("invalid user IDs")
	}

	if ownerID == targetUserID {
		return nil, false, errors.New("cannot create private chat with yourself")
	}

	exst, err := s.chatMembeRepo.FindTwoUsersChat(ownerID, targetUserID)
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

	_, err = s.chatMembeRepo.AddMember(chat.ID, ownerID)
	if err != nil {
		_ = s.chatRepo.Delete(chat.ID)
		return nil, false, err
	}

	_, err = s.chatMembeRepo.AddMember(chat.ID, targetUserID)
	if err != nil {
		_ = s.chatMembeRepo.RemoveMember(chat.ID, ownerID)
		_ = s.chatRepo.Delete(chat.ID)
		return nil, false, err
	}

	return chat, true, nil
}

func (s *ChatService) createGroupChat(ownerID uint, chatName string, userIDs []uint) (*models.Chat, bool, error) {
	if chatName == "" {
		return nil, false, errors.New("chat_name is required for group chat")
	}

	if len(userIDs) == 0 {
		return nil, false, errors.New("group chat must contain at least one user")
	}

	chat, err := s.chatRepo.CreateGroup()
	if err != nil {
		return nil, false, err
	}

	_, err = s.chatMembeRepo.AddMember(chat.ID, ownerID)
	if err != nil {
		_ = s.chatRepo.Delete(chat.ID)
		return nil, false, err
	}

	uniqUsers := make([]uint, 0, len(userIDs))
	seen := make(map[uint]struct{}, len(userIDs))

	for _, userID := range userIDs {
		if userID == 0 {
			_ = s.chatRepo.Delete(chat.ID)
			return nil, false, errors.New("invalid user IDs")
		}

		if userID == ownerID {
			continue
		}

		if _, ok := seen[userID]; ok {
			continue
		}

		seen[userID] = struct{}{}
		uniqUsers = append(uniqUsers, userID)
	}

	for _, userID := range uniqUsers {
		_, err = s.chatMembeRepo.AddMember(chat.ID, userID)
		if err != nil {
			_ = s.chatRepo.Delete(chat.ID)
			return nil, false, err
		}
	}

	updatedChat, err := s.chatRepo.UpdateChatName(chat.ID, chatName)
	if err != nil {
		_ = s.chatRepo.Delete(chat.ID)
		return nil, false, err
	}

	return updatedChat, true, nil
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
