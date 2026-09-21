package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL      string
	StorageEndpoint  string
	StorageAccessKey string
	StorageSecretKey string
	StorageUseSSL    bool
	StorageBucket    string
	MaxMediaSize     int64
	ServerPort       string
}

func Load() Config {
	_ = godotenv.Load("../.env", ".env")

	return Config{
		DatabaseURL:      os.Getenv("DATABASE_URL"),
		StorageEndpoint:  os.Getenv("STORAGE_ENDPOINT"),
		StorageAccessKey: os.Getenv("STORAGE_ACCESS_KEY"),
		StorageSecretKey: os.Getenv("STORAGE_SECRET_KEY"),
		StorageUseSSL:    os.Getenv("STORAGE_USE_SSL") == "true",
		StorageBucket:    valueOrDefault("STORAGE_BUCKET", "session-media"),
		MaxMediaSize:     20 * 1024 * 1024,
		ServerPort:       valueOrDefault("SERVER_PORT", "8080"),
	}
}

func valueOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
