package main

import (
	"context"
	"log"
	"tg-bridge/internal/config"
	"tg-bridge/internal/tgsession"

	"github.com/gotd/td/session"
	"github.com/gotd/td/telegram"
)

func main() {
	cfg := config.LoadConfig()

	sessionStorage, err := tgsession.CreateSessionStorage(cfg)
	client := CreateTelegramClient(cfg, sessionStorage)

	err = client.Run(context.Background(), func(ctx context.Context) error {
		status, err := client.Auth().Status(ctx)
		if err != nil {
			return err
		}
		if !status.Authorized {
			log.Fatal("Not unauthorized, session is not working")
		}
		log.Println("Successfully authenticated")

		return nil
	})

	if err != nil {
		panic(err)
	}
}

func CreateTelegramClient(cfg config.Config, sessionStorage session.Storage) *telegram.Client {
	return telegram.NewClient(cfg.TelegramApiId, cfg.TelegramApiHash, telegram.Options{
		SessionStorage: sessionStorage,
	})
}
