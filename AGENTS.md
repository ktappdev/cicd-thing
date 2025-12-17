# AGENTS.md

This file provides guidance for AI agents working with the CICD-Thing codebase.

## Project Overview

CICD-Thing is a Go-based deployment orchestrator that automatically deploys applications when code is pushed to GitHub. It receives webhook notifications from GitHub, downloads the latest code, runs build commands, and restarts applications.

## Essential Commands

### Build and Run
```bash
# Install dependencies
go mod tidy

# Method 1: Install globally with go install (recommended)
go install github.com/ktappdev/cicd-thing@latest

# Method 2: Build locally
go build -o cicd-thing .

# Run the application
./cicd-thing
# or if installed with go install:
cicd-thing
```

### Testing
```bash
# Run all tests
go test ./...

# Run tests with verbose output
go test -v ./...

# Run tests for specific package
go test ./internal/config
```

### Development Workflow
```bash
# Before making changes, ensure dependencies are current
go mod tidy

# After changes, run tests to verify
go test ./...

# Build to ensure compilation works
go build -o cicd-thing .
```

## Code Organization and Structure

### Project Layout
```
main.go                     # Application entry point
internal/                   # Private application code
├── config/                # Configuration management (TOML-based)
├── deployment/            # Deployment execution logic
├── logger/                # Structured logging with project identification
├── mapping/               # Repository to path mapping
├── security/              # Authentication and middleware
├── server/                # HTTP server and endpoints
└── webhook/               # GitHub webhook processing
```

### Module Structure
- Module name: `github.com/ktappdev/cicd-thing`
- Go version: 1.24.4
- Only external dependency: `github.com/BurntSushi/toml`

## Naming Conventions and Style Patterns

### Go Code Style
- Standard Go conventions: camelCase for variables, PascalCase for exported types
- Package names are lowercase, single words (config, server, webhook, etc.)
- Interface names typically end in -er (Handler, Executor, Mapper)
- Error handling with explicit error returns
- Context passing for operations that may timeout or cancel

### Configuration
- TOML-based configuration with clear section headers
- Snake_case for configuration keys
- Required fields: webhook_secret, api_key, repositories
- Optional fields have sensible defaults

### Logging Patterns
- Project prefixes in logs: `[project-name]`, `[SYSTEM]`, `[UNKNOWN]`
- Structured logging with consistent timestamp format
- Separate methods for different log levels: LogInfo, LogError, LogDeploymentResult

## Testing Approach and Patterns

### Test Organization
- Tests located alongside source files (`*_test.go`)
- Package-level tests for functionality
- Integration tests for webhook processing and deployment flow

### Test Patterns
- Table-driven tests for multiple scenarios
- Mock configuration for isolated testing
- Error path testing alongside success cases

### Running Tests
- Always run `go test ./...` after changes
- Use `-v` flag for verbose output during development
- Target specific packages with `go test ./internal/packagename`

## Important Gotchas and Non-Obvious Patterns

### Configuration System
- Config files searched in specific order, with user config taking precedence
- Creates default config automatically if none exists
- Contains configuration marker (`# CONFIGURATION_NEEDED`) that must be removed
- Application will refuse to start until required fields are configured

### Deployment Model
- Repository-to-path mapping through configuration
- Per-app command customization with fallback to defaults
- Branch filtering with global defaults and per-app overrides
- Concurrency limits to prevent resource exhaustion
- Deployment locking to prevent concurrent deployments of same app

### Security Model
- GitHub webhook signature verification using HMAC-SHA256
- API key authentication for manual deployments
- Rate limiting on log endpoint (30 requests/minute)
- Command execution timeouts to prevent hanging

### Error Handling Patterns
- Structured error messages with context
- Graceful degradation where possible
- Detailed logging for troubleshooting
- Early validation to fail fast on configuration issues

### Logging System
- Multi-writer approach (file + stdout)
- Project identification through prefixes
- Real-time web log viewer with auto-refresh
- Log rotation based on size limits

## Development Context

### Key Architectural Decisions
- Standard library only for core functionality (except TOML parsing)
- Goroutine-based workers for concurrent deployment processing
- Channel-based communication between components
- Separation of concerns with distinct packages for each responsibility

### Common Modifications
- Adding new webhook event types in `internal/webhook/`
- Modifying deployment command execution in `internal/deployment/`
- Adding new HTTP endpoints in `internal/server/`
- Extending configuration options in `internal/config/`

### When Making Changes
1. Read existing code in the relevant package to understand patterns
2. Follow existing error handling conventions
3. Add appropriate logging for visibility
4. Update configuration schema if needed
5. Test thoroughly with `go test ./...`
6. Verify configuration loading still works after changes

## Dependencies and External Integrations

### Go Modules
- Use `go mod tidy` to maintain dependencies
- Current Go version: 1.24.4
- External dependency: `github.com/BurntSushi/toml v1.5.0`

### GitHub Integration
- Webhook endpoint: `/admin/webhook` (POST, configurable via `webhook_path`)
- Event types: push events trigger deployments
- Signature verification using `X-Hub-Signature-256` header
- Repository name extraction from webhook payload

### File System
- Configuration search paths with precedence rules
- Log file creation and rotation
- Working directory changes during command execution
- Proper file permissions and cleanup