package kafka

import (
	"context"
	"encoding/json"
	"github.com/segmentio/kafka-go"
	"log"
	"message_service/internal/services"
)

type Consumer struct {
	reader      *kafka.Reader
	userService *services.UserService
}

func NewConsumer(brokers []string, topic string, groupID string, userService *services.UserService) *Consumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: brokers,
		Topic:   topic,
		GroupID: groupID,
	})

	return &Consumer{
		reader:      reader,
		userService: userService,
	}
}

func (c *Consumer) Start(ctx context.Context) {
	log.Printf("Starting Kafka consumer for topic: %s", c.reader.Config().Topic)

	for {
		select {
		case <-ctx.Done():
			log.Println("kafka consumer stopped")
			return
		default:
			msg, err := c.reader.ReadMessage(ctx)
			if err != nil {
				log.Printf("error reading message: %v", err)
				continue
			}

			go c.processMessage(msg)
		}
	}
}

func (c *Consumer) processMessage(msg kafka.Message) {
	key := string(msg.Key)
	log.Printf("received %s, %d, %d", key, msg.Partition, msg.Offset)

	switch key {
	case "user.created":
		c.handleUserCreated(msg.Value)
	case "user.updated":
		c.handleUserUpdated(msg.Value)
	default:
		log.Printf("unknown message key: %s", key)
	}
}

func (c *Consumer) handleUserCreated(payload []byte) {
	var event UserCreatedEvent
	if err := json.Unmarshal(payload, &event); err != nil {
		log.Printf("error unmarshaling UserCreatedEvent: %v", err)
		return
	}

	if _, err := c.userService.Create(uint(event.UserID), event.Username); err != nil {
		log.Printf("error creating user: %v", err)
		return
	}

	log.Printf("user created: ID=%d, Username=%s", event.UserID, event.Username)
}

func (c *Consumer) handleUserUpdated(payload []byte) {
	var event UserUpdatedEvent
	if err := json.Unmarshal(payload, &event); err != nil {
		log.Printf("error unmarshaling UserUpdatedEvent: %v", err)
		return
	}

	if event.Field == "username" {
		if err := c.userService.UpdateUsername(uint(event.UserID), event.NewValue); err != nil {
			log.Printf("error updating username: %v", err)
			return
		}
		log.Printf("user updated: ID=%d, from %s to %s", event.UserID, event.OldValue, event.NewValue)
	}
}

func (c *Consumer) Close() error {
	log.Println("closing Kafka consumer")
	return c.reader.Close()
}
