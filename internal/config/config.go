package config

import (
	"log"
	"os"
	"strconv"
)

type Config struct {
	TelegramApiId   int
	TelegramApiHash string
	TelegramSession string
	HttpPort        int
}

func LoadConfig() Config {
	config := InitConfig()
	CheckConfigFields(config)
	return config
}

func CheckConfigFields(config Config) {
	if config.TelegramApiId == 0 ||
		config.TelegramApiHash == "" ||
		config.TelegramSession == "" {
		log.Fatalf("One or more environment variables are missing.")
	}
}

func InitConfig() Config {
	telegramApiId, _ := strconv.Atoi(os.Getenv("TELEGRAM_API_ID"))
	port, _ := strconv.Atoi(os.Getenv("HTTP_PORT"))
	if port == 0 {
		port = 8080
	}
	config := Config{
		TelegramApiId:   telegramApiId,
		TelegramApiHash: os.Getenv("TELEGRAM_API_HASH"),
		TelegramSession: os.Getenv("TELEGRAM_SESSION"),
		HttpPort:        port,
	}

	return config
}
