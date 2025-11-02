package services

import (
	"errors"
	"message_service/internal/models"
	"message_service/internal/repositories"

	"gorm.io/gorm"
)

type MessageService struct {
	messageRepo    *repositories.MessageRepository
	chatRepo       *repositories.ChatRepository
	chatMemberRepo *repositories.ChatMemberRepository
}

func NewMessageService(messageRepo *repositories.MessageRepository,
	chatRepo *repositories.ChatRepository,
	chatMemberRepo *repositories.ChatMemberRepository,
) *MessageService {
	return &MessageService{
		messageRepo:    messageRepo,
		chatRepo:       chatRepo,
		chatMemberRepo: chatMemberRepo,
	}
}

type MessagesPaginationRequest struct {
	Chat     uint
	PageNum  int
	PageSize int
}

func (s *MessageService) GetMessages(r MessagesPaginationRequest, user uint) (*[]models.Message, error) {
	if r.Chat == 0 {
		return nil, errors.New("invalid chat ID")
	}
	if r.PageNum < 1 {
		return nil, errors.New("page number must be greater than 0")
	}
	if r.PageSize < 1 {
		return nil, errors.New("page size must be greater than 0")
	}
	if r.PageSize > 50 {
		return nil, errors.New("page size must not exceed 100")
	}

	_, err := s.chatRepo.FindByID(r.Chat)
	if err != nil {
		return nil, err
	}

	isMmbr, err := s.chatMemberRepo.IsUserInChat(r.Chat, user)
	if err != nil {
		return nil, err
	}
	if !isMmbr {
		return nil, errors.New("user is not a member of this chat")
	}

	totalItems, err := s.messageRepo.CountMsgsInChat(r.Chat)
	if err != nil {
		return nil, err
	}

	totalPages := int(totalItems) / r.PageSize

	if int(totalItems)%r.PageSize != 0 {
		totalPages++
	}

	if r.PageNum > totalPages {
		return nil, errors.New("page number must be less than total pages")
	}
	offset := (r.PageNum - 1) * r.PageSize

	messages, err := s.messageRepo.GetMsgsWithOffset(r.Chat, r.PageSize, offset)
	if err != nil {
		return nil, err
	}
	return &messages, nil
}

func (s *MessageService) CreateMessage(chat, sender uint, content string) (*models.Message, error) {
	if chat == 0 {
		return nil, errors.New("invalid chat ID")
	}
	if sender == 0 {
		return nil, errors.New("invalid sender ID")
	}
	if content == "" {
		return nil, errors.New("message cannot be empty")
	}

	_, err := s.chatRepo.FindByID(chat)
	if err != nil {

		return nil, err
	}

	isMember, err := s.chatMemberRepo.IsUserInChat(chat, sender)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, errors.New("sender is not a member of this chat")
	}

	return s.messageRepo.Create(chat, sender, content)
}

/*
	func (s *MessageService) GetMessage(messageID, requestUserID uint) (*models.Message, error) {
		isMember, err := s.chatMemberRepo.IsUserInChat(message.ChatID, requestUserID)
		if err != nil {
			return nil, err
		}
		if !isMember {
			return nil, errors.New("user is not a member of this chat")
		}

		if messageID == 0 {
			return nil, errors.New("invalid message ID")
		}

		message, err := s.messageRepo.GetByID(messageID)
		if err != nil {
			return nil, err
		}

		return message, nil
	}
*/
func (s *MessageService) DeleteMessage(msg, user uint) error {
	if msg == 0 {
		return errors.New("invalid message ID")
	}

	_, err := s.messageRepo.GetByID(msg)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("message not found")
		}
		return err
	}

	return s.messageRepo.Delete(msg)
}
