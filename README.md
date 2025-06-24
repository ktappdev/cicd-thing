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

1. **Clone and Setup**
   ```bash
   git clone <your-repo>
   cd cicd-thing
   go mod tidy
   ```

2. **Configure Environment**
   ```bash
   cp .env.example .env
   # Edit .env with your settings
   ```

3. **Start the Server**
   ```bash
   go run main.go
   # Or build and run
   go build -o cicd-thing .
   ./cicd-thing
   ```

4. **Configure GitHub Webhooks**
   - Go to your repository settings
   - Add webhook pointing to `http://your-server:3000/webhook`
   - Set content type to `application/json`
   - Add your webhook secret
   - Select "Push events"

## Configuration

### Environment Variables

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `PORT` | Server port | `3000` | No |
| `WEBHOOK_SECRET` | GitHub webhook secret | - | Yes |
| `API_KEY` | API authentication key | - | Yes |
| `LOG_FILE` | Log file path | `/var/log/deployer.log` | No |
| `REPO_MAP` | Repository mappings | - | Yes |
| `DEFAULT_COMMANDS` | Default deployment commands | `git pull && npm ci && npm run build` | No |
| `BRANCH_FILTER` | Branch to deploy | `main` | No |
| `CONCURRENCY_LIMIT` | Max concurrent deployments | `2` | No |
| `TIMEOUT_SECONDS` | Deployment timeout | `300` | No |
| `DRY_RUN` | Test mode | `false` | No |

### Repository Mapping

Format: `repo1:path1,repo2:path2`

Example:
```
REPO_MAP=octocat/Hello-World:~/projects/hello-world,myorg/api:~/apps/api
```

### Per-App Commands

Format: `COMMANDS_<app-name>=<command>`

Example:
```
COMMANDS_Hello-World=git pull && npm ci && npm run build && pm2 restart hello-world
COMMANDS_api=git pull && go build && systemctl restart api
```

### Rollback Commands

Format: `ROLLBACK_COMMANDS_<app-name>=<command>`

Example:
```
ROLLBACK_COMMANDS_Hello-World=git reset --hard HEAD~1 && pm2 restart hello-world
```

## API Endpoints

### `POST /webhook`
GitHub webhook receiver. Automatically triggered by GitHub.

**Headers:**
- `X-Hub-Signature-256`: GitHub signature
- `X-GitHub-Event`: Event type (must be "push")

### `POST /deploy`
Manual deployment trigger.

**Authentication:** Bearer token required
```bash
curl -X POST "http://localhost:3000/deploy?repo=octocat/Hello-World&branch=main" \
  -H "Authorization: Bearer your_api_key"
```

**Query Parameters:**
- `repo` (required): Repository full name
- `branch` (optional): Branch to deploy (defaults to configured branch filter)
- `commit` (optional): Commit hash (defaults to "HEAD")

### `GET /health`
Health check endpoint with system information.

**Response:**
```json
{
  "status": "healthy",
  "service": "cicd-thing",
  "version": "1.0.0",
  "config": {
    "port": "3000",
    "concurrency_limit": 2,
    "repositories": 2
  },
  "features": {
    "webhook_listener": true,
    "manual_deployment": true,
    "rollback_support": true
  }
}
```

### `GET /status`
Deployment status and configuration information.

## Usage Examples

### Basic Deployment Flow

1. **Push to GitHub** → Webhook triggered → Deployment executed
2. **Manual deployment** via API
3. **Automatic rollback** on failure (if configured)

### Example Deployment Commands

**Node.js Application:**
```bash
COMMANDS_myapp=git pull && npm ci && npm run build && pm2 restart myapp
```

**Go Application:**
```bash
COMMANDS_api=git pull && go build -o api . && systemctl restart api
```

**Docker Application:**
```bash
COMMANDS_webapp=git pull && docker build -t webapp . && docker-compose up -d
```

### Security Setup

1. **Generate webhook secret:**
   ```bash
   openssl rand -hex 20
   ```

2. **Generate API key:**
   ```bash
   openssl rand -hex 32
   ```

3. **Configure IP allowlist (optional):**
   ```
   IP_ALLOWLIST=192.168.1.0/24,10.0.0.0/8
   ```

## Monitoring

### Logs

All deployment events are logged to the configured log file and stdout:

```
2025-06-24T10:15:00Z | Hello-World | main | 1481a2de | STARTED
2025-06-24T10:15:10Z | Hello-World | main | 1481a2de | SUCCESS | 10s
2025-06-24T10:16:00Z | api | main | 2592b3ef | FAILED | 5s | error: build failed
```

### Health Monitoring

Monitor the `/health` endpoint for service status and configuration.

## Troubleshooting

### Common Issues

1. **Webhook not received:**
   - Check GitHub webhook configuration
   - Verify webhook secret matches
   - Check server logs for signature verification errors

2. **Deployment fails:**
   - Check repository mapping in REPO_MAP
   - Verify local path exists and is accessible
   - Check deployment commands are correct
   - Review timeout settings

3. **Permission errors:**
   - Ensure server has access to local repositories
   - Check file permissions on deployment paths
   - Verify user has necessary privileges for commands

### Debug Mode

Enable dry run mode for testing:
```
DRY_RUN=true
```

This will simulate deployments without executing commands.

## License

MIT License
