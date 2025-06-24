package server

import (
	"bufio"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

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
	http.HandleFunc("/logs", s.security.IPAllowlistMiddleware(s.security.AuthMiddleware(s.handleLogs)))

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

// handleLogs handles log viewer requests
func (s *Server) handleLogs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse query parameters
	limitStr := r.URL.Query().Get("limit")
	if limitStr == "" {
		limitStr = "50" // Default to 50 lines
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 || limit > 1000 {
		limit = 50 // Default fallback
	}

	// Read log lines
	logLines, err := s.readLogLines(limit)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to read logs: %v", err), http.StatusInternalServerError)
		return
	}

	// Serve HTML page
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)

	// Render HTML template
	tmpl := template.Must(template.New("logs").Parse(logViewerHTML))
	data := struct {
		Logs  []string
		Limit int
	}{
		Logs:  logLines,
		Limit: limit,
	}

	if err := tmpl.Execute(w, data); err != nil {
		log.Printf("Failed to render log viewer template: %v", err)
	}
}

// readLogLines reads the last n lines from the log file and adds project prefixes
func (s *Server) readLogLines(n int) ([]string, error) {
	logFile := s.logger.GetLogFile()
	file, err := os.Open(logFile)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}
	defer file.Close()

	// Read all lines into memory (simple approach)
	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		originalLine := scanner.Text()
		prefixedLine := s.addProjectPrefix(originalLine)
		lines = append(lines, prefixedLine)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to read log file: %w", err)
	}

	// Return last n lines
	if len(lines) <= n {
		return lines, nil
	}
	return lines[len(lines)-n:], nil
}

// addProjectPrefix adds a project prefix to log lines
func (s *Server) addProjectPrefix(line string) string {
	// Parse log line format: timestamp | repository | branch | commit | status | duration | error
	// OR: timestamp | LEVEL | message
	parts := strings.Split(line, " | ")
	if len(parts) < 3 {
		return "[UNKNOWN] " + line
	}

	// Check if it's a system log (format: timestamp | LEVEL | message)
	if parts[1] == "INFO" || parts[1] == "ERROR" || parts[1] == "WARN" || parts[1] == "DEBUG" {
		return "[SYSTEM] " + line
	}

	// It's a deployment log, use repository name as project
	repository := parts[1]
	if repository == "" {
		return "[UNKNOWN] " + line
	}

	return "[" + repository + "] " + line
}

// logViewerHTML contains the HTML template for the log viewer
const logViewerHTML = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>CI/CD Thing - Log Viewer</title>
    <style>
        body {
            font-family: 'Courier New', monospace;
            margin: 0;
            padding: 20px;
            background-color: #1e1e1e;
            color: #d4d4d4;
        }
        .header {
            display: flex;
            justify-content: space-between;
            align-items: center;
            margin-bottom: 20px;
            padding: 10px;
            background-color: #2d2d30;
            border-radius: 5px;
        }
        .controls {
            display: flex;
            gap: 10px;
            align-items: center;
        }
        select, button {
            padding: 8px 12px;
            border: 1px solid #3e3e42;
            background-color: #2d2d30;
            color: #d4d4d4;
            border-radius: 3px;
            cursor: pointer;
        }
        button:hover {
            background-color: #3e3e42;
        }
        .log-container {
            background-color: #0d1117;
            border: 1px solid #30363d;
            border-radius: 5px;
            padding: 15px;
            max-height: 80vh;
            overflow-y: auto;
            font-size: 14px;
            line-height: 1.4;
        }
        .log-line {
            margin: 2px 0;
            white-space: pre-wrap;
            word-break: break-all;
        }
        .log-line:hover {
            background-color: #21262d;
        }
        .timestamp {
            color: #7c3aed;
        }
        .error {
            color: #f87171;
        }
        .success {
            color: #34d399;
        }
        .info {
            color: #60a5fa;
        }
        .warning {
            color: #fbbf24;
        }
    </style>
</head>
<body>
    <div class="header">
        <h1>CI/CD Thing - Log Viewer</h1>
        <div class="controls">
            <label for="limit">Show:</label>
            <select id="limit" onchange="updateLimit()">
                <option value="10" {{if eq .Limit 10}}selected{{end}}>10 lines</option>
                <option value="20" {{if eq .Limit 20}}selected{{end}}>20 lines</option>
                <option value="50" {{if eq .Limit 50}}selected{{end}}>50 lines</option>
                <option value="100" {{if eq .Limit 100}}selected{{end}}>100 lines</option>
                <option value="200" {{if eq .Limit 200}}selected{{end}}>200 lines</option>
            </select>
            <button onclick="refreshLogs()">🔄 Refresh</button>
        </div>
    </div>
    
    <div class="log-container">
        {{range .Logs}}
        <div class="log-line">{{.}}</div>
        {{else}}
        <div class="log-line">No logs available</div>
        {{end}}
    </div>

    <script>
        function refreshLogs() {
            window.location.reload();
        }
        
        function updateLimit() {
            const limit = document.getElementById('limit').value;
            const url = new URL(window.location);
            url.searchParams.set('limit', limit);
            window.location.href = url.toString();
        }
        
        // Auto-refresh every 30 seconds
        setInterval(refreshLogs, 30000);
    </script>
</body>
</html>
`
