package main

import (
	"context"
	"errors"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"tg-bridge/internal/config"
	"tg-bridge/internal/tgclient"
	"tg-bridge/internal/tgsession"
	"time"
)

func main() {
	cfg := config.LoadConfig()

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Set up signal handling for Ctrl+C
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	sessionStorage, err := tgsession.CreateInMemorySessionStorage(cfg.TelegramSession)
	if err != nil {
		log.Fatalf("Failed to create Telegram session: %v", err)
	}
	client := tgclient.CreateTelegramClient(cfg.TelegramApiId, cfg.TelegramApiHash, sessionStorage)

	// Test connection first
	connectCtx, connectCancel := context.WithTimeout(ctx, 30*time.Second)
	err = tgclient.CheckTelegramSession(connectCtx, client, func() error {
		return errors.New("not authorized, session is not valid")
	})
	connectCancel()
	if err != nil {
		log.Fatalf("Failed to connect to Telegram: %v", err)
	}

	// Use WaitGroup to manage goroutines
	var wg sync.WaitGroup

	// Channel to signal when Telegram request is done
	telegramDone := make(chan struct{})

	// Start Prometheus metrics server
	wg.Add(1)
	go func() {
		defer wg.Done()
		// TODO: startPrometheusServer(ctx, prometheusPort)
	}()

	// Start Telegram API requests goroutine
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer close(telegramDone) // Signal completion
		// TODO: replace test API request with real logic to read messages from channels
		err := tgclient.ReadAuthenticatedUserInfo(ctx, client)
		if err != nil {
			log.Printf("Telegram request failed: %v", err)
		}
	}()

	// Start metrics collection goroutine
	wg.Add(1)
	go func() {
		defer wg.Done()
		// TODO: sendMetricsToPrometheus(ctx, 5*time.Second) // Update metrics every 5 seconds
	}()

	log.Println("Application started. Press Ctrl+C to force shutdown...")

	// Wait for Telegram request to complete
	select {
	case <-telegramDone:
		log.Println("✅ Telegram request completed, initiating shutdown...")
	case sig := <-sigChan:
		log.Printf("🛑 Received signal: %v, initiating immediate shutdown...", sig)
	}

	// Cancel context to signal all goroutines to stop
	cancel()

	// Wait for all goroutines to finish
	wg.Wait()
	log.Println("Application stopped")
}
