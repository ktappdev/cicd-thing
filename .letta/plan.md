# Plan: Make Webhook and Secured Endpoints Configurable

## Overview
Currently, all endpoints are hardcoded (`/webhook`, `/deploy`, `/logs`). The goal is to:
1. Make these paths configurable via `config.toml`
2. Move all secured endpoints under `/admin/` by default
3. Keep public endpoints (`/health`, `/status`) as-is

## Changes Required

### 1. Config Structure Changes
**File:** `internal/config/config.go`

Add three new fields to the `Config` struct (under "Server settings" section):
```go
// Endpoint paths
WebhookPath string `toml:"webhook_path"`
DeployPath  string `toml:"deploy_path"`
LogsPath    string `toml:"logs_path"`
```

Add defaults in the `Load()` function:
```go
WebhookPath: "/admin/webhook",
DeployPath:  "/admin/deploy",
LogsPath:    "/admin/logs",
```

Update `createDefaultConfig()` template to include:
```toml
# Endpoint paths (configurable for security through obscurity or organization)
webhook_path = "/admin/webhook"  # GitHub webhook receiver
deploy_path = "/admin/deploy"    # Manual deployment endpoint
logs_path = "/admin/logs"        # Log viewer endpoint
```

### 2. Server Route Registration
**File:** `internal/server/server.go`

Change `Start()` method from:
```go
http.HandleFunc("/webhook", s.webhookHandler.HandleWebhook)
http.HandleFunc("/deploy", s.security.AuthMiddleware(s.handleManualDeploy))
http.HandleFunc("/logs", s.security.RateLimitMiddleware(s.handleLogs))
```

To:
```go
http.HandleFunc(s.config.WebhookPath, s.webhookHandler.HandleWebhook)
log.Printf("Webhook endpoint: %s", s.config.WebhookPath)

if s.config.APIKey != "" {
    http.HandleFunc(s.config.DeployPath, s.security.AuthMiddleware(s.handleManualDeploy))
    log.Printf("Manual deployment endpoint: %s", s.config.DeployPath)
} else {
    log.Printf("⚠️  Manual deployment endpoint disabled (no api_key configured)")
}

http.HandleFunc(s.config.LogsPath, s.security.RateLimitMiddleware(s.handleLogs))
log.Printf("Logs viewer endpoint: %s", s.config.LogsPath)
```

Keep `/health` and `/status` as-is (public endpoints).

### 3. Update config.toml.example
**File:** `config.toml.example`

Add after the `port`/`webhook_secret`/`api_key` section:
```toml
# Endpoint paths - Customize these for security or organization
webhook_path = "/admin/webhook"  # GitHub webhook receiver
deploy_path = "/admin/deploy"    # Manual deployment API (requires api_key)
logs_path = "/admin/logs"        # Log viewer (rate limited)
```

### 4. Documentation Updates

**Files to update with new default paths:**

- **README.md:**
  - Line ~143: Change to `http://your-server:3000/admin/webhook`
  - Section "Available Endpoints" (~line 283-314): Update all paths
  - Line ~295: `/webhook` → describe configurable path
  - Line ~372: Update security recommendations
  - Line ~445: Update webhook path reference

- **AGENTS.md:**
  - Line ~172: Update endpoint list with new defaults

- **API.md:**
  - Line ~173: Update POST /webhook documentation
  - Line ~255: Update GitHub setup instructions
  - All curl examples throughout

- **FAQ.md:**
  - Line ~90: Update webhook URL example

- **GETTING_STARTED.md:**
  - Line ~103: Update webhook URL in GitHub setup

- **DEPLOYMENT_EXAMPLES.md:**
  - Line ~188: Update security recommendations
  - Line ~257: Update curl example

- **CLAUDE.md:**
  - Line ~102: Update endpoint list

- **WARP.md:**
  - Line ~89: Update endpoint reference

## Breaking Changes & Migration

**What breaks:**
1. **GitHub webhook configuration** - Must update Payload URL from `/webhook` to `/admin/webhook` (or custom path)
2. **Manual deployment scripts** - Any curl commands using `/deploy` must change to `/admin/deploy`
3. **Log viewer bookmarks** - `/logs` → `/admin/logs`

**Migration path (since app is in development):**
1. Deploy the change
2. Update GitHub webhook URL in repository settings
3. Update any scripts/bookmarks
4. (Optional) Customize paths in config.toml for additional obscurity

## Configuration Example

After changes, users can customize like:
```toml
# Standard approach - organized under /admin
webhook_path = "/admin/webhook"
deploy_path = "/admin/deploy"
logs_path = "/admin/logs"

# Or security through obscurity:
webhook_path = "/secret-gh-hook-8dj2k"
deploy_path = "/deploy-endpoint-x9f2p"
logs_path = "/view-logs-k3m9z"

# Or simple/legacy:
webhook_path = "/webhook"
deploy_path = "/deploy"  
logs_path = "/logs"
```

## Files Modified

**Code:**
- `internal/config/config.go` - Add fields, defaults, template
- `internal/server/server.go` - Use config paths, add logging
- `config.toml.example` - Add new settings

**Documentation (8 files):**
- `README.md`
- `AGENTS.md`
- `API.md`
- `FAQ.md`
- `GETTING_STARTED.md`
- `DEPLOYMENT_EXAMPLES.md`
- `CLAUDE.md`
- `WARP.md`

## Testing Checklist
- [ ] Build succeeds: `go build -o cicd-thing .`
- [ ] Default config created with new paths
- [ ] Server logs show custom endpoint paths on startup
- [ ] GitHub webhook works with `/admin/webhook`
- [ ] Manual deploy works with `/admin/deploy` (if api_key set)
- [ ] Logs viewer accessible at `/admin/logs`
- [ ] `/health` and `/status` still work (unchanged)
