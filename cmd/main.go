package main

import (
	"context"
	"github.com/gin-gonic/gin"
	"log"
	database "message_service/internal/data_base"
	"message_service/internal/handlers"
	"message_service/internal/kafka"
	"message_service/internal/repositories"
	"message_service/internal/services"
	"message_service/internal/websocket"
	"message_service/pkg/config"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

func main() {
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

	hub := websocket.NewHub(messageService)
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

	api := router.Group("/api/v1/message")
	{
		api.GET("/messages/:chat_id", messageHandler.GetMessages)
		api.POST("/chat/adduser", chatHandler.AddUserToChat)
		api.POST("/chat/user", chatHandler.GetUserChats)
		api.POST("/chat/bytwouser", chatHandler.GetChatByTwoUsers)
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
