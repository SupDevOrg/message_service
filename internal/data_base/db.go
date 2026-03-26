package database

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"message_service/internal/models"
	"message_service/pkg/config"
)

const defaultDBName = "messageDB"

var GormDB *gorm.DB

func InitDB() error {
	log.Println("InitDB START")
	dsn := config.Cnfg.DBurl
	if dsn == "" {
		return fmt.Errorf("database DSN is empty")
	}

	dbName := extractDBName(dsn)
	if dbName == "" {
		dbName = defaultDBName
		log.Printf("WARN: dbname not found in DSN, using default %s", dbName)
	}

	if err := waitForPostgres(dsn, 10, 2*time.Second); err != nil {
		return fmt.Errorf("postgres is not ready: %w", err)
	}

	if err := ensureDatabaseExists(dbName, dsn); err != nil {
		return fmt.Errorf("ensure database exists: %w", err)
	}

	if err := connect(dsn); err != nil {
		return err
	}
	if err := migrate(); err != nil {
		return err
	}

	log.Printf("database ready: %s", dbName)
	return nil
}

func waitForPostgres(dsn string, attempts int, delay time.Duration) error {
	sysDSN := buildSystemDSN(dsn)

	for i := 0; i < attempts; i++ {
		db, err := sql.Open("pgx", sysDSN)
		if err == nil {
			err = db.Ping()
			db.Close()
			if err == nil {
				log.Println("postgres is ready")
				return nil
			}
		}

		log.Printf("waiting for postgres... (%d/%d)", i+1, attempts)
		time.Sleep(delay)
	}

	return fmt.Errorf("postgres not available after %d attempts", attempts)
}

func Close() error {
	if GormDB == nil {
		return nil
	}
	sqlDB, err := GormDB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

func connect(dsn string) error {
	var err error
	GormDB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		return fmt.Errorf("open connection: %w", err)
	}

	sqlDB, err := GormDB.DB()
	if err != nil {
		return fmt.Errorf("get underlying sql.DB: %w", err)
	}

	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(5 * time.Minute)
	sqlDB.SetConnMaxIdleTime(1 * time.Minute)

	if err := sqlDB.Ping(); err != nil {
		return fmt.Errorf("ping database: %w", err)
	}

	log.Println("connected to database")
	return nil
}

func migrate() error {
	if err := GormDB.AutoMigrate(
		&models.User{},
		&models.Chat{},
		&models.ChatMember{},
		&models.Message{},
	); err != nil {
		return fmt.Errorf("auto migrate: %w", err)
	}

	const q = `
		CREATE UNIQUE INDEX IF NOT EXISTS idx_chat_user_unique
		ON chat_members(chat_id, user_id)
	`
	if err := GormDB.Exec(q).Error; err != nil {
		return fmt.Errorf("create unique index: %w", err)
	}

	log.Println("migrations completed")
	return nil
}

func ensureDatabaseExists(targetDB, dsn string) error {
	sysDSN := buildSystemDSN(dsn)

	log.Printf("connecting to system database: %s", sysDSN)

	db, err := sql.Open("pgx", sysDSN)
	if err != nil {
		return fmt.Errorf("connect to system database: %w", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		return fmt.Errorf("ping system database: %w", err)
	}

	var exists bool
	if err := db.QueryRow(
		`SELECT EXISTS (SELECT 1 FROM pg_database WHERE lower(datname) = lower($1))`,
		targetDB,
	).Scan(&exists); err != nil {
		return fmt.Errorf("check database existence: %w", err)
	}

	if exists {
		log.Printf("database already exists: %s", targetDB)
		return nil
	}

	if _, err := db.Exec(`CREATE DATABASE ` + quoteIdentifier(targetDB)); err != nil {
		return fmt.Errorf("create database %q: %w", targetDB, err)
	}

	log.Printf("database created: %s", targetDB)
	return nil
}

func extractDBName(dsn string) string {
	for _, part := range strings.Fields(dsn) {
		if strings.HasPrefix(part, "dbname=") {
			return strings.TrimPrefix(part, "dbname=")
		}
	}
	return ""
}

func buildSystemDSN(dsn string) string {
	parts := strings.Fields(dsn)
	result := make([]string, 0, len(parts))

	for _, p := range parts {
		if !strings.HasPrefix(p, "dbname=") {
			result = append(result, p)
		}
	}

	result = append(result, "dbname=postgres")
	return strings.Join(result, " ")
}

func quoteIdentifier(name string) string {
	return `"` + strings.ReplaceAll(name, `"`, `""`) + `"`
}