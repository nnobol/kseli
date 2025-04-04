package config

import (
	"log"
	"os"
)

var (
	SecretKey string
	APIKey    string
)

func LoadConfig() {
	SecretKey = os.Getenv("SECRET_KEY")
	APIKey = os.Getenv("API_KEY")

	if SecretKey == "" || APIKey == "" {
		log.Fatal("Missing mandatory environment variables (SECRET_KEY, API_KEY)")
	}
}
