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
	DryRun          bool
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
	if params.DryRun {
		return generateSampleSession(ctx, storage)
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
			codeErr := initiateAuthCodeRequest(ctx, params, client)
			if codeErr != nil {
				return codeErr
			}
		}

		fmt.Println("Successfully authenticated!")
		fmt.Println("Session has been created and saved.")

		userErr := readAuthenticatedUserInfo(ctx, client)
		if userErr != nil {
			return userErr
		}

		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

func readAuthenticatedUserInfo(ctx context.Context, client *telegram.Client) error {
	api := client.API()
	user, err := api.UsersGetFullUser(ctx, &tg.InputUserSelf{})
	if err != nil {
		return fmt.Errorf("failed to get user info: %w", err)
	}
	fmt.Printf("Logged in as: %s\n", user.Users[0].(*tg.User).FirstName)
	return nil
}

func generateSampleSession(ctx context.Context, session *session.FileStorage) error {
	// sample session with minimally required session fields
	sessionJson := `{
  "Version": 1,
  "Data": {
    "DC": 0,
    "Addr": "",
    "AuthKey": "AUTH_KEY_HERE",
    "AuthKeyID": "AUTH_KEY_ID_HERE",
    "Salt": 12345
  }
}`
	err := session.StoreSession(ctx, []byte(sessionJson))
	if err != nil {
		return fmt.Errorf("failed to save stub sesion to file: %w", err)
	}
	return nil
}

func initiateAuthCodeRequest(ctx context.Context, params AuthParams, client *telegram.Client) error {
	flow := auth.NewFlow(
		auth.Constant(params.Phone, "password", auth.CodeAuthenticatorFunc(codePrompt)),
		auth.SendCodeOptions{},
	)
	if err := client.Auth().IfNecessary(ctx, flow); err != nil {
		return fmt.Errorf("failed to authenticate: %w", err)
	}
	return nil
}

func codePrompt(_ context.Context, _ *tg.AuthSentCode) (string, error) {
	fmt.Print("Enter verification code: ")
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	return scanner.Text(), scanner.Err()
}
