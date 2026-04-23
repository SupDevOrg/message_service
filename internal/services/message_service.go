package services

import (
	"errors"
	"message_service/internal/dto"
	"message_service/internal/models"
	"message_service/internal/repositories"

	"gorm.io/gorm"
)

func mapMessageToDTO(m models.Message) dto.MessageResponse {
	return dto.MessageResponse{
		ID:        m.ID,
		ChatID:    m.ChatID,
		SenderID:  m.SenderID,
		Content:   m.Content,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}
}

func mapMessagesToDTO(messages []models.Message) dto.MessagesResponse {
	resp := dto.MessagesResponse{
		Messages: make([]dto.MessageResponse, len(messages)),
	}

	for i, m := range messages {
		resp.Messages[i] = mapMessageToDTO(m)
	}

	return resp
}

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

func (s *MessageService) GetMessages(r MessagesPaginationRequest, user uint) (*dto.MessagesResponse, error) {
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

	response := mapMessagesToDTO(messages)
	return &response, nil
}

func (s *MessageService) CreateMessage(chat, sender uint, content string) (*dto.MessageResponse, error) {
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
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("chat not found")
		}
		return nil, err
	}

	isMember, err := s.chatMemberRepo.IsUserInChat(chat, sender)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, errors.New("sender is not a member of this chat")
	}

	msg, err := s.messageRepo.Create(chat, sender, content)
	if err != nil {
		return nil, err
	}
	resp := mapMessageToDTO(*msg)
	return &resp, nil
}

func (s *MessageService) ChangeMessage(messageID, userID uint, сontent string) (*dto.MessageResponse, error) {
	if messageID == 0 {
		return nil, errors.New("invalid message id")
	}

	if сontent == "" {
		return nil, errors.New("content cannot be empty")
	}

	msg, err := s.messageRepo.GetByID(messageID)
	if err != nil {
		return nil, err
	}

	if msg.SenderID != userID {
		return nil, errors.New("user cannot edit this message")
	}

	updatedMsg, err := s.messageRepo.UpdateContent(messageID, сontent)
	if err != nil {
		return nil, err
	}

	resp := mapMessageToDTO(*updatedMsg)
	return &resp, nil
}

func (s *MessageService) GetMessage(messageID, userID uint) (*dto.MessageResponse, error) {
	if messageID == 0 {
		return nil, errors.New("invalid message ID")
	}
	if userID == 0 {
		return nil, errors.New("invalid user ID")
	}

	message, err := s.messageRepo.GetByID(messageID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("message not found")
		}
		return nil, err
	}

	isMember, err := s.chatMemberRepo.IsUserInChat(message.ChatID, userID)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, errors.New("user is not a member of this chat")
	}

	resp := mapMessageToDTO(*message)
	return &resp, nil
}

func (s *MessageService) DeleteMessage(msg, user uint) error {
	if msg == 0 {
		return errors.New("invalid message ID")
	}
	if user == 0 {
		return errors.New("invalid user ID")
	}

	message, err := s.messageRepo.GetByID(msg)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("message not found")
		}
		return err
	}

	isMember, err := s.chatMemberRepo.IsUserInChat(message.ChatID, user)
	if err != nil {
		return err
	}
	if !isMember {
		return errors.New("user is not a member of this chat")
	}

	if message.SenderID != user {
		return errors.New("user cannot delete this message")
	}

	return s.messageRepo.Delete(msg)
}
