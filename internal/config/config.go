package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	ApiId   string
	ApiHash string
}

func LoadConfig() Config {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	config := Config{
		ApiId:   os.Getenv("API_ID"),
		ApiHash: os.Getenv("API_HASH"),
	}

	if config.ApiId == "" || config.ApiHash == "" {
		log.Fatalf("One or more environment variables are missing. Check .env file.")
	}

	return config
}
