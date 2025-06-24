package server

import (
	"log"
	"net/http"

	"github.com/ktappdev/cicd-thing/internal/config"
	"github.com/ktappdev/cicd-thing/internal/security"
	"github.com/ktappdev/cicd-thing/internal/webhook"
)

// Server represents the HTTP server
type Server struct {
	config         *config.Config
	webhookHandler *webhook.Handler
	security       *security.Middleware
}

// New creates a new server instance
func New(cfg *config.Config) *Server {
	return &Server{
		config:         cfg,
		webhookHandler: webhook.New(cfg),
		security:       security.New(cfg),
	}
}

// Start starts the HTTP server
func (s *Server) Start() error {
	// Set up routes with security middleware
	http.HandleFunc("/webhook", s.security.IPAllowlistMiddleware(s.webhookHandler.HandleWebhook))
	http.HandleFunc("/health", s.handleHealth)
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

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"healthy","service":"cicd-thing"}`))
}

// handleManualDeploy handles manual deployment requests
func (s *Server) handleManualDeploy(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Check API key authentication
	apiKey := r.Header.Get("Authorization")
	if apiKey != "Bearer "+s.config.APIKey {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// TODO: Implement manual deployment logic
	// This will be implemented in the deployment package

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Manual deployment triggered"))
}
