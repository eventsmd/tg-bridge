package main

import (
	"tg-bridge/internal/tgclient"

	"github.com/spf13/cobra"
)

var (
	dryRun  bool
	apiId   int
	apiHash string
	phone   string
)

// generateCmd defines the `generate` subcommand
var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate a Telegram session using a phone number",
	Long: `Generates and prints base64 encoded Telegram session string.

Examples:
  tg-session generate --apiId 123 --apiHash "123456" --phone "+1234567890"`,

	RunE: func(cmd *cobra.Command, args []string) error {
		return tgclient.AuthWithPhoneNumber(tgclient.AuthParams{
			DryRun:          dryRun,
			Phone:           phone,
			TelegramApiId:   apiId,
			TelegramApiHash: apiHash,
		})
	},
}

func init() {
	generateCmd.Flags().BoolVarP(&dryRun, "dry-run", "d", false, "Perform a dry run without generating the actual session")
	generateCmd.Flags().IntVar(&apiId, "apiId", 0, "Telegram API ID")
	generateCmd.Flags().StringVar(&apiHash, "apiHash", "", "Telegram API Hash")
	generateCmd.Flags().StringVar(&phone, "phone", "", "Phone number to log in with (for user account)")

	requiredFlags := []string{"apiId", "apiHash", "phone"}
	for _, flag := range requiredFlags {
		_ = generateCmd.MarkFlagRequired(flag)
	}
	rootCmd.AddCommand(generateCmd)
}
