package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	TelegramApiId   int
	TelegramApiHash string
	TelegramSession string
}

func LoadConfig() Config {
	LoadEnvFile()
	config := InitConfig()
	CheckConfigFields(config)
	return config
}

func CheckConfigFields(config Config) {
	if config.TelegramApiId == 0 ||
		config.TelegramApiHash == "" ||
		config.TelegramSession == "" {
		log.Fatalf("One or more environment variables are missing. Check .env file.")
	}
}

func InitConfig() Config {
	telegramApiId, _ := strconv.Atoi(os.Getenv("TELEGRAM_API_ID"))
	config := Config{
		TelegramApiId:   telegramApiId,
		TelegramApiHash: os.Getenv("TELEGRAM_API_HASH"),
		TelegramSession: os.Getenv("TELEGRAM_SESSION"),
	}
	return config
}

func LoadEnvFile() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
}
