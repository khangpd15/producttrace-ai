package config

import "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/utils"

type DBConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

func NewDBConfig() *DBConfig {
	return &DBConfig{
		Host:     utils.GetEnv("POSTGRES_HOST", "localhost"),
		Port:     utils.GetEnv("POSTGRES_PORT", "5432"),
		User:     utils.GetEnv("POSTGRES_USER", "postgres"),
		Password: utils.GetEnv("POSTGRES_PASSWORD", "123456"),
		DBName:   utils.GetEnv("POSTGRES_DB", "product_trace_db"),
		SSLMode:  utils.GetEnv("POSTGRES_SSLMODE", "disable"),
	}
}
