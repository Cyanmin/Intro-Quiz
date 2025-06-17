package config

import (
	"log"

	"github.com/joho/godotenv"
)

// LoadEnv loads environment variables from a .env file.
func LoadEnv() {
	if err := godotenv.Load(); err != nil {
		log.Println("no .env file found")
	}
}
