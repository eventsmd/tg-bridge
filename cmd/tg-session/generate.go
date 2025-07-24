package main

import (
	"errors"
	"tg-bridge/internal/tgclient"

	"github.com/spf13/cobra"
)

var (
	apiId      int
	apiHash    string
	phone      string
	sessionDir string
)

// generateCmd defines the `generate` subcommand
var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate a Telegram session using a phone number",
	Long: `Generates and stores a Telegram session file.

Examples:
  tg-session generate --apiId 123 --apiHash "123456" --phone "+1234567890" --sessionDir "./session-dir"`,

	RunE: func(cmd *cobra.Command, args []string) error {
		if phone != "" ||
			apiId != 0 ||
			apiHash != "" ||
			sessionDir != "" {
			return errors.New("one of the params is missing. See help for more details")
		}

		return tgclient.AuthWithPhoneNumber(tgclient.AuthParams{
			Phone:           phone,
			TelegramApiId:   apiId,
			TelegramApiHash: apiHash,
			SessionDir:      sessionDir,
		})
	},
}

func init() {
	generateCmd.Flags().IntVar(&apiId, "apiId", 0, "Telegram API ID")
	generateCmd.Flags().StringVar(&apiHash, "apiHash", "", "Telegram API Hash")
	generateCmd.Flags().StringVar(&phone, "phone", "", "Phone number to log in with (for user account)")
	generateCmd.Flags().StringVar(&sessionDir, "session-dir", "./generated-session",
		"Directory to store generated session files")
	rootCmd.AddCommand(generateCmd)
}
