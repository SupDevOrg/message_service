package notification

import (
	"context"
	"strings"
	"time"

	"message_service/internal/grpc/notificationpb"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type MessageNotification struct {
	MessageID      uint64
	ChatID         uint64
	SenderID       uint64
	SenderUsername string
	Content        string
	RecipientIDs   []uint64
	CreatedAt      time.Time
}

type Client interface {
	SendMessage(MessageNotification) error
	Close() error
}

type grpcClient struct {
	client  notificationpb.NotificationServiceClient
	conn    *grpc.ClientConn
	timeout time.Duration
}

type noopClient struct{}

func NewClient(addr string, timeout time.Duration) (Client, error) {
	if strings.TrimSpace(addr) == "" {
		return noopClient{}, nil
	}

	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	return &grpcClient{
		client:  notificationpb.NewNotificationServiceClient(conn),
		conn:    conn,
		timeout: timeout,
	}, nil
}

func (c *grpcClient) SendMessage(notification MessageNotification) error {
	ctx := context.Background()
	if c.timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, c.timeout)
		defer cancel()
	}

	_, err := c.client.SendMessageNotification(ctx, &notificationpb.SendMessageNotificationRequest{
		MessageId:       notification.MessageID,
		ChatId:          notification.ChatID,
		SenderId:        notification.SenderID,
		SenderUsername:  notification.SenderUsername,
		Content:         notification.Content,
		RecipientIds:    notification.RecipientIDs,
		CreatedAtUnixMs: notification.CreatedAt.UnixMilli(),
	})
	return err
}

func (c *grpcClient) Close() error {
	return c.conn.Close()
}

func (noopClient) SendMessage(MessageNotification) error {
	return nil
}

func (noopClient) Close() error {
	return nil
}
