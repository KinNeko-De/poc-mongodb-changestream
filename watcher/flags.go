// flags.go - Command line flags for the watcher application
package main

import (
	"flag"
	"fmt"
	"os"
	"time"
)

// CommandFlags holds the command line flags for the application
type CommandFlags struct {
	ResetToken    bool
	StartFromTime string
	ConfigFile    string
	ReplayAll     bool
	Help          bool
}

// ParseFlags parses the command line flags
func ParseFlags() *CommandFlags {
	flags := &CommandFlags{}

	flag.BoolVar(&flags.ResetToken, "reset-token", false, "Reset the resume token and start a new change stream")
	flag.StringVar(&flags.StartFromTime, "start-from", "", "Start change stream from a specific time (format: 2006-01-02T15:04:05Z)")
	flag.StringVar(&flags.ConfigFile, "config", "", "Path to a configuration file")
	flag.BoolVar(&flags.ReplayAll, "replay-all", false, "Replay all existing documents as insert events before watching for changes")
	flag.BoolVar(&flags.Help, "help", false, "Show help")

	// Parse the flags
	flag.Parse()

	// Show help if requested
	if flags.Help {
		printUsage()
		os.Exit(0)
	}

	return flags
}

// printUsage prints the usage information
func printUsage() {
	fmt.Println("MongoDB Change Stream Watcher")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  watcher [options]")
	fmt.Println()
	fmt.Println("Options:")
	flag.PrintDefaults()
	fmt.Println()
	fmt.Println("Environment Variables:")
	fmt.Println("  MONGO_URI             MongoDB connection string")
	fmt.Println("  MONGO_DB              Database name")
	fmt.Println("  MONGO_COLLECTION      Collection name")
	fmt.Println("  RESUME_TOKEN_FILE     File to store the resume token")
	fmt.Println("  DEBUG                 Enable debug mode (true/false)")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  watcher --reset-token")
	fmt.Println("  watcher --start-from=2023-05-01T00:00:00Z")
	fmt.Println("  watcher --replay-all")
	fmt.Println("  MONGO_URI=mongodb://localhost:27017/?replicaSet=rs0 watcher")
}

// HandleFlags processes the command line flags
func HandleFlags(flags *CommandFlags, config *Config) error {
	// Handle reset token flag
	if flags.ResetToken {
		fmt.Println("Resetting resume token...")
		if err := os.Remove(config.ResumeTokenFile); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("failed to remove resume token file: %w", err)
		}
		fmt.Printf("Resume token reset. File removed: %s\n", config.ResumeTokenFile)
	}

	// Handle start from time
	if flags.StartFromTime != "" {
		t, err := time.Parse(time.RFC3339, flags.StartFromTime)
		if err != nil {
			return fmt.Errorf("invalid time format for --start-from: %w", err)
		}
		// Set the start time in config for later use
		config.StartTime = &t
		fmt.Printf("Will start change stream from: %s\n", t.Format(time.RFC3339))
	}

	// Handle replay-all flag
	if flags.ReplayAll {
		fmt.Println("Will replay all documents in the collection before watching for changes")
		// Actual replay is handled in main.go
	}

	return nil
}
