package tgclient

import (
	"bufio"
	"context"
	"encoding/base64"
	"log"
	"os"
	"tg-bridge/internal/tgsession"

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
	Password        string
}

func CreateTelegramClient(apiId int, apiHash string, sessionStorage session.Storage) *telegram.Client {
	return telegram.NewClient(apiId, apiHash, telegram.Options{
		SessionStorage: sessionStorage,
	})
}

func CheckTelegramSession(ctx context.Context, client *telegram.Client, onNotAuthorized func() error) error {
	status, err := client.Auth().Status(ctx)
	if err != nil {
		return err
	}
	if !status.Authorized {
		err := onNotAuthorized()
		if err != nil {
			return err
		}
	}
	log.Println("Successfully authenticated")
	return nil
}

func AuthWithPhoneNumber(params AuthParams) error {
	ctx := context.Background()
	storage := tgsession.NewMemorySessionStorage([]byte{})
	if params.DryRun {
		return GenerateSampleSession(ctx, storage)
	}

	client := CreateTelegramClient(params.TelegramApiId, params.TelegramApiHash, storage)
	err := client.Run(ctx, func(ctx context.Context) error {
		err := initiateAuthCodeRequest(ctx, params, client)
		if err != nil {
			return err
		}
		errSession := PrintEncodedSession(ctx, storage)
		if errSession != nil {
			return errSession
		}

		return ReadAuthenticatedUserInfo(ctx, client)
	})
	if err != nil {
		return err
	}
	return nil
}

func PrintEncodedSession(ctx context.Context, sessionStorage session.Storage) error {
	loadedSession, err := sessionStorage.LoadSession(ctx)
	if err != nil {
		return err
	}
	shortSession, err := tgsession.TrimSession(loadedSession)
	if err != nil {
		return err
	}
	encodedSession := base64.StdEncoding.EncodeToString(shortSession)
	log.Printf("Base64 encoded session: %s", encodedSession)
	return nil
}

func ReadAuthenticatedUserInfo(ctx context.Context, client *telegram.Client) error {
	api := client.API()
	user, err := api.UsersGetFullUser(ctx, &tg.InputUserSelf{})
	if err != nil {
		log.Printf("failed to get user info: %s", err)
		return err
	}
	log.Printf("Logged in as: %s\n", user.Users[0].(*tg.User).FirstName)
	return nil
}

func initiateAuthCodeRequest(ctx context.Context, params AuthParams, client *telegram.Client) error {
	flow := auth.NewFlow(
		auth.Constant(params.Phone, "password", auth.CodeAuthenticatorFunc(codePrompt)),
		auth.SendCodeOptions{},
	)
	if err := client.Auth().IfNecessary(ctx, flow); err != nil {
		log.Printf("failed to authenticate: %s", err)
		return err
	}
	return nil
}

func codePrompt(_ context.Context, _ *tg.AuthSentCode) (string, error) {
	log.Print("Enter verification code: ")
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	return scanner.Text(), scanner.Err()
}
