# CI/CD Thing API Documentation

## Overview

The CI/CD Thing provides a REST API for managing deployments and monitoring the system.

## Authentication

Most endpoints require authentication using a Bearer token:

```
Authorization: Bearer your_api_key_here
```

The API key is configured via the `API_KEY` environment variable.

## Endpoints

### Health Check

**GET /health**

Returns the health status and configuration of the service.

**Response:**
```json
{
  "status": "healthy",
  "service": "cicd-thing",
  "version": "1.0.0",
  "uptime": "running",
  "config": {
    "port": "3000",
    "concurrency_limit": 2,
    "timeout_seconds": 300,
    "branch_filter": "main",
    "dry_run": false,
    "repositories": 2
  },
  "features": {
    "webhook_listener": true,
    "manual_deployment": true,
    "rollback_support": true,
    "ip_allowlist": false,
    "notifications": false
  }
}
```

### System Status

**GET /status**

Returns detailed system status including deployment information.

**Response:**
```json
{
  "service": "cicd-thing",
  "status": "running",
  "deployments": {
    "active": 0,
    "queued": 0,
    "completed": 0
  },
  "repositories": {
    "octocat/Hello-World": "~/projects/hello-world",
    "myorg/api": "~/apps/api"
  },
  "configuration": {
    "concurrency_limit": 2,
    "timeout_seconds": 300,
    "branch_filter": "main",
    "dry_run": false
  }
}
```

### Manual Deployment

**POST /deploy**

Triggers a manual deployment for a specified repository.

**Authentication:** Required

**Query Parameters:**
- `repo` (required): Repository full name (e.g., "octocat/Hello-World")
- `branch` (optional): Branch to deploy (defaults to configured branch filter)
- `commit` (optional): Commit hash to deploy (defaults to "HEAD")

**Example Request:**
```bash
curl -X POST "http://localhost:3000/deploy?repo=octocat/Hello-World&branch=main&commit=abc123" \
  -H "Authorization: Bearer your_api_key"
```

**Success Response (200):**
```json
{
  "status": "success",
  "message": "Manual deployment triggered",
  "repository": "octocat/Hello-World",
  "branch": "main",
  "commit": "abc123"
}
```

**Error Responses:**

**400 Bad Request:**
```json
{
  "error": "Missing required parameter: repo"
}
```

**401 Unauthorized:**
```json
{
  "error": "Unauthorized"
}
```

**500 Internal Server Error:**
```json
{
  "error": "Failed to trigger deployment: deployment already in progress for app Hello-World"
}
```

### Log Viewer

**GET /logs**

Provides a web-based log viewer interface for monitoring deployment logs and system activity.

**Authentication:** Required

**Query Parameters:**
- `limit` (optional): Number of log lines to display (10, 20, 50, 100, 200). Defaults to 50.

**Example Request:**
```bash
curl "http://localhost:3000/logs?limit=100" \
  -H "Authorization: Bearer your_api_key"
```

**Response:**
Returns an HTML page with:
- Interactive log viewer with dark theme
- Dropdown to select number of lines (10, 20, 50, 100, 200)
- Refresh button for manual updates
- Auto-refresh every 30 seconds
- Syntax highlighting for different log levels
- Responsive design for mobile and desktop

**Features:**
- Real-time log monitoring
- Configurable line limits
- Color-coded log levels (ERROR, INFO, SUCCESS, WARNING)
- Timestamp highlighting
- Scrollable log container
- Auto-refresh functionality

**Security:**
- Requires API key authentication
- Respects IP allowlist configuration
- Read-only access to log files

### GitHub Webhook

**POST /webhook**

Receives GitHub webhook events and triggers deployments automatically.

**Headers:**
- `X-Hub-Signature-256`: GitHub webhook signature
- `X-GitHub-Event`: Event type (must be "push")
- `Content-Type`: application/json

**Request Body:** GitHub webhook payload (JSON)

**Success Response (200):**
```
Deployment triggered successfully
```

**Other Responses:**
- `200`: "No deployment triggered" (wrong branch or no mapping)
- `200`: "Event type not supported" (non-push events)
- `400`: "Failed to parse webhook payload"
- `401`: "Invalid signature"
- `405`: "Method not allowed"

## Error Handling

All API endpoints return appropriate HTTP status codes:

- `200`: Success
- `400`: Bad Request (invalid parameters)
- `401`: Unauthorized (missing or invalid API key)
- `403`: Forbidden (IP not allowed)
- `405`: Method Not Allowed
- `500`: Internal Server Error

Error responses include a descriptive message:

```json
{
  "error": "Description of the error"
}
```

## Rate Limiting

Currently, no rate limiting is implemented, but it can be added via middleware.

## IP Allowlisting

If configured, the `/webhook` and `/deploy` endpoints will only accept requests from allowed IP addresses or CIDR blocks.

## Examples

### Check Service Health

```bash
curl http://localhost:3000/health
```

### Trigger Manual Deployment

```bash
curl -X POST "http://localhost:3000/deploy?repo=myorg/myapp" \
  -H "Authorization: Bearer your_api_key"
```

### Get System Status

```bash
curl http://localhost:3000/status
```

### Configure GitHub Webhook

1. Go to your repository settings
2. Navigate to "Webhooks"
3. Click "Add webhook"
4. Set Payload URL: `http://your-server:3000/webhook`
5. Set Content type: `application/json`
6. Set Secret: Your webhook secret
7. Select "Just the push event"
8. Ensure "Active" is checked
9. Click "Add webhook"

## Security Considerations

1. **Always use HTTPS** in production
2. **Keep API keys secure** and rotate them regularly
3. **Configure IP allowlisting** to restrict access
4. **Monitor logs** for suspicious activity
5. **Use strong webhook secrets** (at least 20 characters)

## Monitoring Integration

The health and status endpoints can be integrated with monitoring systems:

- **Prometheus**: Scrape `/health` endpoint
- **Nagios**: Monitor `/health` for status changes
- **Custom monitoring**: Parse JSON responses for alerts
