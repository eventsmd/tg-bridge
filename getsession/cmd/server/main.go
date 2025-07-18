package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"tg-bridge/getsession/internal/config"

	"github.com/gotd/td/session"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/auth"
	"github.com/gotd/td/tg"
)

func main() {
	cfg := config.LoadConfig()

	// Create session storage (file-based)
	sessionDir := "./generated-session"
	if err := os.MkdirAll(sessionDir, 0700); err != nil {
		panic(err)
	}

	// Create session storage
	storage := &session.FileStorage{
		Path: sessionDir + "/session.json",
	}

	// Create Telegram client
	client := telegram.NewClient(cfg.TelegramApiId, cfg.TelegramApiHash, telegram.Options{
		SessionStorage: storage,
	})

	// Run the client
	err := client.Run(context.Background(), func(ctx context.Context) error {
		// Check if we're already authorized
		status, err := client.Auth().Status(ctx)
		if err != nil {
			return fmt.Errorf("failed to get auth status: %w", err)
		}

		if !status.Authorized {
			fmt.Println("Not authorized, starting authentication flow...")

			// Create flow for authentication
			flow := auth.NewFlow(
				auth.Constant(cfg.Phone, "password", auth.CodeAuthenticatorFunc(codePrompt)),
				auth.SendCodeOptions{},
			)

			// Start authentication
			if err := client.Auth().IfNecessary(ctx, flow); err != nil {
				return fmt.Errorf("failed to authenticate: %w", err)
			}
		}

		fmt.Println("Successfully authenticated!")
		fmt.Println("Session has been created and saved.")

		// Test the session by getting user info
		api := client.API()
		user, err := api.UsersGetFullUser(ctx, &tg.InputUserSelf{})
		if err != nil {
			return fmt.Errorf("failed to get user info: %w", err)
		}

		fmt.Printf("Logged in as: %s\n", user.Users[0].(*tg.User).FirstName)

		return nil
	})

	if err != nil {
		panic(err)
	}
}

// Custom code prompt
func codePrompt(ctx context.Context, sentCode *tg.AuthSentCode) (string, error) {
	fmt.Print("Enter verification code: ")
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	return scanner.Text(), scanner.Err()
}
