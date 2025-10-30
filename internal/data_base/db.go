package database

import (
	"fmt"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"message_service/internal/models"
	"message_service/pkg/config"
)

var GormDB *gorm.DB

func InitDB() error {
	config.GetDBString()

	log.Println(config.Cnfg.DBurl)
	var err error
	GormDB, err = gorm.Open(postgres.Open(config.Cnfg.DBurl), &gorm.Config{})

	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	DB, err := GormDB.DB()

	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	if err := DB.Ping(); err != nil {
		return fmt.Errorf("unable to ping database: %w", err)
	}

	log.Println("Successfully connected to the database")

	// Миграции для message_service
	err = GormDB.AutoMigrate(
		&models.Chat{},
		&models.ChatMember{},
		&models.Message{},
	)
	if err != nil {
		log.Fatalf("AutoMigrate failed: %v", err)
	}

	err = GormDB.Exec(`
		CREATE UNIQUE INDEX IF NOT EXISTS idx_chat_user_unique 
		ON chat_members(chat_id, user_id)
	`).Error
	if err != nil {
		log.Fatalf("Failed to create unique index: %v", err)
	}

	log.Println("Database migration completed successfully")
	return nil
}
