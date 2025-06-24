# CI/CD Thing - Go-Based GitHub Webhook Deployment Orchestrator

A robust, production-ready deployment orchestrator in Go that listens for GitHub webhook events and executes configurable deployment commands.

## Features

- **Webhook Listener**: Receives and processes GitHub webhook events
- **Repository Mapping**: Maps GitHub repositories to local folders
- **Configurable Commands**: Per-app deployment commands with fallbacks
- **Security**: Webhook secret verification and IP allowlisting
- **Concurrency Control**: Per-app queuing and deployment locking
- **Timeout Handling**: Configurable timeouts for deployments
- **Rollback Support**: Automatic rollback on deployment failure
- **Manual API**: Authenticated endpoint for manual deployments
- **Comprehensive Logging**: Detailed deployment event tracking
- **Health Monitoring**: Health check endpoints
- **Branch Filtering**: Deploy only from specified branches
- **Dry Run Mode**: Test configurations without execution

## Project Structure

```
cicd-thing/
├── main.go                    # Application entry point
├── cmd/
│   └── deployer/             # Alternative entry points
├── internal/
│   ├── config/               # Configuration management
│   ├── server/               # HTTP server and routing
│   ├── webhook/              # GitHub webhook handling
│   ├── deployment/           # Deployment execution engine
│   ├── logger/               # Logging system
│   └── security/             # Security features
├── .env.example              # Example configuration
└── README.md                 # This file
```

## Quick Start

1. Copy `.env.example` to `.env` and configure your settings
2. Run `go mod tidy` to install dependencies
3. Run `go run main.go` to start the server
4. Configure your GitHub repository webhooks to point to your server

## Configuration

See `.env.example` for all available configuration options.

## API Endpoints

- `POST /webhook` - GitHub webhook receiver
- `POST /deploy` - Manual deployment trigger (authenticated)
- `GET /health` - Health check endpoint

## License

MIT License
