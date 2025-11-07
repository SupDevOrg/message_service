package websocket

import (
	"github.com/gorilla/websocket"
)

type Client struct {
	Hub    *Hub
	Socket *websocket.Conn
	Recive chan []byte
	UserID uint
	ChatID uint
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

		c.Hub.Broadcast <- &Message{
			ChatID:   c.ChatID,
			SenderID: c.UserID,
			Content:  msg,
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
