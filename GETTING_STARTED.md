# Getting Started with CI/CD Thing 🚀

## What Is This Tool? 🤔

**CI/CD Thing** is like having a robot assistant that automatically updates your website or app whenever you make changes to your code on GitHub.

### In Simple Terms:
- You make changes to your website code
- You save those changes to GitHub
- This tool notices and automatically updates your live website
- Your visitors see the new changes immediately!

## Before You Start 📋

You'll need:
- ✅ A website or app stored on GitHub
- ✅ A server where your website runs (like a VPS, cloud server, etc.)
- ✅ Basic knowledge of using the command line
- ✅ Your website's code already on your server

## Step-by-Step Setup 👣

### Step 1: Download the Tool
```bash
# Go to your server and download the tool
git clone https://github.com/your-username/cicd-thing
cd cicd-thing
```

### Step 2: Install Dependencies
```bash
# This installs everything the tool needs to work
go mod tidy
```

### Step 3: Create Your Configuration

1. **Copy the example settings:**
   ```bash
   cp .env.example .env
   ```

2. **Edit your settings:**
   ```bash
   nano .env
   ```

3. **Fill in these important settings:**
   ```bash
   # Create a secret password (use any random text)
   WEBHOOK_SECRET=your-secret-password-here
   
   # Create an API key (use any random text)
   API_KEY=your-api-key-here
   
   # Tell it where your website code is
   # Format: github-username/repository-name:path-on-your-server
   REPO_MAP=johndoe/my-website:~/websites/my-website
   
   # What commands to run when deploying (examples below)
   COMMANDS_my-website=git pull && npm install && npm run build && pm2 restart my-website
   ```

### Step 4: Set Up Your Deployment Commands

This is what the tool will do when it deploys your code. Here are examples for different types of projects:

**For a Node.js website:**
```bash
COMMANDS_my-website=git pull && npm install && npm run build && pm2 restart my-website
```

**For a simple HTML website:**
```bash
COMMANDS_my-website=git pull && rsync -av ./ /var/www/html/
```

**For a Python Flask app:**
```bash
COMMANDS_my-app=git pull && pip install -r requirements.txt && systemctl restart my-app
```

### Step 5: Start the Tool
```bash
# Start the tool
./cicd-thing

# You should see something like:
# Starting CI/CD Thing deployment orchestrator
# Server initialized
# Starting server on port 3000
```

### Step 6: Connect GitHub

1. **Go to your GitHub repository**
2. **Click "Settings" (in your repo, not your account)**
3. **Click "Webhooks" in the left sidebar**
4. **Click "Add webhook"**
5. **Fill in the form:**
   - **Payload URL:** `http://your-server-ip:3000/webhook`
   - **Content type:** `application/json`
   - **Secret:** The same password you put in `WEBHOOK_SECRET`
   - **Which events:** Select "Just the push event"
6. **Click "Add webhook"**

## Testing It Works 🧪

### Test 1: Check the Tool is Running
Open a web browser and go to: `http://your-server-ip:3000/health`

You should see something like:
```json
{
  "status": "healthy",
  "service": "cicd-thing"
}
```

### Test 2: Make a Small Change
1. Edit a file in your GitHub repository (like README.md)
2. Commit and push the change
3. Watch your server logs - you should see deployment activity
4. Check if your website was updated

## Common Issues & Solutions 🔧

### "Webhook not received"
- ✅ Check your server's firewall allows port 3000
- ✅ Make sure your server IP is correct in the GitHub webhook
- ✅ Verify the webhook secret matches exactly

### "Deployment failed"
- ✅ Check the deployment commands work when you run them manually
- ✅ Make sure the tool has permission to access your website files
- ✅ Check the logs for specific error messages

### "Permission denied"
- ✅ Make sure the user running the tool has permission to update your website
- ✅ Check file ownership and permissions on your website directory

## Understanding the Logs 📊

The tool shows you what's happening:

```
2025-06-24T10:15:00Z | my-website | main | abc123 | STARTED
2025-06-24T10:15:10Z | my-website | main | abc123 | SUCCESS | 10s
```

This means:
- **Time:** When it happened
- **Project:** Which website/app
- **Branch:** Which branch was deployed (usually "main")
- **Commit:** The code version (first 6 characters)
- **Status:** What happened (STARTED, SUCCESS, FAILED)
- **Duration:** How long it took

## Advanced Features 🎯

### Dry Run Mode (Testing)
Add this to your `.env` to test without actually deploying:
```bash
DRY_RUN=true
```

### Deploy Only from Specific Branch
Only deploy when you push to the "main" branch:
```bash
BRANCH_FILTER=main
```

### Manual Deployment
You can trigger a deployment manually:
```bash
curl -X POST "http://your-server:3000/deploy?repo=username/repository" \
  -H "Authorization: Bearer your-api-key"
```

### Automatic Rollback
If a deployment fails, automatically go back to the previous version:
```bash
ROLLBACK_COMMANDS_my-website=git reset --hard HEAD~1 && npm run build && pm2 restart my-website
```

## Getting Help 🆘

### Check the Logs
```bash
# See what the tool is doing
tail -f /var/log/deployer.log

# Or if you're running it directly:
# The logs will show in your terminal
```

### Check Tool Status
Visit: `http://your-server:3000/status`

### Common Commands
```bash
# Stop the tool
Ctrl+C (if running in terminal)

# Start the tool in background
nohup ./cicd-thing &

# Check if it's running
ps aux | grep cicd-thing
```

## Security Tips 🔒

1. **Use strong secrets:** Make your `WEBHOOK_SECRET` and `API_KEY` long and random
2. **Firewall:** Only allow GitHub's IP addresses to access your webhook endpoint
3. **HTTPS:** Use HTTPS in production (put a reverse proxy like nginx in front)
4. **Permissions:** Run the tool with minimal permissions needed

## Next Steps 🎉

Once you have it working:
1. Set up automatic startup (systemd service)
2. Configure log rotation
3. Set up monitoring and alerts
4. Add more repositories
5. Explore notification options (Slack, email)

**Congratulations!** You now have automatic deployments set up! 🚀
