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
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/events/publisher"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/events/rabbitmq"
	authService "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/authen/service"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/pkg/cache"
)

func otpWorker() {
	if err := godotenv.Load(); err != nil {
		log.Println("[WARN] .env file not found, using system environment variables")
	}

	redisClient, err := database.ConnectRedis()
	if err != nil {
		log.Fatalf("redis connection failed: %v", err)
	}
	defer redisClient.Close()

	rabbitURL := os.Getenv("RABBITMQ_URL")
	if rabbitURL == "" {
		rabbitURL = "amqp://guest:guest@localhost:5672/"
	}

	mgr, err := rabbitmq.NewManager(rabbitURL)
	if err != nil {
		log.Fatalf("rabbitmq manager initialization failed: %v", err)
	}
	defer mgr.Close()

	cons := consumer.NewConsumer(mgr)
	pub := publisher.New(mgr)
	redisCache := cache.NewRedisCache(redisClient)

	service := authService.NewAuthenService(nil, redisCache, pub, cons)
	if err := service.ConsumerOTPEvent(context.Background()); err != nil {
		log.Fatalf("failed to start OTP consumer: %v", err)
	}
	log.Println("[OTPWorker] Consumer started, waiting for events...")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("[OTPWorker] Shutdown signal received, stopping gracefully...")
}
