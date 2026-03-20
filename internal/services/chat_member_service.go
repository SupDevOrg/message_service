package services

import (
	"errors"
	"message_service/internal/dto"
	"message_service/internal/models"
	"message_service/internal/repositories"

	"gorm.io/gorm"
)

type ChatMemberService struct {
	chatRepo       *repositories.ChatRepository
	chatMemberRepo *repositories.ChatMemberRepository
}

func NewChatMemberService(chatRepo *repositories.ChatRepository, chatMemberRepo *repositories.ChatMemberRepository) *ChatMemberService {
	return &ChatMemberService{
		chatRepo:       chatRepo,
		chatMemberRepo: chatMemberRepo,
	}
}

func (s *ChatMemberService) AddUserToChat(chat, adduser, chatmmbr uint) (*models.ChatMember, error) {
	return nil, nil
}

func (s *ChatMemberService) AddUsersToChat(chatID uint, usersToAdd []uint, chatmember uint) error {
	if chatID == 0 {
		return errors.New("invalid chat ID")
	}

	_, err := s.chatRepo.FindByID(chatID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("chat not found")
		}
		return err
	}

	isMember, err := s.chatMemberRepo.IsUserInChat(chatID, chatmember)
	if err != nil {
		return err
	}
	if !isMember {
		return errors.New("only chat members can add new users")
	}

	if len(usersToAdd) == 0 {
		return nil
	}

	uniqUsers := make([]uint, 0, len(usersToAdd))
	m := make(map[uint]struct{}, len(usersToAdd))

	for _, userID := range usersToAdd {
		if _, ok := m[userID]; ok {
			continue
		}
		m[userID], uniqUsers = struct{}{}, append(uniqUsers, userID)
	}

	if len(uniqUsers) == 0 {
		return nil
	}

	alredayInChat, err := s.chatMemberRepo.GetUsersInChat(chatID, uniqUsers)
	if err != nil {
		return err
	}

	members := make([]models.ChatMember, 0, len(uniqUsers))
	for _, userID := range uniqUsers {
		if _, ok := alredayInChat[userID]; ok {
			continue
		}

		members = append(members, models.ChatMember{
			ChatID: chatID,
			UserID: userID,
		})
	}

	if len(members) == 0 {
		return nil
	}

	return s.chatMemberRepo.AddMembers(members)
}

func (s *ChatMemberService) RemoveUserFromChat(chat, removeuser, chatuser uint) error {
	if removeuser == 0 {
		return errors.New("invalid chat ID")
	}
	if chatuser == 0 {
		return errors.New("invalid user ID")
	}

	_, err := s.chatRepo.FindByID(chat)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("chat not found")
		}
		return err
	}

	isRemoveMember, err := s.chatMemberRepo.IsUserInChat(chat, removeuser)
	if err != nil {
		return err
	}
	if !isRemoveMember {
		return errors.New("user is not a member of this chat")
	}

	isChatMember, err := s.chatMemberRepo.IsUserInChat(chat, chatuser)
	if err != nil {
		return err
	}
	if !isChatMember && chatuser != removeuser {
		return errors.New("only chat members can remove users")
	}

	return s.chatMemberRepo.RemoveMember(chat, removeuser)
}

func (s *ChatMemberService) GetChatMembers(chat uint, user uint) ([]uint, error) {
	if chat == 0 {
		return nil, errors.New("invalid chat ID")
	}

	isMember, err := s.chatMemberRepo.IsUserInChat(chat, user)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, errors.New("only chat members can view member list")
	}

	return s.chatMemberRepo.GetChatMembers(chat)
}

func (s *ChatMemberService) GetUserChats(user uint) ([]dto.ChatDTO, error) {
	if user == 0 {
		return nil, errors.New("invalid user ID")
	}

	chatIDs, err := s.chatMemberRepo.GetUserChats(user)
	if err != nil {
		return nil, err
	}

	modelChats, err := s.chatMemberRepo.GetChatsByIDs(chatIDs)
	if err != nil {
		return nil, err
	}

	var dtochats []dto.ChatDTO
	for _, chat := range modelChats {
		dtochats = append(dtochats, dto.ChatDTO{ID: chat.ID,
			CreatedAt: chat.CreatedAt,
			ChatName:  chat.ChatName,
			IsGroup:   chat.IsGroup})
	}
	return dtochats, nil
}

func (s *ChatMemberService) GetChatMemberCount(chat uint) (int64, error) {
	if chat == 0 {
		return 0, errors.New("invalid chat ID")
	}

	return s.chatMemberRepo.CountChatMembers(chat)
}
