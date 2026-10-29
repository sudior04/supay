package platform

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	Name            string
	Address         string
	DatabaseURL     string
	MaxConnections  int32
	ShutdownTimeout time.Duration
}

func Env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
func Load(name string, database bool) (Config, error) {
	cfg := Config{Name: name, Address: Env("HTTP_ADDR", ":8080"), DatabaseURL: os.Getenv("DATABASE_URL"), ShutdownTimeout: 10 * time.Second}
	n, err := strconv.Atoi(Env("DB_MAX_CONNECTIONS", "10"))
	if err != nil || n < 1 || n > 100 {
		return cfg, fmt.Errorf("DB_MAX_CONNECTIONS must be between 1 and 100")
	}
	cfg.MaxConnections = int32(n)
	if database && cfg.DatabaseURL == "" {
		return cfg, fmt.Errorf("DATABASE_URL is required for %s", name)
	}
	return cfg, nil
}
