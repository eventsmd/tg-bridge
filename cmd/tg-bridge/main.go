package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"tg-bridge/internal/config"
	"tg-bridge/internal/domain"
	"tg-bridge/internal/healthserver"
	"tg-bridge/internal/metricsserver"
	"tg-bridge/internal/tgclient"
	"tg-bridge/internal/tgsession"
	"time"

	"tg-bridge/internal/persistence"
	"tg-bridge/internal/temporalpub"
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

	// DB connection for offsets
	db, err := persistence.NewDatabase(cfg.PostgresConnectionString)
	if err != nil {
		log.Fatalf("failed to connect postgres: %v", err)
	}
	defer db.Close()

	// Publisher for Temporal workflows
	publisher, err := temporalpub.NewPublisher(
		cfg, nil, nil,
	)
	if err != nil {
		log.Fatalf("failed to init temporal publisher: %v", err)
	}

	log.Println("Application started. Press Ctrl+C to force shutdown...")

	var wg sync.WaitGroup

	// Health/Readiness server
	hs := healthserver.New(fmt.Sprintf(":%d", cfg.HttpPort))
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := hs.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()
	log.Printf("Health/Ready server listening on %s", hs.Addr())

	// Prometheus metrics server
	ms := metricsserver.New(fmt.Sprintf(":%d", cfg.MetricsPort))
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := ms.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Metrics server error: %v", err)
		}
	}()
	log.Printf("Metrics server listening on %s/metrics", ms.Addr())

	// Channel to signal when Telegram request is done
	telegramDone := make(chan struct{})

	// Start Telegram API requests goroutine
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer close(telegramDone) // Signal completion

		err := client.Run(ctx, func(ctx context.Context) error {
			// Test connection first
			connectCtx, connectCancel := context.WithTimeout(ctx, 30*time.Second)
			err := tgclient.CheckTelegramSession(connectCtx, client, func() error {
				return errors.New("not authorized, session is not valid")
			})
			connectCancel()
			if err != nil {
				return fmt.Errorf("failed to connect to Telegram: %w", err)
			}

			// Mark service as ready once Telegram session is validated
			hs.SetReady(true)

			if err := tgclient.ReadAuthenticatedUserInfo(ctx, client); err != nil {
				log.Printf("Couldn't read user info: %v", err)
			}

			// Resolve configured channels
			cfg.TelegramChannelsSession = make(map[domain.Supplier]tgclient.Channel, len(cfg.TelegramChannels))
			for supplier, channelName := range cfg.TelegramChannels {
				channel, err := tgclient.NewChannel(ctx, client, channelName, supplier)
				if err != nil {
					return fmt.Errorf("failed to find channel: %w", err)
				}
				log.Printf("💬 Found channel for %s supplier -> %s = %v",
					supplier.Type,
					channelName,
					channel.Id())
				cfg.TelegramChannelsSession[supplier] = *channel
			}

			// Main polling loop: for each channel, fetch messages from offset,
			// start workflow per message, then persist offsets.
			interval := time.Duration(cfg.TelegramFetchInterval) * time.Second

			for {
				select {
				case <-ctx.Done():
					return ctx.Err()
				default:
				}

				var toPersist []domain.Message

				for supplier, ch := range cfg.TelegramChannelsSession {
					// Load the last processed message id (offset) per chat
					offset, err := db.GetLastMessageID(domain.ChatID(ch.Id()))
					if err != nil {
						log.Printf("get offset error for supplier %s: %v", supplier.Type, err)
						continue
					}

					// Fetch messages after offset
					msgs, err := ch.Messages(ctx, cfg.TelegramPageSize, int(offset))
					if err != nil {
						log.Printf("fetch messages error for supplier %s: %v", supplier.Type, err)
						continue
					}
					if len(msgs) == 0 {
						continue
					}

					// Start workflow per message
					for _, m := range msgs {
						if _, _, err := publisher.StartTelegramWorkflow(ctx, m); err != nil {
							log.Printf("start workflow error (supplier=%s, msg=%d): %v", supplier.Type, m.ID, err)
							// continue with other messages; offset will advance on successful saves
						}
					}

					// Collect for batch persistence of max offsets per chat
					toPersist = append(toPersist, msgs...)
				}

				// Persist offsets (per chat only max is sent inside)
				if len(toPersist) > 0 {
					if err := db.SaveLastMessageID(toPersist); err != nil {
						log.Printf("save offsets error: %v", err)
					}
				}

				// Sleep before the next iteration to avoid rate limits
				time.Sleep(interval)
			}
		})
		if err != nil {
			hs.SetReady(false)
			log.Printf("Telegram loop failed: %v", err)
		}
	}()

	// Wait for Telegram request to complete or a shutdown signal
	select {
	case <-telegramDone:
		log.Println("✅ Telegram loop completed, initiating shutdown...")
	case sig := <-sigChan:
		log.Printf("🛑 Received signal: %v, initiating immediate shutdown...", sig)
	}

	// Cancel context to signal all goroutines to stop
	hs.SetReady(false)
	cancel()

	// Graceful shutdown of HTTP servers
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	if err := hs.Shutdown(shutdownCtx); err != nil {
		log.Printf("HTTP server shutdown error: %v", err)
	}
	if err := ms.Shutdown(shutdownCtx); err != nil {
		log.Printf("Metrics server shutdown error: %v", err)
	}
	shutdownCancel()

	// Wait for all goroutines to finish
	wg.Wait()
	log.Println("Application stopped")
}
