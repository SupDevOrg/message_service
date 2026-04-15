package config

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	DBurl                   string
	KafkaBrokers            string
	KafkaTopic              string
	KafkaGroupID            string
	NotificationGRPCAddr    string
	NotificationGRPCTimeout time.Duration
}

var Cnfg Config

func GetDBString() {
	if err := godotenv.Load(); err != nil {
		log.Println("Error loading .env file")
	}

	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	dbname := os.Getenv("DB_NAME")
	sslmode := os.Getenv("DB_SSLMODE")

	if sslmode == "" {
		sslmode = "disable"
	}

	Cnfg.DBurl = fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		host, user, password, dbname, port, sslmode,
	)
}

func LoadKafkaConfig() {
	if err := godotenv.Load(); err != nil {
		log.Println("Error loading .env file")
	}

	Cnfg.KafkaBrokers = os.Getenv("KAFKA_BROKERS")
	if Cnfg.KafkaBrokers == "" {
		Cnfg.KafkaBrokers = "localhost:9092"
	}

	Cnfg.KafkaTopic = os.Getenv("KAFKA_TOPIC")
	if Cnfg.KafkaTopic == "" {
		Cnfg.KafkaTopic = "user-events"
	}

	Cnfg.KafkaGroupID = os.Getenv("KAFKA_GROUP_ID")
	if Cnfg.KafkaGroupID == "" {
		Cnfg.KafkaGroupID = "message-service-group"
	}
}

func LoadNotificationConfig() {
	if err := godotenv.Load(); err != nil {
		log.Println("Error loading .env file")
	}

	Cnfg.NotificationGRPCAddr = os.Getenv("NOTIFICATION_GRPC_ADDR")

	timeout := os.Getenv("NOTIFICATION_GRPC_TIMEOUT")
	if timeout == "" {
		Cnfg.NotificationGRPCTimeout = 500 * time.Millisecond
		return
	}

	parsedTimeout, err := time.ParseDuration(timeout)
	if err != nil {
		log.Printf("invalid NOTIFICATION_GRPC_TIMEOUT %q, using default 500ms", timeout)
		Cnfg.NotificationGRPCTimeout = 500 * time.Millisecond
		return
	}

	Cnfg.NotificationGRPCTimeout = parsedTimeout
}
