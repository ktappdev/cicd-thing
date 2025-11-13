# WARP.md

This file provides guidance to WARP (warp.dev) when working with code in this repository.

## Project Overview

**CICD-Thing** is a Go-based CI/CD deployment orchestrator that automatically deploys websites and applications when code is pushed to GitHub. It receives GitHub webhooks, executes deployment commands, and provides monitoring capabilities through a web interface.

## Core Architecture

### Main Components

1. **Configuration System** (`internal/config/`)
   - TOML-based configuration with hierarchical search paths
   - Auto-generates default config if none found
   - Validates required fields on startup

2. **Deployment Executor** (`internal/deployment/`)
   - Worker pool pattern with configurable concurrency
   - Command execution with timeout and context cancellation
   - Deployment locking mechanism to prevent concurrent deploys

3. **HTTP Server** (`internal/server/`)
   - REST API endpoints for webhooks, manual deploys, health, status, logs
   - Security middleware for IP allowlisting and rate limiting
   - Authentication via API keys

4. **Webhook Handler** (`internal/webhook/`)
   - GitHub webhook signature verification
   - Parses push events and triggers deployments

5. **Logger System** (`internal/logger/`)
   - Structured logging to file and stdout
   - Project-specific log prefixing

## Development Commands

### Build and Run
```bash
# Install dependencies
go mod tidy

# Build the application
go build -o cicd-thing .

# Run with auto-generated config
./cicd-thing

# Build and run in one command
go run main.go
```

### Development and Testing
```bash
# Test compilation without building
go build -o /dev/null .

# Check for Go syntax issues
go vet ./...

# Format code
go fmt ./...

# Run with race detection (for development)
go run -race main.go
```

## Configuration System

The application searches for `config.toml` in these locations (in order):
1. `./config.toml` (current directory) - Best for development
2. `./config/config.toml` (local config directory)
3. `/etc/cicd-thing/config.toml` (system-wide) - Best for production
4. `/usr/local/etc/cicd-thing/config.toml`
5. `~/.config/cicd-thing/config.toml` (user home)

### Required Configuration
- `webhook_secret`: GitHub webhook secret for signature verification
- `api_key`: API key for manual deployment endpoints
- `repositories`: Map of GitHub repos to local paths

### Key Configuration Sections
- `[repositories]`: Maps GitHub repo names to local deployment paths
- `[commands]`: Per-repository deployment commands

## API Endpoints

The server runs on port 3000 (configurable) and provides:
- `POST /webhook` - GitHub webhook receiver
- `POST /deploy` - Manual deployment trigger (requires auth)
- `GET /health` - Health check and configuration info
- `GET /status` - Deployment status and repository configuration
- `GET /logs` - Web-based log viewer with rate limiting

## Internal Package Structure

### Key Patterns
- **Worker Pool**: `deployment.Executor` uses goroutine workers for concurrent deployments
- **Channel Communication**: Results and requests flow through buffered channels
- **Context Cancellation**: All command execution supports timeout and cancellation
- **Middleware Chain**: HTTP handlers wrapped with security, auth, and rate limiting

### Core Types
- `deployment.Request`: Represents a deployment job
- `deployment.Result`: Contains execution results and timing
- `config.Config`: Centralized configuration with validation
- `deployment.Lock`: Prevents concurrent deployments of same app

## Security Features

1. **GitHub Signature Verification**: Validates webhook authenticity using HMAC-SHA256
2. **API Key Authentication**: Bearer token auth for manual deployment endpoints
3. **Rate Limiting**: Prevents abuse of log viewer endpoint (30 req/min)
4. **External Network Controls**: Use reverse proxy / firewall / WAF for IP restrictions

## Development Workflow

### Adding New Endpoints
1. Add handler function to `internal/server/server.go`
2. Register route in `Start()` method with appropriate middleware
3. Update API documentation in `API.md`

### Modifying Deployment Logic
1. Core execution logic in `internal/deployment/executor.go`
2. Types and constants in `internal/deployment/types.go`
3. Worker pool processes requests from `queue` channel
4. Results sent to `results` channel for logging

### Configuration Changes
1. Update `Config` struct in `internal/config/config.go`
2. Add validation in `validate()` method if required
3. Update default config template in `createDefaultConfig()`
4. Update example file: `config.toml.example`

## Production Deployment

### System Installation
```bash
# Build for production
go build -o cicd-thing .

# Install to system location
sudo cp cicd-thing /usr/local/bin/
sudo mkdir -p /etc/cicd-thing
sudo cp config.toml.example /etc/cicd-thing/config.toml

# Configure production settings
sudo nano /etc/cicd-thing/config.toml
```

### Systemd Service (Optional)
Create `/etc/systemd/system/cicd-thing.service` for automatic startup and process management.

## Monitoring and Debugging

### Log Analysis
- Deployment logs: `timestamp | repository | branch | commit | status | duration`
- Web log viewer at `/logs` with project identification
- Configurable log file location via `log_file` setting

### Dry Run Mode
Set `dry_run = true` in configuration to simulate deployments without execution.

### Health Monitoring
- `/health` endpoint provides service status and feature flags
- `/status` endpoint shows repository configuration and deployment stats

## Common Patterns

### Command Execution
All deployment commands executed via `os/exec` with:
- Context timeout from configuration
- Working directory set to local repository path
- Combined stdout/stderr capture
- Exit code tracking for error handling

### Error Recovery
- Deployment locks prevent race conditions
- Queue buffering handles burst webhook activity
- Graceful shutdown on SIGTERM/SIGINT

## Dependencies

The project uses minimal external dependencies:
- `github.com/BurntSushi/toml` for configuration parsing
- Standard library for HTTP server, command execution, and JSON handling

This keeps the binary lightweight and reduces security surface area.