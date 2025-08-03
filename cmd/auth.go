/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"log"

	"github.com/clobrano/TaskwarriorAgenda/pkg/auth" // Adjust import path
	"github.com/spf13/cobra"
)

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Authenticate with Google Calendar API",
	Long: `Authenticates with your Google account to access Google Calendar.
This command will guide you through the OAuth 2.0 process to get the necessary
tokens for API access.`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()

		// GetCalendarService will handle the full OAuth flow if needed,
		// including opening the browser and capturing the token.
		_, err := auth.GetCalendarService(ctx)
		if err != nil {
			log.Fatalf("Authentication failed: %v", err)
		}
		log.Printf("Authentication successful! Token saved to %s", auth.TokenFile)
		log.Println("You can now run 'TaskwarriorAgenda sync' to synchronize your tasks.")
	},
}

func init() {
	rootCmd.AddCommand(authCmd)
}
