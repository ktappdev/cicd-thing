package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/BurntSushi/toml"
)

// Config holds all configuration for the deployment orchestrator
type Config struct {
	// Server settings
	Port         string `toml:"port"`
	WebhookSecret string `toml:"webhook_secret"`
	APIKey       string `toml:"api_key"`

	// Logging
	LogFile string `toml:"log_file"`

	// Repository mappings (repo -> local path)
	RepoMap map[string]string `toml:"repositories"`

	// Commands per app
	Commands        map[string]string `toml:"commands"`
	DefaultCommands string            `toml:"default_commands"`

	// Rollback commands per app
	RollbackCommands map[string]string `toml:"rollback_commands"`

	// Branch filtering
	BranchFilter string `toml:"branch_filter"`

	// Concurrency and timeouts
	ConcurrencyLimit int           `toml:"concurrency_limit"`
	TimeoutSeconds   int           `toml:"timeout_seconds"`
	Timeout          time.Duration `toml:"-"` // Computed field

	// Notifications
	NotifyOnRollback bool `toml:"notify_on_rollback"`

	// Security
	IPAllowlist []string `toml:"ip_allowlist"`

	// Features
	DryRun bool `toml:"dry_run"`
}

// Load reads configuration from config.toml file in multiple locations
func Load() (*Config, error) {
	cfg := &Config{
		// Set defaults
		Port:             "3000",
		LogFile:          "./deployer.log",
		DefaultCommands:  "git pull && npm ci && npm run build",
		BranchFilter:     "main",
		ConcurrencyLimit: 2,
		TimeoutSeconds:   300,
		NotifyOnRollback: false,
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

	// Compute derived fields
	cfg.Timeout = time.Duration(cfg.TimeoutSeconds) * time.Second

	// Validate required fields
	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return cfg, nil
}

// findConfigFile searches for config.toml in multiple locations
func findConfigFile() (string, error) {
	// Define search paths in order of preference
	searchPaths := []string{
		"./config.toml",                    // Current directory
		"./config/config.toml",             // Local config directory
		"/etc/cicd-thing/config.toml",       // System-wide config
		"/usr/local/etc/cicd-thing/config.toml", // Alternative system config
	}

	// Add user home directory path
	if homeDir, err := os.UserHomeDir(); err == nil {
		searchPaths = append(searchPaths, filepath.Join(homeDir, ".config", "cicd-thing", "config.toml"))
	}

	// Search for existing config file
	for _, path := range searchPaths {
		if _, err := os.Stat(path); err == nil {
			fmt.Printf("Found config file: %s\n", path)
			return path, nil
		}
	}

	// No config file found, create default in current directory
	defaultPath := "./config.toml"
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
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	defaultConfig := `# CICD-Thing Configuration File
# Please configure the required settings below

# Server settings
port = "3000"
webhook_secret = "YOUR_WEBHOOK_SECRET_HERE"  # REQUIRED: Set your GitHub webhook secret
api_key = "YOUR_API_KEY_HERE"                # REQUIRED: Set your API key

# Logging
log_file = "./deployer.log"

# Default commands to run for deployments
default_commands = "git pull && npm ci && npm run build"

# Branch filtering (only deploy from this branch)
branch_filter = "main"

# Performance settings
concurrency_limit = 2
timeout_seconds = 300

# Notifications
notify_on_rollback = false

# Features
dry_run = false

# Security (optional)
# ip_allowlist = ["192.168.1.0/24", "10.0.0.0/8"]

# Repository mappings - REQUIRED
# Map repository names to local deployment paths
[repositories]
# "my-app" = "/var/www/my-app"
# "api-service" = "/opt/api-service"

# Per-application deployment commands (optional)
[commands]
# "my-app" = "git pull && npm ci && npm run build && pm2 restart my-app"
# "api-service" = "git pull && go build && systemctl restart api-service"

# Per-application rollback commands (optional)
[rollback_commands]
# "my-app" = "git checkout HEAD~1 && npm ci && npm run build && pm2 restart my-app"
# "api-service" = "git checkout HEAD~1 && go build && systemctl restart api-service"
`

	return os.WriteFile(path, []byte(defaultConfig), 0644)
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
