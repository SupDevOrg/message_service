package kafka

import (
	"context"
	"encoding/json"
	"log"

	"github.com/segmentio/kafka-go"
)

type Consumer struct {
	reader *kafka.Reader
}

func NewConsumer(brokers []string, topic string, groupID string) *Consumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: brokers,
		Topic:   topic,
		GroupID: groupID,
	})

	return &Consumer{
		reader: reader,
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

	log.Printf("USER CREATED: ID=%d, Username=%s, Timestamp=%d",
		event.UserID, event.Username, event.Timestamp)

	//userService.Create(event.UserID, event.Username)
}

func (c *Consumer) handleUserUpdated(payload []byte) {
	var event UserUpdatedEvent
	if err := json.Unmarshal(payload, &event); err != nil {
		log.Printf("error unmarshaling UserUpdatedEvent: %v", err)
		return
	}

	log.Printf("USER UPDATED: ID=%d, Field=%s, OldValue=%s, NewValue=%s, Timestamp=%d",
		event.UserID, event.Field, event.OldValue, event.NewValue, event.Timestamp)

	//userService.UpdateUsername(event.UserID, event.NewValue)
}

func (c *Consumer) Close() error {
	log.Println("closing Kafka consumer")
	return c.reader.Close()
}
