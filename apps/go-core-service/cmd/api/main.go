package main

import (
	"log"
	"os"

	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/app"

	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/database"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/events/publisher"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/events/rabbitmq"
)

func main() {
	// 1. Connect to PostgreSQL (GORM client)
	databasePostgres, err := database.ConnectPostgres()
	if err != nil {
		log.Fatalf("database connection failed: %v", err)
	}
	log.Println("PostgreSQLGORM connected successfully")

	databasePostgresSQL, err := database.ConnectPostgresSQL()
	if err != nil {
		log.Fatalf("database connection failed: %v", err)
	}
	log.Println("PostgreSQL connected successfully")

	// 2. Connect to Redis client
	redisClient, err := database.ConnectRedis()
	if err != nil {
		log.Fatalf("redis connection failed: %v", err)
	}
	log.Println("Redis connected successfully")

	// 3. Setup RabbitMQ Manager
	rabbitURL := os.Getenv("RABBITMQ_URL")
	if rabbitURL == "" {
		rabbitURL = "amqp://guest:guest@localhost:5672/"
	}

	mgr, err := rabbitmq.NewManager(rabbitURL)
	if err != nil {
		log.Fatalf("rabbitmq manager initialization failed: %v", err)
	}
	defer mgr.Close()
	log.Println("RabbitMQ manager initialized successfully")

	// 4. Initialize Publisher
	pub := publisher.New(mgr)

	// 5. Bootstrap application
	appli := app.NewApp(databasePostgres, redisClient, pub, databasePostgresSQL)

	if err := appli.Router.Run(":8080"); err != nil {
		log.Fatalf("failed to start server: %v \n", err)
	}
}
