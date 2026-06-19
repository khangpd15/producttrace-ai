package config

import (
	"fmt"

	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/utils"
)

type RedisConfig struct {
	Addr     string
	Username string
	Password string
	DB       int
}

func NewRedisConfig() *RedisConfig {
	host := utils.GetEnv("REDIS_HOST", "")
	port := utils.GetEnv("REDIS_PORT", "6379")

	return &RedisConfig{
		Addr:     fmt.Sprintf("%s:%s", host, port),
		Username: utils.GetEnv("REDIS_USER", ""),
		Password: utils.GetEnv("REDIS_PASSWORD", ""),
		DB:       utils.GetEnvAsInt("REDIS_DB", 0),
	}
}
