/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"log"
	"os"
	"path/filepath"

	"github.com/clobrano/TaskwarriorAgenda/pkg/auth" // Adjust import path
	"github.com/spf13/cobra"
)

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Remove existing Token file and re-authenticate with Google Calendar API",
	Long: `Authenticates with your Google account to access Google Calendar.
This command will guide you through the OAuth 2.0 process to get the necessary
tokens for API access.`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()

		// Delete existing token file
		xdgConfigBase, err := auth.GetXdgHome()
		if err != nil {
			log.Fatalf("could not find path to configuration file: error %v", err)
			return
		}

		tokenFile := filepath.Join(xdgConfigBase, auth.TokenFile)
		_, err = os.Stat(tokenFile)
		if err != nil {
			if !os.IsNotExist(err) {
				log.Fatalf("could not find token file '%s', error %v. Please delete it manually", tokenFile, err)
			}
		} else {
			log.Printf("Removing existing token file at '%s'\n", tokenFile)
			if err = os.Remove(tokenFile); err != nil {
				log.Fatalf("could not delete token file '%s', error %v. Please delete it manually", tokenFile, err)
			}
		}

		// GetCalendarService will handle the full OAuth flow if needed,
		// including opening the browser and capturing the token.
		_, err = auth.GetCalendarService(ctx)
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
