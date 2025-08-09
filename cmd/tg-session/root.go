package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// rootCmd defines the `tg-session` command
var rootCmd = &cobra.Command{
	Use:   "tg-session",
	Short: "A CLI tool to generate Telegram sessions for user accounts",
	Long:  `tg-session is a command-line utility for generating Telegram sessions using a phone number.`,
	// No Run here — root just serves as a base for subcommands
}

// Execute is called by main.main()
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
