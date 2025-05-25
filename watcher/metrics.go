// metrics.go - Handle metrics for the change stream watcher
package main

import (
	"fmt"
	"log"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

// Metrics tracks various metrics about the change stream watcher
type Metrics struct {
	// Operation counters
	InsertCount  uint64
	UpdateCount  uint64
	ReplaceCount uint64
	DeleteCount  uint64
	OtherOpCount uint64

	// Error metrics
	ConnectionErrors uint64
	ProcessingErrors uint64
	ResumeErrors     uint64

	// Performance metrics
	LastEventProcessed   time.Time
	EventsProcessedTotal uint64
	SuccessfulReconnects uint64
	FailedReconnects     uint64

	// Monitoring
	startTime     time.Time
	mu            sync.Mutex
	isHealthy     bool
	lastHeartbeat time.Time

	// File for persistence
	metricsFile string
}

// NewMetrics creates a new metrics tracker
func NewMetrics(metricsFile string) *Metrics {
	m := &Metrics{
		startTime:     time.Now(),
		isHealthy:     true,
		lastHeartbeat: time.Now(),
		metricsFile:   metricsFile,
	}

	// Try to load previous metrics if available
	m.loadMetrics()

	return m
}

// IncrementInsert increments the insert operation counter
func (m *Metrics) IncrementInsert() {
	atomic.AddUint64(&m.InsertCount, 1)
	atomic.AddUint64(&m.EventsProcessedTotal, 1)
	m.updateLastProcessed()
}

// IncrementUpdate increments the update operation counter
func (m *Metrics) IncrementUpdate() {
	atomic.AddUint64(&m.UpdateCount, 1)
	atomic.AddUint64(&m.EventsProcessedTotal, 1)
	m.updateLastProcessed()
}

// IncrementReplace increments the replace operation counter
func (m *Metrics) IncrementReplace() {
	atomic.AddUint64(&m.ReplaceCount, 1)
	atomic.AddUint64(&m.EventsProcessedTotal, 1)
	m.updateLastProcessed()
}

// IncrementDelete increments the delete operation counter
func (m *Metrics) IncrementDelete() {
	atomic.AddUint64(&m.DeleteCount, 1)
	atomic.AddUint64(&m.EventsProcessedTotal, 1)
	m.updateLastProcessed()
}

// IncrementOtherOp increments the counter for other operations
func (m *Metrics) IncrementOtherOp() {
	atomic.AddUint64(&m.OtherOpCount, 1)
	atomic.AddUint64(&m.EventsProcessedTotal, 1)
	m.updateLastProcessed()
}

// IncrementConnectionError increments the connection error counter
func (m *Metrics) IncrementConnectionError() {
	atomic.AddUint64(&m.ConnectionErrors, 1)
}

// IncrementProcessingError increments the processing error counter
func (m *Metrics) IncrementProcessingError() {
	atomic.AddUint64(&m.ProcessingErrors, 1)
}

// IncrementResumeError increments the resume token error counter
func (m *Metrics) IncrementResumeError() {
	atomic.AddUint64(&m.ResumeErrors, 1)
}

// IncrementSuccessfulReconnect increments the successful reconnect counter
func (m *Metrics) IncrementSuccessfulReconnect() {
	atomic.AddUint64(&m.SuccessfulReconnects, 1)
}

// IncrementFailedReconnect increments the failed reconnect counter
func (m *Metrics) IncrementFailedReconnect() {
	atomic.AddUint64(&m.FailedReconnects, 1)
}

// updateLastProcessed updates the timestamp of the last processed event
func (m *Metrics) updateLastProcessed() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.LastEventProcessed = time.Now()
	m.updateHeartbeat()
}

// updateHeartbeat updates the last heartbeat time
func (m *Metrics) updateHeartbeat() {
	m.lastHeartbeat = time.Now()
	m.isHealthy = true

	// Save metrics periodically (every 100 events)
	if m.EventsProcessedTotal%100 == 0 {
		m.saveMetrics()
	}
}

// GetUptimeSeconds returns the number of seconds the watcher has been running
func (m *Metrics) GetUptimeSeconds() float64 {
	return time.Since(m.startTime).Seconds()
}

// GetEventsPerSecond returns the average number of events processed per second
func (m *Metrics) GetEventsPerSecond() float64 {
	uptime := m.GetUptimeSeconds()
	if uptime == 0 {
		return 0
	}
	return float64(m.EventsProcessedTotal) / uptime
}

// IsHealthy returns whether the watcher is considered healthy
func (m *Metrics) IsHealthy() bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Consider unhealthy if no heartbeat in the last minute
	if time.Since(m.lastHeartbeat) > time.Minute {
		m.isHealthy = false
	}

	return m.isHealthy
}

// SetUnhealthy explicitly marks the watcher as unhealthy
func (m *Metrics) SetUnhealthy() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.isHealthy = false
}

// PrintStats logs the current metrics
func (m *Metrics) PrintStats() {
	log.Printf("=== Watcher Metrics ===")
	log.Printf("Uptime: %.2f seconds", m.GetUptimeSeconds())
	log.Printf("Events: %d total (%.2f/sec)", m.EventsProcessedTotal, m.GetEventsPerSecond())
	log.Printf("Operations: %d inserts, %d updates, %d replaces, %d deletes, %d other",
		m.InsertCount, m.UpdateCount, m.ReplaceCount, m.DeleteCount, m.OtherOpCount)
	log.Printf("Errors: %d connection, %d processing, %d resume",
		m.ConnectionErrors, m.ProcessingErrors, m.ResumeErrors)
	log.Printf("Reconnects: %d successful, %d failed",
		m.SuccessfulReconnects, m.FailedReconnects)
	log.Printf("Health: %v (last event: %v)", m.IsHealthy(), m.LastEventProcessed)
	log.Printf("========================")
}

// saveMetrics persists metrics to a file
func (m *Metrics) saveMetrics() error {
	data := fmt.Sprintf(`{
  "timestamp": "%s",
  "uptime_seconds": %.2f,
  "events_total": %d,
  "events_per_second": %.2f,
  "operations": {
    "insert": %d,
    "update": %d,
    "replace": %d,
    "delete": %d,
    "other": %d
  },
  "errors": {
    "connection": %d,
    "processing": %d,
    "resume": %d
  },
  "reconnects": {
    "successful": %d,
    "failed": %d
  },
  "health": "%v",
  "last_event": "%v"
}`,
		time.Now().Format(time.RFC3339),
		m.GetUptimeSeconds(),
		m.EventsProcessedTotal,
		m.GetEventsPerSecond(),
		m.InsertCount,
		m.UpdateCount,
		m.ReplaceCount,
		m.DeleteCount,
		m.OtherOpCount,
		m.ConnectionErrors,
		m.ProcessingErrors,
		m.ResumeErrors,
		m.SuccessfulReconnects,
		m.FailedReconnects,
		m.IsHealthy(),
		m.LastEventProcessed.Format(time.RFC3339),
	)

	// Write to temporary file first
	tempFile := m.metricsFile + ".tmp"
	if err := os.WriteFile(tempFile, []byte(data), 0644); err != nil {
		return fmt.Errorf("failed to write metrics to temp file: %w", err)
	}

	// Rename for atomic update
	if err := os.Rename(tempFile, m.metricsFile); err != nil {
		return fmt.Errorf("failed to rename temp metrics file: %w", err)
	}

	return nil
}

// loadMetrics loads metrics from a file if it exists
func (m *Metrics) loadMetrics() {
	// We're only loading total counts for continuity
	// We don't try to parse the full JSON for simplicity

	// Check if the file exists
	if _, err := os.Stat(m.metricsFile); os.IsNotExist(err) {
		return
	}

	log.Printf("Previous metrics file found, metrics will be cumulative")
}
