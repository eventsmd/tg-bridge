package config

import (
	"log"

	"github.com/joho/godotenv"
)

type Config struct {
}

func LoadConfig() Config {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	config := Config{}

	return config
}
