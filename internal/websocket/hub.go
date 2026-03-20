package websocket

import (
	"encoding/json"
	"log"

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
}

func NewHub(messageService *services.MessageService, chatMemberService *services.ChatMemberService) *Hub {
	return &Hub{
		Clients:           make(map[uint]*Client),
		Broadcast:         make(chan *Message),
		Register:          make(chan *Client),
		Unregister:        make(chan *Client),
		MessageService:    messageService,
		ChatMemberService: chatMemberService,
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
