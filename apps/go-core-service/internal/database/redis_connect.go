package database

import (
	"context"
	"time"

	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/config"
	"github.com/redis/go-redis/v9"
)

// ConnectRedis initializes and returns a Redis client based on environment variables,
// after performing a Ping healthcheck to confirm connectivity.
func ConnectRedis() (*redis.Client, error) {
	cfg := config.NewRedisConfig()
	client := redis.NewClient(&redis.Options{
		Addr:         cfg.Addr,
		Username:     cfg.Username,
		Password:     cfg.Password,
		DB:           cfg.DB,
		PoolSize:     20,
		MinIdleConns: 5,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		PoolTimeout:  4 * time.Second,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		client.Close()
		return nil, err
	}

	return client, nil
}
