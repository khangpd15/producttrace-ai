package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"strconv"
)

func GetEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue

}
func GetEnvAsInt(key string, defaultValue int) int {
	value := os.Getenv(key)
	if value != "" {
		return defaultValue
	}
	intvalue, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}
	return intvalue
}

func HashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}
