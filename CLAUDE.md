# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

CICD-Thing is a Go-based deployment orchestrator that automatically deploys applications when code is pushed to GitHub. It receives webhook notifications from GitHub, downloads the latest code, runs build commands, and restarts applications.

## Installation

### Prerequisites
- Go 1.24.4 or later
- Git (for deployment commands)

### Install with go install (Recommended)
```bash
# Install the latest version globally
go install github.com/ktappdev/cicd-thing@latest

# The binary will be available in your GOPATH/bin (usually ~/go/bin/)
# Make sure ~/go/bin is in your PATH
export PATH=$PATH:$(go env GOPATH)/bin

# Run the application
cicd-thing
```

### Install from Source
```bash
# Clone the repository
git clone https://github.com/ktappdev/cicd-thing.git
cd cicd-thing

# Install dependencies
go mod tidy

# Build and install locally
go install .
```

## Development Commands

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

## Architecture

### Core Components

1. **Configuration** (`internal/config/`) - TOML-based configuration system
2. **Webhook Handler** (`internal/webhook/`) - Processes GitHub webhook notifications
3. **Deployment Executor** (`internal/deployment/`) - Executes deployment commands
4. **Server** (`internal/server/`) - HTTP server with REST endpoints
5. **Logger** (`internal/logger/`) - Structured logging with project identification
6. **Security** (`internal/security/`) - Authentication and authorization middleware
7. **Mapping** (`internal/mapping/`) - Repository to path mapping logic

### Key Dependencies

- `github.com/BurntSushi/toml` - Configuration file parsing
- Standard library only for core functionality

### Configuration System

The application searches for `config.toml` in multiple locations:
1. `./config.toml` (current directory)
2. `./config/config.toml`
3. `/etc/cicd-thing/config.toml`
4. `/usr/local/etc/cicd-thing/config.toml`
5. `~/.config/cicd-thing/config.toml`

Creates default configuration automatically if none exists.

### HTTP Endpoints

- `POST /admin/webhook` - GitHub webhook receiver (configurable)
- `POST /deploy` - Manual deployment trigger
- `GET /health` - Health check endpoint
- `GET /status` - Deployment status information
- `GET /logs` - Real-time log viewer (rate limited: 30 req/min)

### Deployment Flow

1. Configuration loads from TOML file
2. Server starts and listens for webhooks
3. GitHub webhook triggers deployment
4. Deployment executor runs configured commands
5. Results are logged and notifications sent
6. Web log viewer provides real-time monitoring

### Logging System

- Structured logging with project identification
- Project prefixes: `[project-name]`, `[SYSTEM]`, `[UNKNOWN]`
- Web viewer with dark theme and auto-refresh
- Rate limiting for log endpoint performance

### Security Features

- Webhook signature verification
- API key authentication for manual deployments
- Optional IP allowlist
- Command execution timeout controls