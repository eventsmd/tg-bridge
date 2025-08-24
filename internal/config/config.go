package config

import (
	"log"
	"os"
	"strconv"
)

type Config struct {
	PostgresConnectionString string
	TelegramApiId            int
	TelegramApiHash          string
	TelegramSession          string
	TemporalHostPort         string
	TemporalNamespace        string
	TemporalTaskQueue        string
	TemporalWorkflowType     string
	HttpPort                 int
}

func LoadConfig() Config {
	config := InitConfig()
	CheckConfigFields(config)
	return config
}

func CheckConfigFields(config Config) {
	if config.PostgresConnectionString == "" ||
		config.TelegramApiId == 0 ||
		config.TelegramApiHash == "" ||
		config.TelegramSession == "" ||
		config.TemporalHostPort == "" ||
		config.TemporalNamespace == "" ||
		config.TemporalTaskQueue == "" ||
		config.TemporalWorkflowType == "" {
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
		PostgresConnectionString: os.Getenv("POSTGRES_CONNECTION_STRING"),
		TelegramApiId:            telegramApiId,
		TelegramApiHash:          os.Getenv("TELEGRAM_API_HASH"),
		TelegramSession:          os.Getenv("TELEGRAM_SESSION"),
		TemporalHostPort:         os.Getenv("TEMPORAL_HOST_PORT"),
		TemporalNamespace:        os.Getenv("TEMPORAL_NAMESPACE"),
		TemporalTaskQueue:        os.Getenv("TEMPORAL_TASK_QUEUE"),
		TemporalWorkflowType:     os.Getenv("TEMPORAL_WORKFLOW_TYPE"),
		HttpPort:                 port,
	}

	return config
}
