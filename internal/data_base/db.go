package database

import (
	"database/sql"
	"fmt"
	"log"
	"strings"

	_ "github.com/lib/pq"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"message_service/internal/models"
	"message_service/pkg/config"
)

var GormDB *gorm.DB

func createSystemDSN(dsn string) string {
	parts := strings.Split(dsn, " ")
	var newParts []string
	for _, part := range parts {
		if !strings.HasPrefix(part, "dbname=") {
			newParts = append(newParts, part)
		}
	}
	return strings.Join(newParts, " ") + " dbname=postgres"
}

func CreateDatabaseIfNotExists(targetDB string, dsn string) error {
	sysDSN := createSystemDSN(dsn)

	db, err := sql.Open("postgres", sysDSN)
	if err != nil {
		return fmt.Errorf("failed to connect to system database: %w", err)
	}
	defer db.Close()

	var exists bool
	err = db.QueryRow(`
		SELECT EXISTS (
			SELECT 1 FROM pg_database WHERE datname = $1
		)
	`, targetDB).Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to check database existence: %w", err)
	}

	if !exists {
		_, err = db.Exec(fmt.Sprintf("CREATE DATABASE %s", targetDB))
		if err != nil {
			return fmt.Errorf("failed to create database %s: %w", targetDB, err)
		}
		log.Printf("Database %s created successfully", targetDB)
	} else {
		log.Printf("Database %s already exists", targetDB)
	}

	return nil
}

func extractDBName(dsn string) string {
	for _, part := range strings.Split(dsn, " ") {
		if strings.HasPrefix(part, "dbname=") {
			return strings.TrimPrefix(part, "dbname=")
		}
	}
	return ""
}

func InitDB() error {
	config.GetDBString()
	dsn := config.Cnfg.DBurl
	dbName := extractDBName(dsn)
	if dbName == "" {
		dbName = "messageDB"
		log.Printf("dbname not found in DSN, using default %s", dbName)
	}

	if err := CreateDatabaseIfNotExists(dbName, dsn); err != nil {
		return fmt.Errorf("failed to ensure database exists: %w", err)
	}

	var err error
	GormDB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	sqlDB, err := GormDB.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}
	if err := sqlDB.Ping(); err != nil {
		return fmt.Errorf("unable to ping database: %w", err)
	}
	log.Println("Successfully connected to the database")

	err = GormDB.AutoMigrate(
		&models.User{},
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