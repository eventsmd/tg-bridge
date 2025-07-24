package tgclient

import (
	"bufio"
	"context"
	"fmt"
	"os"

	"github.com/gotd/td/session"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/auth"
	"github.com/gotd/td/tg"
)

type AuthParams struct {
	TelegramApiId   int
	TelegramApiHash string
	Phone           string
	Code            string
	Password        string
	SessionDir      string
}

func AuthWithPhoneNumber(params AuthParams) error {
	ctx := context.Background()

	storage := &session.FileStorage{
		Path: params.SessionDir + "/session.json",
	}
	client := telegram.NewClient(params.TelegramApiId, params.TelegramApiHash, telegram.Options{
		SessionStorage: storage,
	})
	err := client.Run(ctx, func(ctx context.Context) error {
		status, err := client.Auth().Status(ctx)
		if err != nil {
			return fmt.Errorf("failed to get auth status: %w", err)
		}

		if !status.Authorized {
			fmt.Println("Not authorized, starting authentication flow...")
			flow := auth.NewFlow(
				auth.Constant(params.Phone, "password", auth.CodeAuthenticatorFunc(codePrompt)),
				auth.SendCodeOptions{},
			)
			if err := client.Auth().IfNecessary(ctx, flow); err != nil {
				return fmt.Errorf("failed to authenticate: %w", err)
			}
		}

		fmt.Println("Successfully authenticated!")
		fmt.Println("Session has been created and saved.")

		api := client.API()
		user, err := api.UsersGetFullUser(ctx, &tg.InputUserSelf{})
		if err != nil {
			return fmt.Errorf("failed to get user info: %w", err)
		}

		fmt.Printf("Logged in as: %s\n", user.Users[0].(*tg.User).FirstName)

		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

func codePrompt(_ context.Context, _ *tg.AuthSentCode) (string, error) {
	fmt.Print("Enter verification code: ")
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	return scanner.Text(), scanner.Err()
}
