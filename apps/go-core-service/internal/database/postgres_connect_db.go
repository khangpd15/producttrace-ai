package database

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/config"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func ConnectPostgres() *gorm.DB {
	cfg := config.NewDBConfig()

	log.Printf(
		"Connecting PostgreSQL: host=%s port=%s db=%s user=%s",
		cfg.Host,
		cfg.Port,
		cfg.DBName,
		cfg.User,
	)

	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s TimeZone=Asia/Ho_Chi_Minh",
		cfg.Host,
		cfg.Port,
		cfg.User,
		cfg.Password,
		cfg.DBName,
		cfg.SSLMode,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Connect database error: %v", err)
	}

	log.Println("PostgreSQL connected successfully")

	return db
}

func ConnectPostgresSQL() (*sql.DB, error) {
	cfg := config.NewDBConfig()

	log.Printf(
		"Connecting PostgreSQL: host=%s port=%s db=%s user=%s",
		cfg.Host,
		cfg.Port,
		cfg.DBName,
		cfg.User,
	)

	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s TimeZone=Asia/Ho_Chi_Minh",
		cfg.Host,
		cfg.Port,
		cfg.User,
		cfg.Password,
		cfg.DBName,
		cfg.SSLMode,
	)

	return sql.Open("rpx", dsn)
}
