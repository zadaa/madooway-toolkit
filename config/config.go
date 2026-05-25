package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port          string
	DBUser        string
	DBPassword    string
	DBName        string
	DBHost        string
	DBPort        string
	SessionSecret string
}

var AppConfig *Config

func LoadConfig() {
	// Ignore error since environment variables might be set via system env
	_ = godotenv.Load()

	AppConfig = &Config{
		Port:          getEnv("PORT", "8080"),
		DBUser:        getEnv("DB_USER", "root"),
		DBPassword:    getEnv("DB_PASSWORD", ""),
		DBName:        getEnv("DB_NAME", "task_manager"),
		DBHost:        getEnv("DB_HOST", "127.0.0.1"),
		DBPort:        getEnv("DB_PORT", "3306"),
		SessionSecret: getEnv("SESSION_SECRET", "default-secret-key-change-me"),
	}
	log.Println("Configuration loaded successfully")
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
