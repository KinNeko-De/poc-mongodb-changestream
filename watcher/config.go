// config.go - Configuration for the change stream watcher
package main

import (
	"os"
	"strconv"
	"strings"
	"time"
)

// Config holds all configuration options for the watcher
type Config struct {
	// MongoDB connection
	MongoURI       string
	Database       string
	Collection     string
	ConnectTimeout time.Duration

	// Change stream settings
	ResumeTokenFile    string
	BatchSize          int32
	FullDocumentOption string

	// Reconnection settings
	InitialBackoff    time.Duration
	MaxBackoff        time.Duration
	BackoffMultiplier float64
	MaxReconnectTries int

	// Monitoring
	MetricsFile   string
	LogFile       string
	StatsInterval time.Duration

	// General settings
	Debug bool
}

// NewDefaultConfig creates a new configuration with default values
func NewDefaultConfig() *Config {
	return &Config{
		// MongoDB connection
		MongoURI:       getEnv("MONGO_URI", "mongodb://localhost:27017/?replicaSet=rs0"),
		Database:       getEnv("MONGO_DB", "changestream_poc"),
		Collection:     getEnv("MONGO_COLLECTION", "test_data"),
		ConnectTimeout: getDurationEnv("MONGO_CONNECT_TIMEOUT", 10*time.Second),

		// Change stream settings
		ResumeTokenFile:    getEnv("RESUME_TOKEN_FILE", "resume_token.json"),
		BatchSize:          getInt32Env("BATCH_SIZE", 100),
		FullDocumentOption: getEnv("FULL_DOCUMENT_OPTION", "updateLookup"),

		// Reconnection settings
		InitialBackoff:    getDurationEnv("INITIAL_BACKOFF", 1*time.Second),
		MaxBackoff:        getDurationEnv("MAX_BACKOFF", 1*time.Minute),
		BackoffMultiplier: getFloat64Env("BACKOFF_MULTIPLIER", 2.0),
		MaxReconnectTries: getIntEnv("MAX_RECONNECT_TRIES", 12),

		// Monitoring
		MetricsFile:   getEnv("METRICS_FILE", "watcher_metrics.json"),
		LogFile:       getEnv("LOG_FILE", "watcher.log"),
		StatsInterval: getDurationEnv("STATS_INTERVAL", 30*time.Second),

		// General settings
		Debug: getBoolEnv("DEBUG", false),
	}
}

// Helper functions to get environment variables with defaults

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getBoolEnv(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		b, err := strconv.ParseBool(value)
		if err == nil {
			return b
		}
	}
	return defaultValue
}

func getIntEnv(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		i, err := strconv.Atoi(value)
		if err == nil {
			return i
		}
	}
	return defaultValue
}

func getInt32Env(key string, defaultValue int32) int32 {
	if value := os.Getenv(key); value != "" {
		i, err := strconv.ParseInt(value, 10, 32)
		if err == nil {
			return int32(i)
		}
	}
	return defaultValue
}

func getFloat64Env(key string, defaultValue float64) float64 {
	if value := os.Getenv(key); value != "" {
		f, err := strconv.ParseFloat(value, 64)
		if err == nil {
			return f
		}
	}
	return defaultValue
}

func getDurationEnv(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		// Try to parse as seconds first
		if seconds, err := strconv.ParseInt(value, 10, 64); err == nil {
			return time.Duration(seconds) * time.Second
		}

		// Try to parse as duration string
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}

// ValidateConfig checks if the configuration is valid
func (c *Config) ValidateConfig() []string {
	var errors []string

	// Check if MongoDB URI is valid
	if !strings.Contains(c.MongoURI, "replicaSet") {
		errors = append(errors, "MongoURI must contain a replica set parameter for change streams to work")
	}

	// Check if database and collection are specified
	if c.Database == "" {
		errors = append(errors, "Database must be specified")
	}

	if c.Collection == "" {
		errors = append(errors, "Collection must be specified")
	}

	return errors
}
