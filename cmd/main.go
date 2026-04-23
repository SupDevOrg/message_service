package main

import (
	"context"
	"log"
	database "message_service/internal/data_base"
	"message_service/internal/handlers"
	"message_service/internal/kafka"
	"message_service/internal/notification"
	"message_service/internal/repositories"
	"message_service/internal/services"
	"message_service/internal/websocket"
	"message_service/pkg/config"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "message_service/docs"
)

// @title Message Service API
// @version 1.0
// @description API для чатов, сообщений и WebSocket-подключения
// @BasePath /api/v1/message
// @schemes http h

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name X-Auth-User-ID
func main() {
	config.GetDBString()
	if err := database.InitDB(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	messageRepo := repositories.NewMessageRepository(database.GormDB)
	chatRepo := repositories.NewChatRepository(database.GormDB)
	chatMemberRepo := repositories.NewChatMemberRepository(database.GormDB)
	userRepo := repositories.NewUserRepo(database.GormDB)

	messageService := services.NewMessageService(messageRepo, chatRepo, chatMemberRepo)
	chatService := services.NewChatService(chatRepo, chatMemberRepo)
	chatMemberService := services.NewChatMemberService(chatRepo, chatMemberRepo)
	userService := services.NewUserService(userRepo)

	config.LoadNotificationConfig()
	notificationClient, err := notification.NewClient(
		config.Cnfg.NotificationGRPCAddr,
		config.Cnfg.NotificationGRPCTimeout,
	)
	if err != nil {
		log.Printf("failed to initialize notification grpc client: %v", err)
		notificationClient, _ = notification.NewClient("", 0)
	} else if config.Cnfg.NotificationGRPCAddr == "" {
		log.Println("notification grpc client disabled: NOTIFICATION_GRPC_ADDR is empty")
	}
	defer func() {
		if err := notificationClient.Close(); err != nil {
			log.Printf("error closing notification grpc client: %v", err)
		}
	}()

	hub := websocket.NewHub(messageService, chatMemberService, userService, notificationClient)
	go hub.Run()

	config.LoadKafkaConfig()
	log.Printf("kafka config loaded: Brokers=%s, Topic=%s, GroupID=%s",
		config.Cnfg.KafkaBrokers, config.Cnfg.KafkaTopic, config.Cnfg.KafkaGroupID)

	brokers := strings.Split(config.Cnfg.KafkaBrokers, ",")
	kafkaConsumer := kafka.NewConsumer(
		brokers,
		config.Cnfg.KafkaTopic,
		config.Cnfg.KafkaGroupID,
		userService,
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go kafkaConsumer.Start(ctx)

	messageHandler := handlers.NewMessageHandler(messageService)
	chatHandler := handlers.NewChatHandler(chatMemberService, chatService)
	wsHandler := handlers.NewWebSocketHandler(hub, chatService)

	router := gin.Default()

	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "OK"})
	})

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	api := router.Group("/api/v1/message")
	{
		// messages
		api.PUT("/messages/:message_id", messageHandler.ChangeMessage)
		api.GET("/messages/:message_id", messageHandler.GetMessage)
		api.DELETE("/messages/:message_id", messageHandler.DeleteMessage)
		// chat
		api.GET("/chats", chatHandler.GetUserChats)
		api.POST("/chats", chatHandler.CreateChat)
		api.POST("/chats/group", chatHandler.CreateGroupChat)
		api.GET("/chats/:chat_id", chatHandler.GetUserChats)
		api.GET("/chats/:chat_id/members", chatHandler.GetChatMembers)
		api.POST("/chats/:chat_id/members", chatHandler.AddUsersToChat)
		//DELETE /chats/:chat_id/members/:user_id
		//api.GET("/chats/:chat_id/messages", messageHandler.GetMessages)
		//api.POST("/chats/:chat_id/messages", messageHandler.CreateMessages)
		// websocket
		api.GET("/ws", wsHandler.HandleWebSocket)
	}

	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan

		log.Println("shutting down message service...")
		cancel()
		if err := kafkaConsumer.Close(); err != nil {
			log.Printf("error closing Kafka consumer: %v", err)
		}
		log.Println("message service stopped")
		os.Exit(0)
	}()

	port := ":8080"
	log.Printf("Server started on %s", port)
	if err := router.Run(port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
