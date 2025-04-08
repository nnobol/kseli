package config

import (
	"log"
	"os"
)

// Injected during the build for prod, taken from env variables for development
var (
	SecretKey string
	APIKey    string
)

func LoadConfig() {
	if SecretKey == "" {
		SecretKey = os.Getenv("SECRET_KEY")
	}

	if APIKey == "" {
		APIKey = os.Getenv("API_KEY")
	}

	if SecretKey == "" || APIKey == "" {
		log.Fatal("Missing mandatory environment variables (SECRET_KEY, API_KEY)")
	}
}
