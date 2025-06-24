package logger

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"sync"
	"time"

	"github.com/ktappdev/cicd-thing/internal/config"
	"github.com/ktappdev/cicd-thing/internal/deployment"
)

// Logger handles deployment logging
type Logger struct {
	config   *config.Config
	file     *os.File
	logger   *log.Logger
	mutex    sync.Mutex
	events   chan *deployment.Event
	stopChan chan struct{}
}

// New creates a new logger instance
func New(cfg *config.Config) (*Logger, error) {
	// Create log file if it doesn't exist
	file, err := os.OpenFile(cfg.LogFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file %s: %w", cfg.LogFile, err)
	}

	// Create multi-writer to write to both file and stdout
	multiWriter := io.MultiWriter(file, os.Stdout)
	logger := log.New(multiWriter, "", 0) // No default timestamp, we'll format our own

	l := &Logger{
		config:   cfg,
		file:     file,
		logger:   logger,
		events:   make(chan *deployment.Event, 1000),
		stopChan: make(chan struct{}),
	}

	// Start event processor
	go l.processEvents()

	return l, nil
}

// LogDeploymentEvent logs a deployment event
func (l *Logger) LogDeploymentEvent(event *deployment.Event) {
	select {
	case l.events <- event:
	default:
		// Channel is full, log directly to avoid blocking
		l.logEventDirect(event)
	}
}

// LogDeploymentResult logs a deployment result
func (l *Logger) LogDeploymentResult(result *deployment.Result) {
	event := &deployment.Event{
		ID:         result.Request.ID,
		Repository: result.Request.Repository,
		Branch:     result.Request.Branch,
		Commit:     result.Request.Commit,
		Status:     result.Status,
		Timestamp:  result.EndTime,
		Duration:   result.Duration,
	}

	if result.Error != "" {
		event.Error = result.Error
		event.Message = fmt.Sprintf("Deployment failed: %s", result.Error)
	} else {
		event.Message = "Deployment completed successfully"
	}

	l.LogDeploymentEvent(event)
}

// LogWebhookReceived logs when a webhook is received
func (l *Logger) LogWebhookReceived(repository, branch, commit, author string) {
	event := &deployment.Event{
		Repository: repository,
		Branch:     branch,
		Commit:     commit,
		Status:     "WEBHOOK_RECEIVED",
		Timestamp:  time.Now(),
		Message:    fmt.Sprintf("Webhook received from %s", author),
	}
	l.LogDeploymentEvent(event)
}

// LogManualTrigger logs when a manual deployment is triggered
func (l *Logger) LogManualTrigger(repository, branch, commit string) {
	event := &deployment.Event{
		Repository: repository,
		Branch:     branch,
		Commit:     commit,
		Status:     "MANUAL_TRIGGER",
		Timestamp:  time.Now(),
		Message:    "Manual deployment triggered via API",
	}
	l.LogDeploymentEvent(event)
}

// LogError logs a general error
func (l *Logger) LogError(message string, err error) {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	
	timestamp := time.Now().Format(time.RFC3339)
	if err != nil {
		l.logger.Printf("%s | ERROR | %s: %v", timestamp, message, err)
	} else {
		l.logger.Printf("%s | ERROR | %s", timestamp, message)
	}
}

// LogInfo logs an informational message
func (l *Logger) LogInfo(message string) {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	
	timestamp := time.Now().Format(time.RFC3339)
	l.logger.Printf("%s | INFO | %s", timestamp, message)
}

// processEvents processes events from the channel
func (l *Logger) processEvents() {
	for {
		select {
		case event := <-l.events:
			l.logEventDirect(event)
		case <-l.stopChan:
			// Process remaining events
			for {
				select {
				case event := <-l.events:
					l.logEventDirect(event)
				default:
					return
				}
			}
		}
	}
}

// logEventDirect logs an event directly (thread-safe)
func (l *Logger) logEventDirect(event *deployment.Event) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	// Format: timestamp | repository | branch | commit | status | duration | message
	timestamp := event.Timestamp.Format(time.RFC3339)
	
	var durationStr string
	if event.Duration > 0 {
		durationStr = fmt.Sprintf(" | %v", event.Duration.Round(time.Millisecond))
	}

	var errorStr string
	if event.Error != "" {
		errorStr = fmt.Sprintf(" | error: %s", event.Error)
	}

	logLine := fmt.Sprintf("%s | %s | %s | %s | %s%s%s",
		timestamp,
		event.Repository,
		event.Branch,
		event.Commit,
		event.Status,
		durationStr,
		errorStr,
	)

	l.logger.Println(logLine)
}

// LogJSON logs an event in JSON format (useful for structured logging)
func (l *Logger) LogJSON(event *deployment.Event) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	jsonData, err := json.Marshal(event)
	if err != nil {
		l.logger.Printf("Failed to marshal event to JSON: %v", err)
		return
	}

	l.logger.Println(string(jsonData))
}

// Close closes the logger and cleans up resources
func (l *Logger) Close() error {
	close(l.stopChan)
	
	// Wait a bit for events to be processed
	time.Sleep(100 * time.Millisecond)
	
	if l.file != nil {
		return l.file.Close()
	}
	return nil
}

// GetLogFile returns the path to the log file
func (l *Logger) GetLogFile() string {
	return l.config.LogFile
}

// TailLogs returns the last n lines from the log file
func (l *Logger) TailLogs(n int) ([]string, error) {
	// This is a simple implementation - in production you might want to use a more efficient approach
	file, err := os.Open(l.config.LogFile)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}
	defer file.Close()

	// Read all lines (simple approach for now)
	var lines []string
	// Implementation would read file and return last n lines
	// For now, return empty slice
	return lines, nil
}
