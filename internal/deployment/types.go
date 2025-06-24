package deployment

import (
	"time"
)

// Status represents the deployment status
type Status string

const (
	StatusStarted   Status = "STARTED"
	StatusSuccess   Status = "SUCCESS"
	StatusFailed    Status = "FAILED"
	StatusTimeout   Status = "TIMEOUT"
	StatusRollback  Status = "ROLLBACK"
	StatusCancelled Status = "CANCELLED"
)

// Request represents a deployment request
type Request struct {
	ID         string
	Repository string
	Branch     string
	Commit     string
	Message    string
	Author     string
	Timestamp  time.Time
	LocalPath  string
	Commands   []string
	Manual     bool // true if triggered manually via API
}

// Result represents the result of a deployment
type Result struct {
	Request   *Request
	Status    Status
	StartTime time.Time
	EndTime   time.Time
	Duration  time.Duration
	Output    string
	Error     string
	ExitCode  int
}

// Event represents a deployment event for logging
type Event struct {
	ID         string
	Repository string
	Branch     string
	Commit     string
	Status     Status
	Timestamp  time.Time
	Message    string
	Error      string
	Duration   time.Duration
}

// Lock represents a deployment lock for an application
type Lock struct {
	AppName   string
	StartTime time.Time
	RequestID string
}
