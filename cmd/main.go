package main

import (
	"log"
	database "message_service/internal/data_base"
	"message_service/internal/handlers"
	"message_service/internal/repositories"
	"message_service/internal/services"
	"message_service/internal/websocket"

	"github.com/gin-gonic/gin"
)

func main() {
	if err := database.InitDB(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	messageRepo := repositories.NewMessageRepository(database.GormDB)
	chatRepo := repositories.NewChatRepository(database.GormDB)
	chatMemberRepo := repositories.NewChatMemberRepository(database.GormDB)

	messageService := services.NewMessageService(messageRepo, chatRepo, chatMemberRepo)
	chatService := services.NewChatService(chatRepo, chatMemberRepo)
	chatMemberService := services.NewChatMemberService(chatRepo, chatMemberRepo)

	hub := websocket.NewHub(messageService)
	go hub.Run()

	messageHandler := handlers.NewMessageHandler(messageService)
	chatHandler := handlers.NewChatHandler(chatMemberService)
	wsHandler := handlers.NewWebSocketHandler(hub, chatService)

	router := gin.Default()

	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "OK"})
	})

	api := router.Group("/api/v1/message")
	{
		api.GET("/messages/:chat_id", messageHandler.GetMessages)
		api.POST("/members", chatHandler.AddUserToChat)
		api.GET("/ws", wsHandler.HandleWebSocket)
	}

	port := ":8080"
	log.Printf("Server started on %s", port)
	if err := router.Run(port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
