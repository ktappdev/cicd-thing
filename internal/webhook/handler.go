package webhook

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/ktappdev/cicd-thing/internal/config"
	"github.com/ktappdev/cicd-thing/internal/deployment"
	"github.com/ktappdev/cicd-thing/internal/mapping"
)

// Handler handles GitHub webhook requests
type Handler struct {
	config   *config.Config
	mapper   *mapping.Mapper
	executor *deployment.Executor
}

// New creates a new webhook handler
func New(cfg *config.Config, executor *deployment.Executor) *Handler {
	return &Handler{
		config:   cfg,
		mapper:   mapping.New(cfg),
		executor: executor,
	}
}

// HandleWebhook processes incoming GitHub webhook requests
func (h *Handler) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Read the request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Verify the webhook signature
	if !h.verifySignature(r.Header.Get("X-Hub-Signature-256"), body) {
		http.Error(w, "Invalid signature", http.StatusUnauthorized)
		return
	}

	// Check event type
	eventType := r.Header.Get("X-GitHub-Event")
	if eventType != "push" {
		// We only handle push events for now
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Event type not supported"))
		return
	}

	// Parse the webhook payload
	var payload GitHubWebhookPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		http.Error(w, "Failed to parse webhook payload", http.StatusBadRequest)
		return
	}

	// Process the webhook
	deploymentReq, err := h.processWebhook(&payload)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to process webhook: %v", err), http.StatusBadRequest)
		return
	}

	if deploymentReq == nil {
		// No deployment needed (e.g., wrong branch)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("No deployment triggered"))
		return
	}

	// Convert to deployment request and trigger deployment
	depReq := &deployment.Request{
		Repository: deploymentReq.Repository,
		Branch:     deploymentReq.Branch,
		Commit:     deploymentReq.Commit,
		Message:    deploymentReq.Message,
		Author:     deploymentReq.Author,
		Timestamp:  deploymentReq.Timestamp,
		LocalPath:  deploymentReq.LocalPath,
		Manual:     false,
	}

	// Trigger deployment
	if err := h.executor.Deploy(depReq); err != nil {
		http.Error(w, fmt.Sprintf("Failed to trigger deployment: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Deployment triggered successfully"))
}

// verifySignature verifies the GitHub webhook signature
func (h *Handler) verifySignature(signature string, body []byte) bool {
	if signature == "" || h.config.WebhookSecret == "" {
		return false
	}

	// Remove the "sha256=" prefix
	if !strings.HasPrefix(signature, "sha256=") {
		return false
	}
	signature = strings.TrimPrefix(signature, "sha256=")

	// Calculate the expected signature
	mac := hmac.New(sha256.New, []byte(h.config.WebhookSecret))
	mac.Write(body)
	expectedSignature := hex.EncodeToString(mac.Sum(nil))

	// Compare signatures
	return hmac.Equal([]byte(signature), []byte(expectedSignature))
}

// processWebhook processes the webhook payload and returns a deployment request
func (h *Handler) processWebhook(payload *GitHubWebhookPayload) (*DeploymentRequest, error) {
	// Extract branch name from ref (refs/heads/main -> main)
	branch := strings.TrimPrefix(payload.Ref, "refs/heads/")

	// Check if we should deploy this branch
	if h.config.BranchFilter != "" && branch != h.config.BranchFilter {
		return nil, nil // No deployment needed
	}

	// Get local path using mapper
	localPath, err := h.mapper.GetLocalPath(payload.Repository.FullName)
	if err != nil {
		return nil, err
	}

	// Create deployment request
	deploymentReq := &DeploymentRequest{
		Repository: payload.Repository.FullName,
		Branch:     branch,
		Commit:     payload.After,
		Message:    payload.HeadCommit.Message,
		Author:     payload.HeadCommit.Author.Name,
		Timestamp:  payload.HeadCommit.Timestamp,
		LocalPath:  localPath,
	}

	return deploymentReq, nil
}

// GetDeploymentRequest extracts deployment information from a webhook payload
func (h *Handler) GetDeploymentRequest(payload *GitHubWebhookPayload) (*DeploymentRequest, error) {
	return h.processWebhook(payload)
}
