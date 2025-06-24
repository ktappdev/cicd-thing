# Frequently Asked Questions (FAQ) 🤔

## General Questions

### What exactly does this tool do?
This tool automatically deploys (updates) your website or application whenever you push new code to GitHub. Think of it as a robot that watches your GitHub repository and updates your live website whenever you make changes.

### Do I need to be a programmer to use this?
You need basic command-line knowledge and understanding of how your website is deployed. If you can copy files and run commands on your server, you can use this tool.

### What types of projects can I deploy?
- Static websites (HTML, CSS, JavaScript)
- Node.js applications
- Python web apps (Django, Flask)
- Go applications
- PHP websites
- Docker containers
- Any project that can be deployed with command-line tools

### Is this free?
Yes! This is open-source software. You only pay for your own server hosting.

## Setup Questions

### What do I need to get started?
- A GitHub repository with your code
- A server (VPS, cloud instance, etc.) where your website runs
- Your code already deployed on that server at least once manually
- Basic command-line access to your server

### Can I use this with multiple websites?
Yes! You can configure multiple repositories and deploy them to different locations on your server.

### Do I need a special type of server?
Any Linux server where you can install Go and run command-line tools will work. This includes:
- VPS providers (DigitalOcean, Linode, Vultr)
- Cloud providers (AWS EC2, Google Cloud, Azure)
- Dedicated servers
- Even a Raspberry Pi!

## Configuration Questions

### How do I generate the webhook secret?
You can use any random string, but for security, generate a strong one:
```bash
openssl rand -hex 20
```
Or use an online password generator to create a 40-character random string.

### What should my deployment commands be?
This depends on your project type:

**Static HTML site:**
```bash
git pull && rsync -av ./ /var/www/html/
```

**Node.js with PM2:**
```bash
git pull && npm install && npm run build && pm2 restart myapp
```

**Python Flask:**
```bash
git pull && pip install -r requirements.txt && systemctl restart myapp
```

### Can I deploy from branches other than 'main'?
Yes! Change the `BRANCH_FILTER` setting:
```bash
BRANCH_FILTER=production
```

### How do I deploy multiple projects?
Add multiple entries to `REPO_MAP` separated by commas:
```bash
REPO_MAP=user/website:~/sites/website,user/api:~/apps/api
```

Then set commands for each:
```bash
COMMANDS_website=git pull && npm run build
COMMANDS_api=git pull && go build && systemctl restart api
```

## Troubleshooting

### The webhook isn't being received
1. **Check your server's firewall** - Make sure port 3000 is open
2. **Verify the webhook URL** - Should be `http://your-server-ip:3000/webhook`
3. **Check the secret** - Must match exactly between GitHub and your `.env` file
4. **Look at GitHub's webhook delivery page** - Shows if GitHub is successfully sending webhooks

### Deployments are failing
1. **Test commands manually** - Run your deployment commands by hand to see if they work
2. **Check permissions** - Make sure the tool can access your website files
3. **Look at the logs** - Check `/var/log/deployer.log` for error details
4. **Verify paths** - Make sure the paths in `REPO_MAP` are correct

### "Permission denied" errors
1. **Check file ownership** - The user running the tool needs access to your website files
2. **Check directory permissions** - Make sure directories are readable/writable
3. **Consider running as the web server user** - Often `www-data` or `nginx`

### The tool stops working after a while
1. **Check if it's still running** - `ps aux | grep cicd-thing`
2. **Look for error messages** - Check the logs for crashes
3. **Set up automatic restart** - Use systemd or supervisor to restart if it crashes
4. **Check server resources** - Make sure you're not running out of memory/disk

## Security Questions

### Is this secure?
Yes, when configured properly:
- GitHub webhooks are verified with HMAC signatures
- API endpoints require authentication
- You can restrict access by IP address
- All communication should use HTTPS in production

### Should I use HTTPS?
Yes! In production, put a reverse proxy (like nginx) in front of this tool to handle HTTPS.

### Can I restrict which IPs can trigger deployments?
Yes! Set the `IP_ALLOWLIST` to GitHub's webhook IP ranges:
```bash
IP_ALLOWLIST=140.82.112.0/20,185.199.108.0/22,192.30.252.0/22
```

### What if someone gets my webhook secret?
Change it immediately in both your `.env` file and GitHub webhook settings. Restart the tool after changing.

## Performance Questions

### How many deployments can this handle?
The tool can handle multiple concurrent deployments (configurable with `CONCURRENCY_LIMIT`). For most small to medium projects, the default settings work fine.

### Will this slow down my server?
Deployments only run when you push code, so there's minimal impact. The tool itself uses very little resources when idle.

### Can I deploy large projects?
Yes, but you might need to increase the `TIMEOUT_SECONDS` setting for projects that take a long time to build.

## Advanced Usage

### Can I get notifications when deployments happen?
Yes! The tool supports notifications to Slack, email, or custom webhooks. Check the configuration documentation for setup details.

### Can I roll back failed deployments?
Yes! Set up `ROLLBACK_COMMANDS` for each project:
```bash
ROLLBACK_COMMANDS_myapp=git reset --hard HEAD~1 && npm run build && pm2 restart myapp
```

### Can I deploy to multiple servers?
Not directly, but you can:
1. Run this tool on multiple servers
2. Use your deployment commands to sync to other servers
3. Use container orchestration tools

### Can I test deployments without affecting my live site?
Yes! Set `DRY_RUN=true` to simulate deployments without actually running them.

## Getting Help

### Where can I find more detailed documentation?
- `README.md` - Complete technical documentation
- `API.md` - API endpoint documentation
- `DEPLOYMENT_EXAMPLES.md` - Real-world configuration examples

### The tool isn't working and I can't figure out why
1. **Check the logs** - Look at both the tool's output and `/var/log/deployer.log`
2. **Test manually** - Try running your deployment commands by hand
3. **Use dry run mode** - Set `DRY_RUN=true` to see what would happen
4. **Check the health endpoint** - Visit `http://your-server:3000/health`

### Can I contribute or report bugs?
Yes! This is open-source software. Check the repository for contribution guidelines and issue reporting.

## Best Practices

### How should I set this up for production?
1. Use HTTPS with a reverse proxy
2. Set up the tool as a system service (systemd)
3. Configure log rotation
4. Set up monitoring and alerts
5. Use strong secrets and IP restrictions
6. Test your rollback procedures

### Should I backup before deployments?
The tool can automatically rollback failed deployments, but having regular backups is always a good idea for your database and important files.

### How often should I update the tool?
Check for updates periodically, especially for security fixes. The tool will log its version in the health endpoint.
