package main

// import (
// 	"log"
// 	"os"

// 	"github.com/joho/godotenv"
// 	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/database"
// 	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/events/publisher"
// 	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/events/rabbitmq"
// )

// func main() {
// 	// 0. Load .env file (chỉ có tác dụng khi chạy local; production dùng biến môi trường thật)
// 	if err := godotenv.Load(); err != nil {
// 		log.Println("[WARN] .env file not found, using system environment variables")
// 	}

// 	// 1. Connect to PostgreSQL (GORM client)
// 	databasePostgres := database.ConnectPostgres()
// 	log.Println("PostgreSQL GORM connected successfully")

// 	// 2. Connect to PostgreSQL (Raw SQL)
// 	databasePostgresSQL, err := database.ConnectPostgresSQL()
// 	if err != nil {
// 		log.Fatalf("database connection failed: %v", err)
// 	}
// 	log.Println("PostgreSQL raw connected successfully")

// 	// 3. Connect to Redis client
// 	redisClient, err := database.ConnectRedis()
// 	if err != nil {
// 		log.Fatalf("redis connection failed: %v", err)
// 	}
// 	log.Println("Redis connected successfully")

// 	// 4. Setup RabbitMQ Manager
// 	rabbitURL := os.Getenv("RABBITMQ_URL")
// 	if rabbitURL == "" {
// 		rabbitURL = "amqp://guest:guest@localhost:5672/"
// 	}

// 	mgr, err := rabbitmq.NewManager(rabbitURL)
// 	if err != nil {
// 		log.Fatalf("rabbitmq manager initialization failed: %v", err)
// 	}
// 	defer mgr.Close()
// 	log.Println("RabbitMQ manager initialized successfully")

// 	// 5. Initialize Publisher
// 	pub := publisher.New(mgr)

// }
