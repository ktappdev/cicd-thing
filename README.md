# CICD-Thing 🚀

🚀 **Automatically deploy your websites and applications when you push code to GitHub!**

This tool watches your GitHub repositories and automatically deploys your code to your server whenever you make changes. No more manual deployments - just push your code and it goes live!

## What Does This Do? 🤔

Imagine you have a website or app on GitHub. Every time you make changes and push them to GitHub, this tool will:

1. **🔔 Notice the changes** - Gets notified instantly when you push code
2. **📥 Download the latest code** - Pulls your changes to your server
3. **🔨 Build your project** - Runs commands like installing dependencies and building
4. **🚀 Deploy it live** - Restarts your website/app with the new code
5. **📊 Tell you what happened** - Logs everything and can send notifications

## Why Use This? ✨

- **⚡ Instant Deployments**: Your changes go live seconds after you push to GitHub
- **🔒 Secure**: Only deploys when GitHub sends the correct secret key
- **🛡️ Safe**: Can automatically undo deployments if something goes wrong
- **📱 Multiple Projects**: Handle many websites/apps from one tool
- **🎯 Smart**: Only deploys from specific branches (like `main`)
- **🔍 Transparent**: See exactly what's happening with detailed logs
- **📋 Web Log Viewer**: Real-time log monitoring with project identification and dark theme
- **⚙️ Flexible Configuration**: TOML-based config with multiple location support

## How It Works 🔄

```
1. You push code to GitHub
         ↓
2. GitHub sends a notification to this tool
         ↓
3. Tool downloads your latest code
         ↓
4. Tool runs your build commands (install, build, etc.)
         ↓
5. Tool restarts your website/app
         ↓
6. Your changes are now live! 🎉
```

**If something goes wrong:** The tool can automatically undo the deployment and restore the previous version.

## Quick Start Guide 🚀

### Step 1: Download and Setup
```bash
# Download the code
git clone <your-repo>
cd cicd-thing

# Install dependencies
go mod tidy
```

### Step 2: Configure Your Settings

The application uses TOML configuration files and will automatically create one for you!

```bash
# Build the application
go build -o cicd-thing .

# Run it once to create the default config
./cicd-thing
```

The application will create a `config.toml` file and show you what needs to be configured:

```
=== CONFIGURATION REQUIRED ===
A default configuration file has been created at: ./config.toml
Please edit this file with your settings before running the application again.
Required fields to configure:
  - webhook_secret: Your GitHub webhook secret
  - api_key: Your API key for authentication
  - repositories: Map of repository names to local paths
===============================
```

### Step 3: Edit Your Configuration
```bash
# Edit the config file with your settings
nano config.toml  # or use any text editor
```

### Step 4: Start the Tool
```bash
# Run the application
./cicd-thing
```

### Step 5: Connect to GitHub
1. Go to your GitHub repository
2. Click **Settings** → **Webhooks** → **Add webhook**
3. Set **Payload URL** to: `http://your-server:3000/webhook`
4. Set **Content type** to: `application/json`
5. Set **Secret** to the same value as `webhook_secret` in your config.toml file
6. Select **Just the push event**
7. Click **Add webhook**

🎉 **That's it!** Now when you push code to GitHub, it will automatically deploy!

## Configuration System ⚙️

### Configuration File Locations

The application searches for `config.toml` in these locations (in order):

1. `./config.toml` (current directory) - **Best for development**
2. `./config/config.toml` (local config directory)
3. `/etc/cicd-thing/config.toml` (system-wide config) - **Best for production**
4. `/usr/local/etc/cicd-thing/config.toml` (alternative system config)
5. `~/.config/cicd-thing/config.toml` (user home directory)

### Automatic Configuration Creation

If no configuration file is found, the application will:
- Create a comprehensive default `config.toml` in the current directory
- Include helpful comments and examples for all options
- Clearly mark required vs optional settings
- Exit with instructions for you to configure it

### Configuration Settings

Your `config.toml` file contains all the settings. Here's what each section means:

#### Required Settings (You MUST set these)

```toml
# Server settings
port = "3000"
webhook_secret = "YOUR_WEBHOOK_SECRET_HERE"  # REQUIRED
api_key = "YOUR_API_KEY_HERE"                # REQUIRED

# Repository mappings - REQUIRED
[repositories]
"my-app" = "/var/www/my-app"
"api-service" = "/opt/api-service"
```

| Setting | What It Does | Example |
|---------|--------------|----------|
| `webhook_secret` | Secret password GitHub uses to verify it's really GitHub calling | `abc123secret456` |
| `api_key` | Password for manually triggering deployments | `myapikey789` |
| `repositories` | Which GitHub repos go to which folders on your server | See examples below |

#### Optional Settings (Have good defaults)

```toml
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
ip_allowlist = ["192.168.1.0/24", "10.0.0.0/8"]
```

| Setting | What It Does | Default | Example |
|---------|--------------|---------|----------|
| `port` | What port the tool runs on | `3000` | `8080` |
| `branch_filter` | Only deploy from this branch | `main` | `production` |
| `timeout_seconds` | How long to wait before giving up | `300` (5 minutes) | `600` |
| `dry_run` | Test mode (doesn't actually deploy) | `false` | `true` |

### 📁 Repository Mapping (Which GitHub repos go where)

Tell the tool which GitHub repository goes to which folder on your server:

```toml
[repositories]
"johndoe/my-website" = "/var/www/my-website"
"johndoe/api-service" = "/opt/api-service"
"company/frontend" = "/var/www/frontend"
```

### 🔨 Deployment Commands (What to do when deploying)

Tell the tool what commands to run when deploying each project:

```toml
[commands]
# For a Node.js website:
"my-website" = "git pull && npm ci && npm run build && pm2 restart my-website"

# For a simple HTML site:
"my-site" = "git pull && rsync -av ./ /var/www/html/"

# For a Python app:
"my-app" = "git pull && pip install -r requirements.txt && systemctl restart my-app"

# For a Go application:
"api-service" = "git pull && go build -o api . && systemctl restart api"
```

### 🔄 Rollback Commands (What to do if deployment fails)

If something goes wrong, these commands will undo the deployment:

```toml
[rollback_commands]
"my-website" = "git checkout HEAD~1 && npm ci && npm run build && pm2 restart my-website"
"api-service" = "git checkout HEAD~1 && go build && systemctl restart api-service"
```

## 📚 Documentation for Everyone

- **📖 [Getting Started Guide](GETTING_STARTED.md)** - Step-by-step setup for beginners
- **❓ [FAQ](FAQ.md)** - Common questions and answers
- **🔧 [API Documentation](API.md)** - Technical API reference
- **💡 [Deployment Examples](DEPLOYMENT_EXAMPLES.md)** - Real-world configuration examples

## Available Endpoints 🌐

The tool provides several web endpoints you can use:

### 🔔 `/webhook` - GitHub Notifications
- **What it does:** Receives notifications from GitHub when you push code
- **Who uses it:** GitHub automatically calls this when you push code
- **You don't need to worry about this** - it's automatic!

### 🚀 `/deploy` - Manual Deployment
- **What it does:** Lets you trigger a deployment manually
- **How to use:**
  ```bash
  curl -X POST "http://your-server:3000/deploy?repo=username/repository" \
    -H "Authorization: Bearer your-api-key"
  ```
- **When to use:** When you want to deploy without pushing to GitHub

### ❤️ `/health` - Check if Tool is Working
- **What it does:** Shows if the tool is running properly
- **How to use:** Visit `http://your-server:3000/health` in your browser
- **What you'll see:** Information about the tool's status and configuration

### 📊 `/status` - Deployment Information
- **What it does:** Shows current deployment status and configuration
- **How to use:** Visit `http://your-server:3000/status` in your browser
- **What you'll see:** List of configured repositories and deployment settings

### 📋 `/logs` - Log Viewer
- **What it does:** Displays real-time deployment and system logs with project identification
- **How to use:** Visit `http://your-server:3000/logs?limit=50` in your browser
- **Authentication:** Requires API key (add `?api_key=your-key` to URL)
- **What you'll see:** Color-coded logs with project prefixes, configurable line limits, and auto-refresh

## Usage Examples

### Basic Deployment Flow

1. **Push to GitHub** → Webhook triggered → Deployment executed
2. **Manual deployment** via API
3. **Automatic rollback** on failure (if configured)

### Example Deployment Commands

**Node.js Application:**
```toml
[commands]
"myapp" = "git pull && npm ci && npm run build && pm2 restart myapp"
```

**Go Application:**
```toml
[commands]
"api" = "git pull && go build -o api . && systemctl restart api"
```

**Docker Application:**
```toml
[commands]
"webapp" = "git pull && docker build -t webapp . && docker-compose up -d"
```

**Static Website:**
```toml
[commands]
"website" = "git pull && npm run build && rsync -av dist/ /var/www/html/"
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
   ```toml
   ip_allowlist = ["192.168.1.0/24", "10.0.0.0/8"]
   ```

## Production Deployment 🏭

### System-wide Installation

1. **Build the application:**
   ```bash
   go build -o cicd-thing .
   ```

2. **Install to system location:**
   ```bash
   sudo cp cicd-thing /usr/local/bin/
   sudo mkdir -p /etc/cicd-thing
   sudo cp config.toml.example /etc/cicd-thing/config.toml
   ```

3. **Configure for production:**
   ```bash
   sudo nano /etc/cicd-thing/config.toml
   ```

4. **Create systemd service (optional):**
   ```bash
   sudo tee /etc/systemd/system/cicd-thing.service > /dev/null <<EOF
   [Unit]
   Description=CICD-Thing Deployment Orchestrator
   After=network.target

   [Service]
   Type=simple
   User=deploy
   ExecStart=/usr/local/bin/cicd-thing
   Restart=always
   RestartSec=5

   [Install]
   WantedBy=multi-user.target
   EOF

   sudo systemctl enable cicd-thing
   sudo systemctl start cicd-thing
   ```

## Monitoring

### Logs

All deployment events are logged to the configured log file and stdout:

```
2025-06-24T10:15:00Z | Hello-World | main | 1481a2de | STARTED
2025-06-24T10:15:10Z | Hello-World | main | 1481a2de | SUCCESS | 10s
2025-06-24T10:16:00Z | api | main | 2592b3ef | FAILED | 5s | error: build failed
```

**Log Format:**
- **Deployment logs:** `timestamp | repository | branch | commit | status | duration | error`
- **System logs:** `timestamp | level | message`
- **Web viewer adds project prefixes:** `[project-name]` or `[SYSTEM]` for easy identification

### Web Log Viewer

Access real-time logs through the web interface at `/logs`:

```bash
# View logs in your browser
http://localhost:3000/logs?limit=50
```

**Features:**
- 🏷️ **Project identification** - Each log line shows which project it belongs to:
  - `[my-app]` for deployment logs from specific repositories
  - `[SYSTEM]` for general server messages
  - `[UNKNOWN]` for unrecognized log formats
- 🎨 **Dark theme** optimized for log viewing
- 📊 **Configurable limits** (10, 20, 50, 100, 200 lines)
- 🔄 **Auto-refresh** every 30 seconds
- 🎯 **Manual refresh** button
- 🌈 **Color-coded** log levels (ERROR, INFO, SUCCESS, WARNING)
- 📱 **Mobile-friendly** responsive design
- 🔒 **Secure** - requires API key authentication

**Example log output:**
```
[SYSTEM] 2025-06-24T11:26:30-04:00 | INFO | Server initialized
[my-app] 2025-06-24T11:26:30-04:00 | my-app | main | abc123 | SUCCESS | 2.5s
[api-service] 2025-06-24T11:26:31-04:00 | api-service | develop | def456 | FAILED | error: build failed
```

### Health Monitoring

Monitor the `/health` endpoint for service status and configuration.

## Troubleshooting

### Common Issues

1. **Configuration file not found:**
   - Run the application once to create default config
   - Check the search locations listed above
   - Ensure file permissions are correct

2. **Webhook not received:**
   - Check GitHub webhook configuration
   - Verify webhook secret matches config.toml
   - Check server logs for signature verification errors

3. **Deployment fails:**
   - Check repository mapping in `[repositories]` section
   - Verify local path exists and is accessible
   - Check deployment commands are correct
   - Review timeout settings

4. **Permission errors:**
   - Ensure server has access to local repositories
   - Check file permissions on deployment paths
   - Verify user has necessary privileges for commands

### Debug Mode

Enable dry run mode for testing:
```toml
dry_run = true
```

This will simulate deployments without executing commands.

### Configuration Validation

The application validates your configuration on startup and will show clear error messages for:
- Missing required fields
- Invalid TOML syntax
- Incorrect file paths
- Network configuration issues

## License

MIT License
