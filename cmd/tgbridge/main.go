package main

import (
	"context"
	"errors"
	"tg-bridge/internal/config"
	"tg-bridge/internal/tgclient"
	"tg-bridge/internal/tgsession"
)

func main() {
	cfg := config.LoadConfig()

	sessionStorage, err := tgsession.CreateInMemorySessionStorage(cfg.TelegramSession)
	client := tgclient.CreateTelegramClient(cfg.TelegramApiId, cfg.TelegramApiHash, sessionStorage)

	err = client.Run(context.Background(), func(ctx context.Context) error {
		onNotAuthorized := func() error {
			return errors.New("not authorized, session is not valid")
		}
		return tgclient.CheckTelegramSession(ctx, client, onNotAuthorized)
	})

	if err != nil {
		panic(err)
	}
}
