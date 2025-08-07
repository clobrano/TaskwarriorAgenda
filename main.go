/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package main

import (
	"log"
	"os"
	"path/filepath"

	"github.com/clobrano/TaskwarriorAgenda/cmd"
	"github.com/spf13/viper"
)

func main() {
	// Get the XDG config directory
	xdgConfigDir, err := os.UserConfigDir()
	if err != nil {
		log.Fatalf("Error getting config directory: %s", err)
	}
	configDir := filepath.Join(xdgConfigDir, "taskwarrior-agenda")

	// Set the configuration file name and path
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(configDir) // XDG config directory
	viper.AddConfigPath(".")       // optionally look for config in the working directory

	// Read the configuration file
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Error reading config file, %s", err)
	}

	cmd.Execute()
}
