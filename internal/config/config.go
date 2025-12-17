// Package config holds all configuration for the deployment orchestrator
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
)

// Config holds all configuration for the deployment orchestrator
type Config struct {
	// Server settings
	Port          string `toml:"port"`
	WebhookSecret string `toml:"webhook_secret"`
	APIKey        string `toml:"api_key"`

	// Endpoint paths
	WebhookPath string `toml:"webhook_path"`
	DeployPath  string `toml:"deploy_path"`
	LogsPath    string `toml:"logs_path"`
	HealthPath  string `toml:"health_path"`
	StatusPath  string `toml:"status_path"`

	// Logging
	LogFile        string `toml:"log_file"`
	MaxLogSizeMB   int    `toml:"max_log_size_mb"`  // Max log file size before rotation (MB)
	MaxRotatedLogs int    `toml:"max_rotated_logs"` // Number of rotated logs to keep

	// Configuration status (not saved to TOML)
	Configured     bool   `toml:"-"` // Whether the config has been properly set up
	ConfigFilePath string `toml:"-"` // Absolute path of the loaded config file

	// Repository mappings (repo -> local path)
	RepoMap map[string]string `toml:"repositories"`

	// Optional explicit app names per repository (repo -> app name)
	AppNames map[string]string `toml:"app_names"`

	// Commands per app
	Commands        map[string]string `toml:"commands"`
	DefaultCommands string            `toml:"default_commands"`

	// Branch filtering
	BranchFilter  string            `toml:"branch_filter"`
	BranchFilters map[string]string `toml:"branch_filters"`

	// Concurrency and timeouts
	ConcurrencyLimit int           `toml:"concurrency_limit"`
	TimeoutSeconds   int           `toml:"timeout_seconds"`
	Timeout          time.Duration `toml:"-"` // Computed field

	// Features
	DryRun bool `toml:"dry_run"`
}

// Load reads configuration from config.toml file in multiple locations
func Load() (*Config, error) {
	cfg := &Config{
		// Set defaults
		Port:             "3000",
		WebhookPath:      "/admin/webhook",
		DeployPath:       "/admin/deploy",
		LogsPath:         "/admin/logs",
		LogFile:          "./cicd-thing.log",
		MaxLogSizeMB:     10, // 10MB default
		MaxRotatedLogs:   5,  // Keep 5 rotated logs
		DefaultCommands:  "git pull && npm ci && npm run build",
		BranchFilter:     "main",
		BranchFilters:    make(map[string]string),
		ConcurrencyLimit: 2,
		TimeoutSeconds:   300,
		DryRun:           false,
	}

	// Find config file in multiple locations
	configPath, err := findConfigFile()
	if err != nil {
		return nil, err
	}

	// Decode TOML file
	if _, err := toml.DecodeFile(configPath, cfg); err != nil {
		return nil, fmt.Errorf("error loading config file %s: %w", configPath, err)
	}

	// Record which config file was loaded
	cfg.ConfigFilePath = configPath

	// Compute derived fields
	cfg.Timeout = time.Duration(cfg.TimeoutSeconds) * time.Second

	// Validate required fields
	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return cfg, nil
}

// isConfigured checks if a config file has been properly configured
func isConfigured(configPath string) (bool, error) {
	content, err := os.ReadFile(configPath)
	if err != nil {
		return false, err
	}

	// Check for the configuration marker
	return !strings.Contains(string(content), "# CONFIGURATION_NEEDED"), nil
}

// findConfigFile searches for config.toml in multiple locations
func findConfigFile() (string, error) {
	// Define search paths in order of preference
	searchPaths := []string{}

	// Add user home directory config path first (primary location)
	if homeDir, err := os.UserHomeDir(); err == nil {
		searchPaths = append(searchPaths, filepath.Join(homeDir, ".config", "cicd-thing", "config.toml"))
	}

	// System-wide config locations
	searchPaths = append(searchPaths,
		"/etc/cicd-thing/config.toml",           // System-wide config
		"/usr/local/etc/cicd-thing/config.toml", // Alternative system config
	)

	// Legacy locations (read-only fallback, no auto-creation)
	searchPaths = append(searchPaths,
		"./config.toml",        // Legacy current directory
		"./config/config.toml", // Legacy local config directory
	)

	// Search for existing config file
	for _, path := range searchPaths {
		if _, err := os.Stat(path); err == nil {
			// Check if the config file has been properly configured
			configured, err := isConfigured(path)
			if err != nil {
				return "", fmt.Errorf("failed to check config status: %w", err)
			}

			if configured {
				fmt.Printf("Found config file: %s\n", path)
				return path, nil
			} else {
				// Config file exists but is not configured
				fmt.Printf("\n=== CONFIGURATION REQUIRED ===\n")
				fmt.Printf("Configuration file found but not configured: %s\n", path)
				fmt.Printf("Please edit this file with your settings before running the application again.\n")
				fmt.Printf("IMPORTANT: Remove the line '# CONFIGURATION_NEEDED' after configuring!\n")
				fmt.Printf("Required fields to configure:\n")
				fmt.Printf("  - webhook_secret: Your GitHub webhook secret\n")
				fmt.Printf("  - api_key: Your API key for authentication\n")
				fmt.Printf("  - repositories: Map of repository names to local paths\n")
				fmt.Printf("===============================\n\n")
				return "", fmt.Errorf("configuration file at %s needs to be configured - remove '# CONFIGURATION_NEEDED' line when done", path)
			}
		}
	}

	// No config file found, create default in user config directory (~/.config/cicd-thing/config.toml)
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to determine home directory for default config: %w", err)
	}

	defaultPath := filepath.Join(homeDir, ".config", "cicd-thing", "config.toml")
	if err := createDefaultConfig(defaultPath); err != nil {
		return "", fmt.Errorf("failed to create default config: %w", err)
	}

	fmt.Printf("\n=== CONFIGURATION REQUIRED ===\n")
	fmt.Printf("A default configuration file has been created at: %s\n", defaultPath)
	fmt.Printf("Please edit this file with your settings before running the application again.\n")
	fmt.Printf("Required fields to configure:\n")
	fmt.Printf("  - webhook_secret: Your GitHub webhook secret\n")
	fmt.Printf("  - api_key: Your API key for authentication\n")
	fmt.Printf("  - repositories: Map of repository names to local paths\n")
	fmt.Printf("===============================\n\n")

	return "", fmt.Errorf("configuration file created at %s - please configure it and restart the application", defaultPath)
}

// createDefaultConfig creates a default config.toml file with example values
func createDefaultConfig(path string) error {
	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	defaultConfig := `# CICD-Thing Configuration File
# REMOVE THIS LINE AFTER CONFIGURATION: # CONFIGURATION_NEEDED
# Please configure the required settings below

# Server settings
port = "3000"
webhook_secret = "YOUR_WEBHOOK_SECRET_HERE"  # REQUIRED: Set your GitHub webhook secret
api_key = "YOUR_API_KEY_HERE"                # OPTIONAL: Set to enable /deploy endpoint for manual deployments

# Endpoint paths - Customize these for security or organization
webhook_path = "/admin/webhook"  # GitHub webhook receiver
deploy_path = "/admin/deploy"    # Manual deployment API (requires api_key)
logs_path = "/admin/logs"        # Log viewer (rate limited)
health_path = "/health"          # Health check endpoint (public)
status_path = "/status"          # Status information endpoint (public)

# Logging
log_file = "./cicd-thing.log"
max_log_size_mb = 10      # Rotate log when it reaches this size (MB)
max_rotated_logs = 5      # Keep this many rotated log files

# Default commands to run for deployments
default_commands = "git pull && npm ci && npm run build"

# Branch filtering
# Global default: only deploy from this branch if no per-app override is set
branch_filter = "main"

# Optional: per-app branch filters (keys match app names used in [commands])
[branch_filters]
# "my-app" = "main"
# "api-service" = "release"

# Performance settings
concurrency_limit = 2
timeout_seconds = 300

# Features
dry_run = false

# Repository mappings - REQUIRED
# Map repository names to local deployment paths
[repositories]
# "myorg/web" = "/var/www/web"
# "myorg/api" = "/opt/api"

# Optional: explicit app names per repository (to avoid repeating names)
[app_names]
# "myorg/web" = "web"
# "myorg/api" = "api"

# Per-application deployment commands (optional; keys are app names)
[commands]
# "web" = "git pull && npm ci && npm run build && pm2 restart web"
# "api" = "git pull && go build -o api . && systemctl restart api"

# Note: Built-in rollback configuration is planned for a future version.
# For now, implement any rollback behavior using your own scripts or deployment tooling.
`

	return os.WriteFile(path, []byte(defaultConfig), 0o644)
}

// validate checks that required configuration is present
func (c *Config) validate() error {
	if c.WebhookSecret == "" {
		return fmt.Errorf("WEBHOOK_SECRET is required")
	}
	if c.APIKey == "" {
		return fmt.Errorf("API_KEY is required")
	}
	if len(c.RepoMap) == 0 {
		return fmt.Errorf("REPO_MAP is required")
	}
	return nil
}

// GetAllowedBranch returns the allowed branch for a given app/repository.
// Resolution order:
// 1. If BranchFilters[app] is set, use that
// 2. Else if BranchFilter is set, use that
// 3. Else default to "main"
func (c *Config) GetAllowedBranch(app string) string {
	if c.BranchFilters != nil {
		if b, ok := c.BranchFilters[app]; ok && b != "" {
			return b
		}
	}
	if c.BranchFilter != "" {
		return c.BranchFilter
	}
	return "main"
}
