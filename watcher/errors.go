// errors.go - Error handling for the change stream watcher
package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
)

// ErrorCategory represents the category of an error
type ErrorCategory int

const (
	// ErrorCategoryUnknown represents an unknown error
	ErrorCategoryUnknown ErrorCategory = iota

	// ErrorCategoryConnection represents a connection error
	ErrorCategoryConnection

	// ErrorCategoryTransient represents a transient error that can be retried
	ErrorCategoryTransient

	// ErrorCategoryFatal represents a fatal error that cannot be recovered from
	ErrorCategoryFatal

	// ErrorCategoryResumeToken represents an error with the resume token
	ErrorCategoryResumeToken
)

// String returns the string representation of the error category
func (c ErrorCategory) String() string {
	switch c {
	case ErrorCategoryUnknown:
		return "Unknown"
	case ErrorCategoryConnection:
		return "Connection"
	case ErrorCategoryTransient:
		return "Transient"
	case ErrorCategoryFatal:
		return "Fatal"
	case ErrorCategoryResumeToken:
		return "ResumeToken"
	default:
		return fmt.Sprintf("ErrorCategory(%d)", c)
	}
}

// CategorizeError categorizes an error into an error category
func CategorizeError(err error) ErrorCategory {
	if err == nil {
		return ErrorCategoryUnknown
	}

	// Check for context cancellation (not an error)
	if errors.Is(err, context.Canceled) {
		return ErrorCategoryUnknown
	}

	// Check for connection errors
	if mongo.IsNetworkError(err) {
		return ErrorCategoryConnection
	}

	// Check for timeout errors
	if errors.Is(err, context.DeadlineExceeded) {
		return ErrorCategoryTransient
	}

	// Check for resume token errors
	errStr := err.Error()
	if strings.Contains(errStr, "resume token") ||
		strings.Contains(errStr, "resume point") ||
		strings.Contains(errStr, "change stream resumable error") {
		return ErrorCategoryResumeToken
	}

	// Check for transient MongoDB errors
	if mongo.IsTimeout(err) || isRetryableWrite(err) || isRetryableRead(err) {
		return ErrorCategoryTransient
	}

	// Check for fatal errors that cannot be recovered from
	if isFatalError(err) {
		return ErrorCategoryFatal
	}

	// Default to unknown
	return ErrorCategoryUnknown
}

// isRetryableWrite checks if the error is a retryable write error
func isRetryableWrite(err error) bool {
	// MongoDB driver doesn't expose a public API for this, so we check the error message
	errStr := err.Error()
	return strings.Contains(errStr, "not master") ||
		strings.Contains(errStr, "node is recovering") ||
		strings.Contains(errStr, "replSetStepDown")
}

// isRetryableRead checks if the error is a retryable read error
func isRetryableRead(err error) bool {
	// MongoDB driver doesn't expose a public API for this, so we check the error message
	errStr := err.Error()
	return strings.Contains(errStr, "not master") ||
		strings.Contains(errStr, "node is recovering")
}

// isFatalError checks if the error is a fatal error that cannot be recovered from
func isFatalError(err error) bool {
	errStr := err.Error()
	return strings.Contains(errStr, "authentication failed") ||
		strings.Contains(errStr, "unauthorized") ||
		strings.Contains(errStr, "invalid namespace")
}

// LogError logs an error with its category
func LogError(err error, message string) {
	category := CategorizeError(err)
	log.Printf("[ERROR] [%s] %s: %v", category, message, err)
}

// HandleError handles an error based on its category
func HandleError(err error, metrics *Metrics) (shouldRetry bool, backoffDuration time.Duration) {
	if err == nil {
		return false, 0
	}

	category := CategorizeError(err)

	switch category {
	case ErrorCategoryConnection:
		metrics.IncrementConnectionError()
		return true, time.Second * 5

	case ErrorCategoryTransient:
		metrics.IncrementConnectionError()
		return true, time.Second * 2

	case ErrorCategoryResumeToken:
		metrics.IncrementResumeError()
		// For resume token errors, we should retry without using the resume token
		return true, time.Second * 1

	case ErrorCategoryFatal:
		// Fatal errors cannot be recovered from
		log.Printf("[FATAL] Encountered unrecoverable error: %v", err)
		return false, 0

	default:
		metrics.IncrementProcessingError()
		return true, time.Second * 3
	}
}

// CalculateBackoff calculates the next backoff duration using exponential backoff
func CalculateBackoff(attempt int, config *Config) time.Duration {
	// Start with the initial backoff
	backoff := config.InitialBackoff

	// Apply exponential backoff based on the attempt number
	for i := 0; i < attempt && i < 10; i++ { // Cap at 10 to avoid overflow
		backoff = time.Duration(float64(backoff) * config.BackoffMultiplier)
	}

	// Cap at the maximum backoff
	if backoff > config.MaxBackoff {
		backoff = config.MaxBackoff
	}

	return backoff
}
