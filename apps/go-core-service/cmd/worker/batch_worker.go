package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/database"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/events/consumer"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/events/rabbitmq"
	batchRepo "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/batch/repositories"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/product_item/repositories"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/product_item/services"
	variantRepo "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/product_variant/repositories"
)

func main() {
	// Load .env file (chỉ có tác dụng khi chạy local; production dùng biến môi trường thật)
	if err := godotenv.Load(); err != nil {
		log.Println("[WARN] .env file not found, using system environment variables")
	}

	// Connect to PostgreSQL (GORM client)
	databasePostgres := database.ConnectPostgres()
	log.Println("PostgreSQL GORM connected successfully")

	databasePostgresSQL, err := database.ConnectPostgresSQL()
	if err != nil {
		log.Fatalf("database connection failed: %v", err)
	}
	log.Println("PostgreSQL SQL connected successfully")

	// Setup RabbitMQ Manager
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

	cons := consumer.NewConsumer(mgr)

	// InitializeRepositories
	repo := repositories.NewProductItemRepository(databasePostgres)
	batchRepo := batchRepo.NewBatchRepository(databasePostgresSQL)
	variantRepo := variantRepo.NewProductVariantRepository(databasePostgres)
	service := services.NewProductItemService(repo, batchRepo, variantRepo, cons)

	if err := service.ConsumeBatchEvent(context.Background()); err != nil {
		log.Fatalf("failed to start batch consumer: %v", err)
	}
	log.Println("[BatchWorker] Consumer started, waiting for events...")

	// Block main goroutine cho đến khi nhận SIGINT hoặc SIGTERM.
	// QUAN TRỌNG: ConsumeBatchEvent() chỉ khởi động goroutine và return ngay.
	// Nếu không block ở đây, main() sẽ thoát, defer mgr.Close() chạy và
	// cancel manager context → tất cả handler goroutine nhận "context canceled".
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("[BatchWorker] Shutdown signal received, stopping gracefully...")
}
