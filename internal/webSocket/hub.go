package websocket

import (
	"message_service/internal/services"
)

type Message struct {
	ChatID   uint
	SenderID uint
	Content  []byte
}

type Hub struct {
	Clients        map[uint]map[*Client]bool
	Broadcast      chan *Message
	Register       chan *Client
	Unregister     chan *Client
	MessageService *services.MessageService
}

func NewHub(messageService *services.MessageService) *Hub {
	return &Hub{
		Clients:        make(map[uint]map[*Client]bool),
		Broadcast:      make(chan *Message),
		Register:       make(chan *Client),
		Unregister:     make(chan *Client),
		MessageService: messageService,
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			if h.Clients[client.ChatID] == nil {
				h.Clients[client.ChatID] = make(map[*Client]bool)
			}
			h.Clients[client.ChatID][client] = true

		case client := <-h.Unregister:
			if _, ok := h.Clients[client.ChatID]; ok {
				delete(h.Clients[client.ChatID], client)
				close(client.Recive)
			}

		case message := <-h.Broadcast:

			if h.MessageService != nil {
				_, err := h.MessageService.CreateMessage(message.ChatID, message.SenderID, string(message.Content))
				if err != nil {
					// Логируем ошибку, но продолжаем отправку
					continue
				}
			}

			for client := range h.Clients[message.ChatID] {
				client.Recive <- message.Content
			}
		}
	}
}
