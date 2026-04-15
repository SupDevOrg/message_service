package websocket

import (
	"encoding/json"
	"log"

	"message_service/internal/dto"
	"message_service/internal/notification"
	"message_service/internal/services"
)

type Message struct {
	ChatID   uint
	SenderID uint
	Content  []byte
}

type Hub struct {
	Clients           map[uint]*Client
	Broadcast         chan *Message
	Register          chan *Client
	Unregister        chan *Client
	MessageService    *services.MessageService
	ChatMemberService *services.ChatMemberService
	UserService       *services.UserService
	Notification      notification.Client
}

func NewHub(
	messageService *services.MessageService,
	chatMemberService *services.ChatMemberService,
	userService *services.UserService,
	notificationClient notification.Client,
) *Hub {
	return &Hub{
		Clients:           make(map[uint]*Client),
		Broadcast:         make(chan *Message),
		Register:          make(chan *Client),
		Unregister:        make(chan *Client),
		MessageService:    messageService,
		ChatMemberService: chatMemberService,
		UserService:       userService,
		Notification:      notificationClient,
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.Clients[client.UserID] = client

		case client := <-h.Unregister:
			if _, ok := h.Clients[client.UserID]; ok {
				delete(h.Clients, client.UserID)
				close(client.Recive)
			}

		case message := <-h.Broadcast:
			savedMsg, err := h.MessageService.CreateMessage(message.ChatID, message.SenderID, string(message.Content))
			if err != nil {
				log.Printf("Failed to save message: %v", err)
				continue
			}

			msgJSON, err := json.Marshal(savedMsg)
			if err != nil {
				log.Printf("Failed to marshal message to JSON: %v", err)
				continue
			}

			participants, err := h.ChatMemberService.GetChatMembers(message.ChatID, message.SenderID)
			if err != nil {
				continue
			}

			h.sendNotification(savedMsg, participants)

			for _, userID := range participants {
				if client, ok := h.Clients[userID]; ok {
					select {
					case client.Recive <- msgJSON:
					default:
					}
				}
			}
		}
	}
}

func (h *Hub) sendNotification(savedMsg *dto.MessageResponse, participants []uint) {
	recipientIDs := make([]uint64, 0, len(participants))
	for _, userID := range participants {
		if userID == savedMsg.SenderID {
			continue
		}
		recipientIDs = append(recipientIDs, uint64(userID))
	}

	if len(recipientIDs) == 0 {
		return
	}

	senderUsername := ""
	if h.UserService != nil {
		sender, err := h.UserService.GetByID(savedMsg.SenderID)
		if err != nil {
			log.Printf("Failed to load sender for notification: %v", err)
		} else {
			senderUsername = sender.Username
		}
	}

	payload := notification.MessageNotification{
		MessageID:      uint64(savedMsg.ID),
		ChatID:         uint64(savedMsg.ChatID),
		SenderID:       uint64(savedMsg.SenderID),
		SenderUsername: senderUsername,
		Content:        savedMsg.Content,
		RecipientIDs:   recipientIDs,
		CreatedAt:      savedMsg.CreatedAt,
	}

	go func() {
		if err := h.Notification.SendMessage(payload); err != nil {
			log.Printf("Failed to send message notification: %v", err)
		}
	}()
}
