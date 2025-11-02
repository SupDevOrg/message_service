package services

import (
	"errors"
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
	if chat == 0 {
		return nil, errors.New("invalid chat ID")
	}
	if adduser == 0 {
		return nil, errors.New("invalid user ID")
	}

	_, err := s.chatRepo.FindByID(chat)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("chat not found")
		}
		return nil, err
	}

	isMmbr, err := s.chatMemberRepo.IsUserInChat(chat, chatmmbr)
	if err != nil {
		return nil, err
	}
	if !isMmbr {
		return nil, errors.New("only chat members can add new users")
	}

	inChat, err := s.chatMemberRepo.IsUserInChat(chat, adduser)
	if err != nil {
		return nil, err
	}

	if inChat {
		return nil, errors.New("user is already a member of this chat")
	}

	member, err := s.chatMemberRepo.AddMember(chat, adduser)
	if err != nil {
		return nil, err
	}

	return member, nil
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

func (s *ChatMemberService) GetUserChats(user uint) ([]uint, error) {
	if user == 0 {
		return nil, errors.New("invalid user ID")
	}

	return s.chatMemberRepo.GetUserChats(user)
}

func (s *ChatMemberService) GetChatMemberCount(chat uint) (int64, error) {
	if chat == 0 {
		return 0, errors.New("invalid chat ID")
	}

	return s.chatMemberRepo.CountChatMembers(chat)
}
