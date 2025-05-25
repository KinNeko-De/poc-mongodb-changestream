package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// ResumeToken represents the structure to store in the token file
type ResumeToken struct {
	Token     bson.Raw  `json:"token"`
	Timestamp time.Time `json:"timestamp"`
}

// Global variables
var (
	config  *Config
	metrics *Metrics
)

func main() {
	// Load configuration
	config = NewDefaultConfig()

	// Validate configuration
	if errors := config.ValidateConfig(); len(errors) > 0 {
		for _, err := range errors {
			fmt.Printf("Configuration error: %s\n", err)
		}
		os.Exit(1)
	}

	// Initialize metrics
	metrics = NewMetrics(config.MetricsFile)

	// Set up logging
	log.SetFlags(log.LstdFlags | log.Lmicroseconds | log.Lshortfile)
	logFile := setupLogFile(config.LogFile)
	defer logFile.Close()

	log.Printf("Starting MongoDB Change Stream Watcher v1.1.0")
	log.Printf("Configuration: Database=%s, Collection=%s", config.Database, config.Collection)

	if config.Debug {
		log.Printf("Debug mode enabled. Configuration details:")
		log.Printf("MongoDB URI: %s", sanitizeMongoURI(config.MongoURI))
		log.Printf("Resume Token File: %s", config.ResumeTokenFile)
		log.Printf("Reconnect settings: Initial=%v, Max=%v, Multiplier=%.1f, MaxTries=%d",
			config.InitialBackoff, config.MaxBackoff, config.BackoffMultiplier, config.MaxReconnectTries)
	}

	// Create a channel to handle OS signals for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Create context for overall application
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start metrics reporter
	go reportMetricsPeriodically(ctx, metrics, config.StatsInterval)

	// Start the watcher with reconnection logic
	var wg sync.WaitGroup
	wg.Add(1)
	go runWatcherWithReconnection(ctx, &wg, metrics, config)

	// Wait for interrupt signal
	sig := <-sigChan
	fmt.Printf("\nReceived interrupt signal (%s), shutting down gracefully...\n", sig)
	cancel()

	// Wait with timeout for graceful shutdown
	waitWithTimeout(&wg, 5*time.Second)

	// Save final metrics
	metrics.PrintStats()

	fmt.Println("Watcher has shut down")
}

// waitWithTimeout waits for the WaitGroup with a timeout
func waitWithTimeout(wg *sync.WaitGroup, timeout time.Duration) {
	// Create a channel that is closed when the wait group is done
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	// Wait for either the wait group to finish or the timeout
	select {
	case <-done:
		// Wait group finished normally
		return
	case <-time.After(timeout):
		// Wait group timed out
		log.Println("Warning: Graceful shutdown timed out")
		return
	}
}

// sanitizeMongoURI returns a sanitized version of the MongoDB URI for logging
func sanitizeMongoURI(uri string) string {
	// Very basic sanitization - a proper implementation would use a URL parser
	if uri == "" {
		return ""
	}

	// If it contains @ (authentication), replace the credentials
	if i := indexOf(uri, "@"); i > 0 {
		// Find the start of the credentials (after //)
		start := indexOf(uri, "//")
		if start > 0 {
			start += 2
			// Return the sanitized URI
			return uri[:start] + "***:***" + uri[i:]
		}
	}

	return uri
}

// indexOf returns the index of the first occurrence of sep in s, or -1 if sep is not present
func indexOf(s, sep string) int {
	for i := 0; i <= len(s)-len(sep); i++ {
		if s[i:i+len(sep)] == sep {
			return i
		}
	}
	return -1
}

// setupLogFile creates a log file for the application
func setupLogFile(logFilePath string) *os.File {
	logFile, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}

	// Configure the logger to write to both stdout and the log file
	multiWriter := io.MultiWriter(os.Stdout, logFile)
	log.SetOutput(multiWriter)

	return logFile
}

// reportMetricsPeriodically reports metrics at regular intervals
func reportMetricsPeriodically(ctx context.Context, m *Metrics, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			m.PrintStats()
		}
	}
}

// runWatcherWithReconnection handles MongoDB connection with automatic reconnection
func runWatcherWithReconnection(ctx context.Context, wg *sync.WaitGroup, metrics *Metrics, config *Config) {
	defer wg.Done()

	// Keep track of reconnection attempts
	var reconnectCount int

	log.Printf("Starting watcher with reconnection logic (max attempts: %d)", config.MaxReconnectTries)

	for {
		select {
		case <-ctx.Done():
			log.Println("Context cancelled, stopping watcher reconnection")
			return
		default:
			// Connect and run the watcher
			err := connectAndWatch(ctx, metrics, config)

			// Check if we should exit
			if ctx.Err() != nil {
				// Context was cancelled, exit normally
				return
			}

			// Handle the error
			if err != nil {
				// Categorize and log the error
				LogError(err, "Watcher disconnected")

				// Increment reconnect counter
				reconnectCount++
				metrics.IncrementFailedReconnect()

				// Check if we've exceeded the maximum reconnection attempts
				if reconnectCount > config.MaxReconnectTries {
					log.Printf("Exceeded maximum reconnection attempts (%d). Giving up.", config.MaxReconnectTries)
					metrics.SetUnhealthy()
					return
				}

				// Calculate backoff duration using exponential backoff
				backoffDuration := CalculateBackoff(reconnectCount, config)

				log.Printf("Reconnecting in %v (attempt %d/%d)...",
					backoffDuration, reconnectCount, config.MaxReconnectTries)

				// Wait before reconnecting
				select {
				case <-time.After(backoffDuration):
					// Continue to reconnect
				case <-ctx.Done():
					return
				}
			} else {
				// If we get here without an error, it was a clean shutdown
				// Reset the reconnect counter
				reconnectCount = 0
				log.Println("Watcher exited cleanly")
				return
			}
		}
	}
}

// connectAndWatch establishes a MongoDB connection and starts watching the change stream
func connectAndWatch(ctx context.Context, metrics *Metrics, config *Config) error {
	// Use a separate context for this connection attempt with timeout
	connCtx, connCancel := context.WithTimeout(ctx, config.ConnectTimeout)
	defer connCancel()

	// Connect to MongoDB
	log.Printf("Connecting to MongoDB at: %s", sanitizeMongoURI(config.MongoURI))

	// Configure client options
	clientOptions := options.Client().ApplyURI(config.MongoURI)

	// Apply additional options
	clientOptions.SetHosts([]string{"localhost:27017"})
	clientOptions.SetConnectTimeout(config.ConnectTimeout)
	clientOptions.SetServerSelectionTimeout(config.ConnectTimeout)

	// Create the client
	client, err := mongo.Connect(connCtx, clientOptions)
	if err != nil {
		return fmt.Errorf("failed to connect to MongoDB: %w", err)
	}
	defer func() {
		if err := client.Disconnect(context.Background()); err != nil {
			log.Printf("Warning: Failed to disconnect MongoDB client: %v", err)
		}
	}()

	// Test the connection with a ping using a read preference of Primary
	pingCtx, pingCancel := context.WithTimeout(ctx, 5*time.Second)
	defer pingCancel()

	if err := client.Ping(pingCtx, readpref.Primary()); err != nil {
		return fmt.Errorf("failed to ping MongoDB primary: %w", err)
	}

	log.Println("Connected to MongoDB successfully!")
	metrics.IncrementSuccessfulReconnect()
	metrics.updateHeartbeat()

	// Get the collection
	coll := client.Database(config.Database).Collection(config.Collection)

	// Load the resume token from file if it exists
	var resumeAfter bson.Raw
	var startAfter bson.Raw
	skipResumeToken := false

	resumeToken, err := loadResumeToken(config.ResumeTokenFile)
	if err == nil && resumeToken != nil && len(resumeToken.Token) > 0 {
		age := time.Since(resumeToken.Timestamp)
		log.Printf("Found resume token from %v ago (%v)", age.Round(time.Second), resumeToken.Timestamp)

		// Use the appropriate resume mechanism based on token age
		if age > 24*time.Hour {
			// For very old tokens, use startAfter which is more resilient
			log.Printf("Resume token is older than 24 hours, using startAfter instead of resumeAfter")
			startAfter = resumeToken.Token
		} else {
			resumeAfter = resumeToken.Token
		}
	} else {
		if err != nil {
			log.Printf("No valid resume token found: %v", err)
		} else {
			log.Printf("No resume token found, starting new change stream")
		}
		skipResumeToken = true
	}

	// Create change stream options
	csOptions := options.ChangeStream()
	// Configure the change stream options
	csOptions.SetBatchSize(config.BatchSize)

	// Set FullDocument option based on configuration
	switch config.FullDocumentOption {
	case "updateLookup":
		csOptions.SetFullDocument(options.UpdateLookup)
	case "whenAvailable":
		csOptions.SetFullDocument(options.WhenAvailable)
	case "required":
		csOptions.SetFullDocument(options.Required)
	default:
		// Default to UpdateLookup
		csOptions.SetFullDocument(options.UpdateLookup)
	}

	// Set the resume mechanism
	if !skipResumeToken {
		if resumeAfter != nil {
			log.Printf("Resuming change stream using resumeAfter")
			csOptions.SetResumeAfter(resumeAfter)
		} else if startAfter != nil {
			log.Printf("Resuming change stream using startAfter")
			csOptions.SetStartAfter(startAfter)
		}
	}

	// Create the change stream with a pipeline that matches the collection
	log.Println("Starting change stream watcher...")
	pipeline := mongo.Pipeline{}

	watchCtx, watchCancel := context.WithCancel(ctx)
	defer watchCancel()

	changeStream, err := coll.Watch(watchCtx, pipeline, csOptions)
	if err != nil {
		// If there was an error with the resume token, try again without it
		if CategorizeError(err) == ErrorCategoryResumeToken && !skipResumeToken {
			log.Printf("Error with resume token, retrying without it: %v", err)
			metrics.IncrementResumeError()

			// Retry without the resume token
			csOptions = options.ChangeStream()
			csOptions.SetBatchSize(config.BatchSize)

			// Set the proper full document option
			switch config.FullDocumentOption {
			case "updateLookup":
				csOptions.SetFullDocument(options.UpdateLookup)
			case "whenAvailable":
				csOptions.SetFullDocument(options.WhenAvailable)
			case "required":
				csOptions.SetFullDocument(options.Required)
			default:
				csOptions.SetFullDocument(options.UpdateLookup)
			}

			// Try again with a new context
			retryCtx, retryCancel := context.WithCancel(ctx)
			defer retryCancel()

			changeStream, err = coll.Watch(retryCtx, pipeline, csOptions)
			if err != nil {
				return fmt.Errorf("failed to create change stream (retry without token): %w", err)
			}
		} else {
			return fmt.Errorf("failed to create change stream: %w", err)
		}
	}
	defer changeStream.Close(watchCtx)

	log.Println("Change stream established successfully. Watching for changes...")

	// Create heartbeat ticker for keeping the connection alive
	heartbeatTicker := time.NewTicker(30 * time.Second)
	defer heartbeatTicker.Stop()

	// Create a channel for the heartbeat goroutine
	heartbeatDone := make(chan struct{})
	defer close(heartbeatDone)

	// Start heartbeat goroutine
	go func() {
		for {
			select {
			case <-heartbeatDone:
				return
			case <-heartbeatTicker.C:
				// Update heartbeat and check connection
				metrics.updateHeartbeat()

				// Ping the server to keep the connection alive
				pingCtx, pingCancel := context.WithTimeout(ctx, 5*time.Second)
				err := client.Ping(pingCtx, readpref.Primary())
				pingCancel()

				if err != nil {
					log.Printf("Warning: Failed to ping MongoDB during heartbeat: %v", err)
				}
			}
		}
	}()

	// Monitor the change stream for events
	for changeStream.Next(watchCtx) {
		var changeEvent bson.M
		if err := changeStream.Decode(&changeEvent); err != nil {
			log.Printf("Error decoding change event: %v", err)
			metrics.IncrementProcessingError()
			continue
		}
		// Process the change event
		processChangeEvent(changeEvent, metrics)

		// Save the resume token after processing
		if token := changeStream.ResumeToken(); token != nil {
			if err := saveResumeToken(config.ResumeTokenFile, token); err != nil {
				log.Printf("Failed to save resume token: %v", err)
			}
		}
	}

	// Check if the change stream was closed due to an error
	if err := changeStream.Err(); err != nil {
		// Don't report error if the context was cancelled
		if ctx.Err() != nil {
			return nil
		}
		return fmt.Errorf("change stream error: %w", err)
	}

	// If we get here, the context was likely cancelled
	return nil
}

// processChangeEvent handles incoming change stream events
func processChangeEvent(event bson.M, metrics *Metrics) {
	operationType, ok := event["operationType"].(string)
	if !ok {
		log.Println("Could not determine operation type")
		metrics.IncrementProcessingError()
		return
	}

	fmt.Printf("Change detected: %s\n", operationType)

	switch operationType {
	case "insert":
		metrics.IncrementInsert()
		handleInsert(event)
	case "update":
		metrics.IncrementUpdate()
		handleUpdate(event)
	case "replace":
		metrics.IncrementReplace()
		handleReplace(event)
	case "delete":
		metrics.IncrementDelete()
		handleDelete(event)
	default:
		metrics.IncrementOtherOp()
		fmt.Printf("Unhandled operation type: %s\n", operationType)
	}
}

// handleInsert processes insert operations
func handleInsert(event bson.M) {
	fullDocument, ok := event["fullDocument"].(bson.M)
	if !ok {
		log.Println("Could not extract full document from insert event")
		return
	}

	fmt.Printf("Document inserted: %+v\n", fullDocument)

	// TODO: Add custom logic for handling insert operations
	// This is where you can add specific business logic when a document is inserted
	// Examples:
	// - Send notifications
	// - Update caches
	// - Trigger workflows
	// - Log to external systems
	// - Update metrics
}

// handleUpdate processes update operations
func handleUpdate(event bson.M) {
	documentKey, ok := event["documentKey"].(bson.M)
	if !ok {
		log.Println("Could not extract document key from update event")
		return
	}

	fmt.Printf("Document updated with key: %+v\n", documentKey)

	// Check if we have the full document after update
	if fullDocument, exists := event["fullDocument"]; exists {
		fmt.Printf("Updated document: %+v\n", fullDocument)
	}

	// TODO: Add custom logic for handling update operations
	// This is where you can add specific business logic when a document is updated
}

// handleReplace processes replace operations
func handleReplace(event bson.M) {
	documentKey, ok := event["documentKey"].(bson.M)
	if !ok {
		log.Println("Could not extract document key from replace event")
		return
	}

	fmt.Printf("Document replaced with key: %+v\n", documentKey)

	if fullDocument, exists := event["fullDocument"]; exists {
		fmt.Printf("Replacement document: %+v\n", fullDocument)
	}

	// TODO: Add custom logic for handling replace operations
	// This is where you can add specific business logic when a document is replaced
}

// handleDelete processes delete operations
func handleDelete(event bson.M) {
	documentKey, ok := event["documentKey"].(bson.M)
	if !ok {
		log.Println("Could not extract document key from delete event")
		return
	}

	fmt.Printf("Document deleted with key: %+v\n", documentKey)

	// TODO: Add custom logic for handling delete operations
	// This is where you can add specific business logic when a document is deleted
}

// saveResumeToken saves the change stream resume token to a file
func saveResumeToken(filename string, token bson.Raw) error {
	if token == nil {
		return fmt.Errorf("cannot save nil resume token")
	}

	resumeToken := ResumeToken{
		Token:     token,
		Timestamp: time.Now(),
	}

	data, err := json.Marshal(resumeToken)
	if err != nil {
		return fmt.Errorf("failed to marshal resume token: %w", err)
	}

	// Write to a temporary file first, then rename for atomicity
	tempFile := filename + ".tmp"
	if err := os.WriteFile(tempFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write resume token to temp file: %w", err)
	}

	if err := os.Rename(tempFile, filename); err != nil {
		return fmt.Errorf("failed to rename temp file to resume token file: %w", err)
	}

	return nil
}

// loadResumeToken loads a previously saved resume token from a file
func loadResumeToken(filename string) (*ResumeToken, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read resume token file: %w", err)
	}

	var resumeToken ResumeToken
	if err := json.Unmarshal(data, &resumeToken); err != nil {
		return nil, fmt.Errorf("failed to unmarshal resume token: %w", err)
	}

	return &resumeToken, nil
}
