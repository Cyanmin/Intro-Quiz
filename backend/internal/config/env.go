package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// TimeLimit defines the countdown duration in seconds.
var TimeLimit = 10

// LoadEnv loads environment variables from a .env file.
func LoadEnv() {
	if err := godotenv.Load(); err != nil {
		log.Println("no .env file found")
	}
	if v := os.Getenv("TIME_LIMIT"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			TimeLimit = n
		}
	}
}
