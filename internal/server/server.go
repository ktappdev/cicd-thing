package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/ktappdev/cicd-thing/internal/config"
	"github.com/ktappdev/cicd-thing/internal/deployment"
	"github.com/ktappdev/cicd-thing/internal/logger"
	"github.com/ktappdev/cicd-thing/internal/security"
	"github.com/ktappdev/cicd-thing/internal/webhook"
)

// Server represents the HTTP server
type Server struct {
	config         *config.Config
	webhookHandler *webhook.Handler
	security       *security.Middleware
	executor       *deployment.Executor
	logger         *logger.Logger
}

// New creates a new server instance
func New(cfg *config.Config, executor *deployment.Executor, logger *logger.Logger) *Server {
	return &Server{
		config:         cfg,
		webhookHandler: webhook.New(cfg, executor),
		security:       security.New(cfg),
		executor:       executor,
		logger:         logger,
	}
}

// Start starts the HTTP server
func (s *Server) Start() error {
	// Set up routes with security middleware
	http.HandleFunc("/webhook", s.security.IPAllowlistMiddleware(s.webhookHandler.HandleWebhook))
	http.HandleFunc("/health", s.handleHealth)
	http.HandleFunc("/status", s.handleStatus)
	http.HandleFunc("/deploy", s.security.IPAllowlistMiddleware(s.security.AuthMiddleware(s.handleManualDeploy)))

	// Start server
	addr := ":" + s.config.Port
	log.Printf("Starting server on port %s", s.config.Port)
	return http.ListenAndServe(addr, nil)
}

// handleHealth handles health check requests
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get system information
	health := map[string]interface{}{
		"status":  "healthy",
		"service": "cicd-thing",
		"version": "1.0.0",
		"uptime":  "running", // Could be enhanced with actual uptime
		"config": map[string]interface{}{
			"port":              s.config.Port,
			"concurrency_limit": s.config.ConcurrencyLimit,
			"timeout_seconds":   s.config.TimeoutSeconds,
			"branch_filter":     s.config.BranchFilter,
			"dry_run":           s.config.DryRun,
			"repositories":      len(s.config.RepoMap),
		},
		"features": map[string]bool{
			"webhook_listener":  true,
			"manual_deployment": true,
			"rollback_support":  len(s.config.RollbackCommands) > 0,
			"ip_allowlist":      len(s.config.IPAllowlist) > 0,
			"notifications":     s.config.NotifyOnRollback,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	// Convert to JSON
	jsonData, err := json.Marshal(health)
	if err != nil {
		http.Error(w, "Failed to generate health response", http.StatusInternalServerError)
		return
	}

	w.Write(jsonData)
}

// handleStatus handles status requests showing current deployments
func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get deployment status information
	status := map[string]interface{}{
		"service": "cicd-thing",
		"status":  "running",
		"deployments": map[string]interface{}{
			"active":    0, // Could be enhanced to show actual active deployments
			"queued":    0, // Could be enhanced to show queue length
			"completed": 0, // Could be enhanced to show completed count
		},
		"repositories": s.config.RepoMap,
		"configuration": map[string]interface{}{
			"concurrency_limit": s.config.ConcurrencyLimit,
			"timeout_seconds":   s.config.TimeoutSeconds,
			"branch_filter":     s.config.BranchFilter,
			"dry_run":           s.config.DryRun,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	jsonData, err := json.Marshal(status)
	if err != nil {
		http.Error(w, "Failed to generate status response", http.StatusInternalServerError)
		return
	}

	w.Write(jsonData)
}

// handleManualDeploy handles manual deployment requests
func (s *Server) handleManualDeploy(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse query parameters
	repo := r.URL.Query().Get("repo")
	branch := r.URL.Query().Get("branch")
	commit := r.URL.Query().Get("commit")

	if repo == "" {
		http.Error(w, "Missing required parameter: repo", http.StatusBadRequest)
		return
	}

	// Set defaults
	if branch == "" {
		branch = s.config.BranchFilter
		if branch == "" {
			branch = "main"
		}
	}
	if commit == "" {
		commit = "HEAD"
	}

	// Get local path for repository
	localPath, err := s.executor.GetLocalPath(repo)
	if err != nil {
		http.Error(w, fmt.Sprintf("Repository not configured: %v", err), http.StatusBadRequest)
		return
	}

	// Create deployment request
	depReq := &deployment.Request{
		Repository: repo,
		Branch:     branch,
		Commit:     commit,
		Message:    "Manual deployment via API",
		Author:     "API",
		LocalPath:  localPath,
		Manual:     true,
	}

	// Log manual trigger
	s.logger.LogManualTrigger(repo, branch, commit)

	// Trigger deployment
	if err := s.executor.Deploy(depReq); err != nil {
		http.Error(w, fmt.Sprintf("Failed to trigger deployment: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	response := fmt.Sprintf(`{"status":"success","message":"Manual deployment triggered","repository":"%s","branch":"%s","commit":"%s"}`, repo, branch, commit)
	w.Write([]byte(response))
}
