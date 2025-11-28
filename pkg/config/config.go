package config

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DBurl        string
	KafkaBrokers string
	KafkaTopic   string
	KafkaGroupID string
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
