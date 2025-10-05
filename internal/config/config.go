package config

import (
	"log"
	"os"
	"strconv"
	"strings"
	"tg-bridge/internal/domain"
	"tg-bridge/internal/tgclient"
)

type Config struct {
	PostgresConnectionString string
	TelegramApiId            int
	TelegramApiHash          string
	TelegramChannels         map[domain.Supplier]string
	TelegramChannelsSession  map[domain.Supplier]tgclient.Channel
	TelegramFetchInterval    int
	TelegramPageSize         int
	TelegramSession          string
	TemporalHostPort         string
	TemporalNamespace        string
	TemporalTaskQueue        string
	TemporalWorkflowType     string
	HttpPort                 int
	MetricsPort              int
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
		config.TelegramChannels == nil ||
		len(config.TelegramChannels) == 0 ||
		config.TelegramFetchInterval == 0 ||
		config.TelegramPageSize == 0 ||
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

	metricsPort, _ := strconv.Atoi(os.Getenv("METRICS_PORT"))
	if metricsPort == 0 {
		metricsPort = 8081
	}

	telegramPageSize, _ := strconv.Atoi(os.Getenv("TELEGRAM_PAGE_SIZE"))
	if telegramPageSize == 0 {
		telegramPageSize = 25
	}

	telegramFetchInterval, _ := strconv.Atoi(os.Getenv("TELEGRAM_FETCH_INTERVAL"))
	if telegramFetchInterval == 0 {
		telegramFetchInterval = 60
	}
	config := Config{
		PostgresConnectionString: os.Getenv("POSTGRES_CONNECTION_STRING"),
		TelegramApiId:            telegramApiId,
		TelegramApiHash:          os.Getenv("TELEGRAM_API_HASH"),
		TelegramChannels:         parseChannel(os.Getenv("TELEGRAM_CHANNELS")),
		TelegramFetchInterval:    telegramFetchInterval,
		TelegramPageSize:         telegramPageSize,
		TelegramSession:          os.Getenv("TELEGRAM_SESSION"),
		TemporalHostPort:         os.Getenv("TEMPORAL_HOST_PORT"),
		TemporalNamespace:        os.Getenv("TEMPORAL_NAMESPACE"),
		TemporalTaskQueue:        os.Getenv("TEMPORAL_TASK_QUEUE"),
		TemporalWorkflowType:     os.Getenv("TEMPORAL_WORKFLOW_TYPE"),
		HttpPort:                 port,
		MetricsPort:              metricsPort,
	}

	return config
}

func parseChannel(channel string) map[domain.Supplier]string {
	result := make(map[domain.Supplier]string)
	if channel == "" {
		return result
	}

	parts := strings.Split(channel, ",")
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		kv := strings.SplitN(p, "=", 2)
		if len(kv) != 2 {
			continue
		}
		k := strings.TrimSpace(kv[0])
		v := strings.TrimSpace(kv[1])
		if k == "" || v == "" {
			continue
		}
		result[domain.Supplier{Type: k}] = v
	}
	return result
}
