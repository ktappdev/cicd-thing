package deployment

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/ktappdev/cicd-thing/internal/config"
	"github.com/ktappdev/cicd-thing/internal/mapping"
)

// Executor handles deployment execution
type Executor struct {
	config    *config.Config
	mapper    *mapping.Mapper
	locks     map[string]*Lock
	lockMutex sync.RWMutex
	queue     chan *Request
	results   chan *Result
}

// New creates a new deployment executor
func New(cfg *config.Config) *Executor {
	executor := &Executor{
		config:  cfg,
		mapper:  mapping.New(cfg),
		locks:   make(map[string]*Lock),
		queue:   make(chan *Request, 100), // Buffer for queued deployments
		results: make(chan *Result, 100),  // Buffer for results
	}

	// Start worker goroutines
	for i := 0; i < cfg.ConcurrencyLimit; i++ {
		go executor.worker()
	}

	return executor
}

// Deploy queues a deployment request
func (e *Executor) Deploy(req *Request) error {
	appName := e.mapper.GetAppName(req.Repository)

	// Check if app is locked
	if e.isLocked(appName) {
		return fmt.Errorf("deployment already in progress for app %s", appName)
	}

	// Generate unique ID if not provided
	if req.ID == "" {
		req.ID = generateID()
	}

	// Prepare commands
	if err := e.prepareCommands(req); err != nil {
		return fmt.Errorf("failed to prepare commands: %w", err)
	}

	// Queue the deployment
	select {
	case e.queue <- req:
		return nil
	default:
		return fmt.Errorf("deployment queue is full")
	}
}

// GetResults returns the results channel
func (e *Executor) GetResults() <-chan *Result {
	return e.results
}

// worker processes deployment requests from the queue
func (e *Executor) worker() {
	for req := range e.queue {
		result := e.executeDeployment(req)
		
		// Send result to results channel
		select {
		case e.results <- result:
		default:
			// Results channel is full, log error
			fmt.Printf("Results channel full, dropping result for %s\n", req.ID)
		}
	}
}

// executeDeployment executes a single deployment
func (e *Executor) executeDeployment(req *Request) *Result {
	appName := e.mapper.GetAppName(req.Repository)
	
	// Acquire lock
	if !e.acquireLock(appName, req.ID) {
		return &Result{
			Request:   req,
			Status:    StatusFailed,
			StartTime: time.Now(),
			EndTime:   time.Now(),
			Error:     "Failed to acquire deployment lock",
		}
	}
	defer e.releaseLock(appName)

	result := &Result{
		Request:   req,
		Status:    StatusStarted,
		StartTime: time.Now(),
	}

	// Execute deployment with timeout
	ctx, cancel := context.WithTimeout(context.Background(), e.config.Timeout)
	defer cancel()

	if e.config.DryRun {
		result.Status = StatusSuccess
		result.Output = "DRY RUN: Commands would be executed"
	} else {
		result = e.runCommands(ctx, req, result)
	}

	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime)

	return result
}

// runCommands executes the deployment commands
func (e *Executor) runCommands(ctx context.Context, req *Request, result *Result) *Result {
	var output strings.Builder
	
	for i, command := range req.Commands {
		select {
		case <-ctx.Done():
			result.Status = StatusTimeout
			result.Error = "Deployment timed out"
			return result
		default:
		}

		// Execute command
		cmd := exec.CommandContext(ctx, "sh", "-c", command)
		cmd.Dir = req.LocalPath
		cmd.Env = os.Environ()

		cmdOutput, err := cmd.CombinedOutput()
		output.WriteString(fmt.Sprintf("Command %d: %s\n", i+1, command))
		output.WriteString(string(cmdOutput))
		output.WriteString("\n")

		if err != nil {
			result.Status = StatusFailed
			result.Error = fmt.Sprintf("Command failed: %s - %v", command, err)
			result.ExitCode = cmd.ProcessState.ExitCode()
			result.Output = output.String()
			
			// Attempt rollback if configured
			if e.shouldRollback(req.Repository) {
				e.performRollback(req, result)
			}
			
			return result
		}
	}

	result.Status = StatusSuccess
	result.Output = output.String()
	return result
}

// prepareCommands prepares the commands for deployment
func (e *Executor) prepareCommands(req *Request) error {
	appName := e.mapper.GetAppName(req.Repository)
	
	// Check for app-specific commands
	if commands, exists := e.config.Commands[appName]; exists {
		req.Commands = parseCommands(commands)
	} else if e.config.DefaultCommands != "" {
		// Use default commands and replace placeholder
		defaultCmd := strings.ReplaceAll(e.config.DefaultCommands, "appname", appName)
		req.Commands = parseCommands(defaultCmd)
	} else {
		return fmt.Errorf("no commands configured for app %s", appName)
	}

	return nil
}

// parseCommands splits a command string into individual commands
func parseCommands(commandStr string) []string {
	// Split by && for now, could be enhanced to handle more complex cases
	commands := strings.Split(commandStr, "&&")
	var result []string
	for _, cmd := range commands {
		if trimmed := strings.TrimSpace(cmd); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

// isLocked checks if an app is currently locked
func (e *Executor) isLocked(appName string) bool {
	e.lockMutex.RLock()
	defer e.lockMutex.RUnlock()
	_, exists := e.locks[appName]
	return exists
}

// acquireLock attempts to acquire a lock for an app
func (e *Executor) acquireLock(appName, requestID string) bool {
	e.lockMutex.Lock()
	defer e.lockMutex.Unlock()
	
	if _, exists := e.locks[appName]; exists {
		return false
	}
	
	e.locks[appName] = &Lock{
		AppName:   appName,
		StartTime: time.Now(),
		RequestID: requestID,
	}
	return true
}

// releaseLock releases a lock for an app
func (e *Executor) releaseLock(appName string) {
	e.lockMutex.Lock()
	defer e.lockMutex.Unlock()
	delete(e.locks, appName)
}

// shouldRollback checks if rollback should be performed for a repository
func (e *Executor) shouldRollback(repository string) bool {
	appName := e.mapper.GetAppName(repository)
	_, exists := e.config.RollbackCommands[appName]
	return exists
}

// performRollback performs rollback for a failed deployment
func (e *Executor) performRollback(req *Request, result *Result) {
	appName := e.mapper.GetAppName(req.Repository)
	rollbackCmd, exists := e.config.RollbackCommands[appName]
	if !exists {
		return
	}

	// Execute rollback command
	ctx, cancel := context.WithTimeout(context.Background(), e.config.Timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "sh", "-c", rollbackCmd)
	cmd.Dir = req.LocalPath
	cmd.Env = os.Environ()

	rollbackOutput, err := cmd.CombinedOutput()
	if err != nil {
		result.Error += fmt.Sprintf("\nRollback failed: %v\nRollback output: %s", err, string(rollbackOutput))
	} else {
		result.Status = StatusRollback
		result.Output += fmt.Sprintf("\nRollback executed successfully:\n%s", string(rollbackOutput))
	}
}

// generateID generates a unique ID for deployment requests
func generateID() string {
	return fmt.Sprintf("deploy_%d", time.Now().UnixNano())
}
