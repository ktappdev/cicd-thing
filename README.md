# CI/CD Thing - Automatic Website & App Deployment

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
```bash
# Copy the example configuration
cp .env.example .env

# Edit the .env file with your settings (see Configuration section below)
nano .env  # or use any text editor
```

### Step 3: Start the Tool
```bash
# Option 1: Run directly
go run main.go

# Option 2: Build and run (recommended for production)
go build -o cicd-thing .
./cicd-thing
```

### Step 4: Connect to GitHub
1. Go to your GitHub repository
2. Click **Settings** → **Webhooks** → **Add webhook**
3. Set **Payload URL** to: `http://your-server:3000/webhook`
4. Set **Content type** to: `application/json`
5. Set **Secret** to the same value as `WEBHOOK_SECRET` in your .env file
6. Select **Just the push event**
7. Click **Add webhook**

🎉 **That's it!** Now when you push code to GitHub, it will automatically deploy!

## 📚 Documentation for Everyone

- **📖 [Getting Started Guide](GETTING_STARTED.md)** - Step-by-step setup for beginners
- **❓ [FAQ](FAQ.md)** - Common questions and answers
- **🔧 [API Documentation](API.md)** - Technical API reference
- **💡 [Deployment Examples](DEPLOYMENT_EXAMPLES.md)** - Real-world configuration examples

## Configuration Settings ⚙️

You need to edit the `.env` file to tell the tool about your projects. Here's what each setting means:

### Required Settings (You MUST set these)

| Setting | What It Does | Example |
|---------|--------------|---------|
| `WEBHOOK_SECRET` | Secret password GitHub uses to verify it's really GitHub calling | `abc123secret456` |
| `API_KEY` | Password for manually triggering deployments | `myapikey789` |
| `REPO_MAP` | Which GitHub repos go to which folders on your server | `myuser/website:~/mysite` |

### Optional Settings (Have good defaults)

| Setting | What It Does | Default | Example |
|---------|--------------|---------|---------|
| `PORT` | What port the tool runs on | `3000` | `8080` |
| `BRANCH_FILTER` | Only deploy from this branch | `main` | `production` |
| `TIMEOUT_SECONDS` | How long to wait before giving up | `300` (5 minutes) | `600` |
| `DRY_RUN` | Test mode (doesn't actually deploy) | `false` | `true` |

### 📁 Repository Mapping (Which GitHub repos go where)

Tell the tool which GitHub repository goes to which folder on your server:

```bash
# Format: github-user/repo-name:path-on-your-server
REPO_MAP=johndoe/my-website:~/websites/my-website

# Multiple projects:
REPO_MAP=johndoe/website:~/sites/website,johndoe/api:~/apps/api
```

### 🔨 Deployment Commands (What to do when deploying)

Tell the tool what commands to run when deploying each project:

```bash
# For a Node.js website:
COMMANDS_my-website=git pull && npm install && npm run build && pm2 restart my-website

# For a simple HTML site:
COMMANDS_my-site=git pull && rsync -av ./ /var/www/html/

# For a Python app:
COMMANDS_my-app=git pull && pip install -r requirements.txt && systemctl restart my-app
```

### 🔄 Rollback Commands (What to do if deployment fails)

If something goes wrong, these commands will undo the deployment:

```bash
# Go back to previous version and restart:
ROLLBACK_COMMANDS_my-website=git reset --hard HEAD~1 && npm run build && pm2 restart my-website
```

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
