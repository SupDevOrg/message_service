package websocket

import (
	"encoding/json"

	"github.com/gorilla/websocket"
)

type Client struct {
	Hub    *Hub
	Socket *websocket.Conn
	Recive chan []byte
	UserID uint
}

type IncomingMessage struct {
	ChatID  uint   `json:"chat_id"`
	Content string `json:"content"`
}

func (c *Client) Read() {
	defer func() {
		c.Hub.Unregister <- c
		c.Socket.Close()
	}()

	for {
		_, msg, err := c.Socket.ReadMessage()
		if err != nil {
			break
		}

		var incoming IncomingMessage
		if err := json.Unmarshal(msg, &incoming); err != nil {
			continue
		}

		c.Hub.Broadcast <- &Message{
			ChatID:   incoming.ChatID,
			SenderID: c.UserID,
			Content:  []byte(incoming.Content),
		}
	}
}

func (c *Client) Write() {
	defer c.Socket.Close()

	for msg := range c.Recive {
		err := c.Socket.WriteMessage(websocket.TextMessage, msg)
		if err != nil {
			break
		}
	}
}
