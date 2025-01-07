package config

import (
	"log"
	"os"
)

type Config struct {
	SecretKey string
	APIKey    string
}

var GlobalConfig *Config

func LoadConfig() {
	secretKey := os.Getenv("SECRET_KEY")
	apiKey := os.Getenv("API_KEY")

	if secretKey == "" || apiKey == "" {
		log.Fatal("Missing mandatory environment variables (SECRET_KEY, API_KEY)")
	}

	GlobalConfig = &Config{
		SecretKey: secretKey,
		APIKey:    apiKey,
	}
}
