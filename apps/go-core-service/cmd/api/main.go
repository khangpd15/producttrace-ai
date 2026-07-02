package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"

	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/app"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/database"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/events/publisher"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/events/rabbitmq"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/events/types"
)

type HealthResponse struct {
	Service string `json:"service"`
	Status  string `json:"status"`
}

func main() {
	// 0. Load .env file (chỉ có tác dụng khi chạy local; production dùng biến môi trường thật)
	if err := godotenv.Load(); err != nil {
		log.Println("[WARN] .env file not found, using system environment variables")
	}

	// 1. Connect to PostgreSQL (GORM client)
	databasePostgres := database.ConnectPostgres()
	log.Println("PostgreSQL GORM connected successfully")

	// 2. Connect to PostgreSQL (Raw SQL)
	databasePostgresSQL, err := database.ConnectPostgresSQL()
	if err != nil {
		log.Fatalf("database connection failed: %v", err)
	}
	log.Println("PostgreSQL raw connected successfully")

	// 3. Connect to Redis client
	redisClient, err := database.ConnectRedis()
	if err != nil {
		log.Fatalf("redis connection failed: %v", err)
	}
	log.Println("Redis connected successfully")

	// 4. Setup RabbitMQ Manager
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

	// 5. Initialize Publisher
	pub := publisher.New(mgr)

	// 6. Bootstrap application
	appli := app.NewApp(databasePostgres, redisClient, pub, databasePostgresSQL)

	// Add health check
	appli.Router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, HealthResponse{
			Service: "go-core-service",
			Status:  "ok",
		})
	})

	// Add test event route
	appli.Router.GET("/test-event", func(c *gin.Context) {
		event := types.Event{
			EventID:       uuid.NewString(),
			EventType:     rabbitmq.ProductCreatedRK,
			EventVersion:  "1.0",
			Timestamp:     time.Now().UTC(),
			Producer:      "go-core-service",
			CorrelationID: uuid.NewString(),
			Payload: map[string]any{
				"productId": "p-001",
				"name":      "Coffee Bean",
			},
		}

		if err := pub.Publish(event); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "event published"})
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	defer mgr.Close()
	log.Println("RabbitMQ manager initialized successfully")

	log.Printf("Go service is running on port %s\n", port)
	if err := appli.Router.Run(":" + port); err != nil {
		log.Fatalf("failed to start server: %v \n", err)
	}

}
