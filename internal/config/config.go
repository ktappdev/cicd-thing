package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

// Config holds all configuration for the deployment orchestrator
type Config struct {
	// Server settings
	Port         string
	WebhookSecret string
	APIKey       string

	// Logging
	LogFile string

	// Repository mappings (repo -> local path)
	RepoMap map[string]string

	// Commands per app
	Commands        map[string]string
	DefaultCommands string

	// Rollback commands per app
	RollbackCommands map[string]string

	// Branch filtering
	BranchFilter string

	// Concurrency and timeouts
	ConcurrencyLimit int
	TimeoutSeconds   int
	Timeout          time.Duration

	// Notifications
	NotifyOnRollback bool

	// Security
	IPAllowlist []string

	// Features
	DryRun bool
}

// Load reads configuration from environment variables and .env file
func Load() (*Config, error) {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		// .env file is optional, so we don't fail if it doesn't exist
		if !os.IsNotExist(err) {
			return nil, fmt.Errorf("error loading .env file: %w", err)
		}
	}

	cfg := &Config{
		Port:             getEnv("PORT", "3000"),
		WebhookSecret:    getEnv("WEBHOOK_SECRET", ""),
		APIKey:           getEnv("API_KEY", ""),
		LogFile:          getEnv("LOG_FILE", "/var/log/deployer.log"),
		DefaultCommands:  getEnv("DEFAULT_COMMANDS", "git pull && npm ci && npm run build"),
		BranchFilter:     getEnv("BRANCH_FILTER", "main"),
		ConcurrencyLimit: getEnvInt("CONCURRENCY_LIMIT", 2),
		TimeoutSeconds:   getEnvInt("TIMEOUT_SECONDS", 300),
		NotifyOnRollback: getEnvBool("NOTIFY_ON_ROLLBACK", false),
		DryRun:           getEnvBool("DRY_RUN", false),
	}

	cfg.Timeout = time.Duration(cfg.TimeoutSeconds) * time.Second

	// Parse repository mappings
	cfg.RepoMap = parseRepoMap(getEnv("REPO_MAP", ""))

	// Parse per-app commands
	cfg.Commands = parseCommands("COMMANDS_")

	// Parse rollback commands
	cfg.RollbackCommands = parseCommands("ROLLBACK_COMMANDS_")

	// Parse IP allowlist
	cfg.IPAllowlist = parseIPAllowlist(getEnv("IP_ALLOWLIST", ""))

	// Validate required fields
	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return cfg, nil
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

// getEnv gets an environment variable with a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvInt gets an environment variable as an integer with a default value
func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// getEnvBool gets an environment variable as a boolean with a default value
func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

// parseRepoMap parses the REPO_MAP environment variable
// Format: "repo1:path1,repo2:path2"
func parseRepoMap(repoMapStr string) map[string]string {
	repoMap := make(map[string]string)
	if repoMapStr == "" {
		return repoMap
	}

	pairs := strings.Split(repoMapStr, ",")
	for _, pair := range pairs {
		parts := strings.SplitN(strings.TrimSpace(pair), ":", 2)
		if len(parts) == 2 {
			repo := strings.TrimSpace(parts[0])
			path := strings.TrimSpace(parts[1])
			if repo != "" && path != "" {
				repoMap[repo] = path
			}
		}
	}
	return repoMap
}

// parseCommands parses environment variables with a given prefix
// For example, COMMANDS_Hello-World=git pull && npm build
func parseCommands(prefix string) map[string]string {
	commands := make(map[string]string)
	for _, env := range os.Environ() {
		if strings.HasPrefix(env, prefix) {
			parts := strings.SplitN(env, "=", 2)
			if len(parts) == 2 {
				key := strings.TrimPrefix(parts[0], prefix)
				value := parts[1]
				if key != "" && value != "" {
					commands[key] = value
				}
			}
		}
	}
	return commands
}

// parseIPAllowlist parses the IP_ALLOWLIST environment variable
// Format: "ip1,ip2,ip3"
func parseIPAllowlist(allowlistStr string) []string {
	if allowlistStr == "" {
		return nil
	}

	ips := strings.Split(allowlistStr, ",")
	var result []string
	for _, ip := range ips {
		if trimmed := strings.TrimSpace(ip); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}
