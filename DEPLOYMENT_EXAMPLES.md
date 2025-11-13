# Deployment Examples

This document provides real-world examples of how to configure CICD-Thing using its TOML-based configuration for different types of applications.

## Node.js Applications

### Basic Node.js App with PM2

**config.toml configuration:**
```toml
[repositories]
"myorg/webapp" = "~/apps/webapp"

[commands]
"webapp" = "git pull && npm ci && npm run build && pm2 restart webapp"

# Note: Built-in rollback configuration is planned for a future version. For now, implement any rollback behavior in your own scripts or workflows.
```

### Next.js Application

**config.toml configuration:**
```toml
REPO_MAP=myorg/nextjs-app:~/apps/nextjs-app
COMMANDS_nextjs-app=git pull && npm ci && npm run build && pm2 restart nextjs-app
# (example) Manual rollback strategy could reset to previous commit and redeploy via your own scripts
```

### Node.js API with Database Migrations

**config.toml configuration:**
```toml
REPO_MAP=myorg/api:~/apps/api
COMMANDS_api=git pull && npm ci && npm run migrate && npm run build && pm2 restart api
# (example) Manual rollback strategy: reset to previous commit, run migrate:rollback, rebuild, and restart
```

## Go Applications

### Simple Go Web Server

**config.toml configuration:**
```toml
REPO_MAP=myorg/go-api:~/apps/go-api
COMMANDS_go-api=git pull && go build -o api . && systemctl restart go-api
# Example manual rollback command you might run yourself:
# git reset --hard HEAD~1 && go build -o api . && systemctl restart go-api
```

### Go Application with Tests

**config.toml configuration:**
```toml
REPO_MAP=myorg/go-service:~/apps/go-service
COMMANDS_go-service=git pull && go test ./... && go build -o service . && systemctl restart go-service
# Example manual rollback command you might run yourself:
# git reset --hard HEAD~1 && go build -o service . && systemctl restart go-service
```

## Python Applications

### Django Application

**config.toml configuration:**
```toml
REPO_MAP=myorg/django-app:~/apps/django-app
COMMANDS_django-app=git pull && pip install -r requirements.txt && python manage.py migrate && python manage.py collectstatic --noinput && systemctl restart django-app
# Example manual rollback command you might run yourself:
# git reset --hard HEAD~1 && pip install -r requirements.txt && python manage.py migrate && systemctl restart django-app
```

### Flask API with Gunicorn

**config.toml configuration:**
```toml
REPO_MAP=myorg/flask-api:~/apps/flask-api
COMMANDS_flask-api=git pull && pip install -r requirements.txt && systemctl restart flask-api
# Example manual rollback command you might run yourself:
# git reset --hard HEAD~1 && pip install -r requirements.txt && systemctl restart flask-api
```

## Docker Applications

### Single Container Application

**config.toml configuration:**
```toml
REPO_MAP=myorg/docker-app:~/apps/docker-app
COMMANDS_docker-app=git pull && docker build -t myapp:latest . && docker stop myapp || true && docker run -d --name myapp -p 8080:8080 myapp:latest
# Example manual rollback command you might run yourself:
# git reset --hard HEAD~1 && docker build -t myapp:latest . && docker stop myapp || true && docker run -d --name myapp -p 8080:8080 myapp:latest
```

### Docker Compose Application

**config.toml configuration:**
```toml
REPO_MAP=myorg/compose-app:~/apps/compose-app
COMMANDS_compose-app=git pull && docker-compose down && docker-compose build && docker-compose up -d
# Example manual rollback command you might run yourself:
# git reset --hard HEAD~1 && docker-compose down && docker-compose build && docker-compose up -d
```

## Static Sites

### Hugo Static Site

**config.toml configuration:**
```toml
REPO_MAP=myorg/hugo-site:~/sites/hugo-site
COMMANDS_hugo-site=git pull && hugo --minify && rsync -av --delete public/ /var/www/html/
# Example manual rollback command you might run yourself:
# git reset --hard HEAD~1 && hugo --minify && rsync -av --delete public/ /var/www/html/
```

### Jekyll Site

**config.toml configuration:**
```toml
REPO_MAP=myorg/jekyll-site:~/sites/jekyll-site
COMMANDS_jekyll-site=git pull && bundle install && bundle exec jekyll build && rsync -av --delete _site/ /var/www/html/
# Example manual rollback command you might run yourself:
# git reset --hard HEAD~1 && bundle install && bundle exec jekyll build && rsync -av --delete _site/ /var/www/html/
```

## Multi-Environment Setup

### Staging and Production

**Staging config.toml snippet:**
```bash
PORT=3001
REPO_MAP=myorg/app:~/staging/app
COMMANDS_app=git pull && npm ci && npm run build:staging && pm2 restart app-staging
BRANCH_FILTER=develop
LOG_FILE=/var/log/deployer-staging.log
```

**Production config.toml snippet:**
```bash
PORT=3000
REPO_MAP=myorg/app:~/production/app
COMMANDS_app=git pull && npm ci && npm run build:production && pm2 restart app-production
BRANCH_FILTER=main
LOG_FILE=/var/log/deployer-production.log
```

## Advanced Configurations

### Multiple Repositories

**config.toml configuration:**
```toml
REPO_MAP=myorg/frontend:~/apps/frontend,myorg/backend:~/apps/backend,myorg/docs:~/sites/docs

# Frontend (React)
COMMANDS_frontend=git pull && npm ci && npm run build && pm2 restart frontend

# Backend (Go API)
COMMANDS_backend=git pull && go test ./... && go build -o api . && systemctl restart backend

# Documentation (Hugo)
COMMANDS_docs=git pull && hugo --minify && rsync -av --delete public/ /var/www/docs/

# Example manual rollback commands you might run yourself:
# git reset --hard HEAD~1 && npm ci && npm run build && pm2 restart frontend
# git reset --hard HEAD~1 && go build -o api . && systemctl restart backend
# git reset --hard HEAD~1 && hugo --minify && rsync -av --delete public/ /var/www/docs/
```

### Complex Deployment Pipeline

**config.toml configuration:**
```toml
REPO_MAP=myorg/complex-app:~/apps/complex-app
COMMANDS_complex-app=git pull && npm ci && npm run test && npm run lint && npm run build && docker build -t complex-app:latest . && docker-compose down && docker-compose up -d && npm run smoke-test
# Example manual rollback command you might run yourself:
# git reset --hard HEAD~1 && docker build -t complex-app:latest . && docker-compose down && docker-compose up -d
timeout_seconds = 600
```

## Security Examples

### Network & Auth Controls

- Use a strong `webhook_secret` to validate GitHub requests.
- Use a strong `api_key` to protect the `/deploy` endpoint.
- Use a reverse proxy or firewall (Nginx, Caddy, cloud WAF, security groups, etc.) to restrict access to `/webhook`, `/deploy`, and `/logs`.


### Webhook Security

```bash
# Generate strong webhook secret
WEBHOOK_SECRET=$(openssl rand -hex 20)

# Generate strong API key
API_KEY=$(openssl rand -hex 32)
```

## Monitoring Examples

### Health Check Script

```bash
#!/bin/bash
# health-check.sh

HEALTH_URL="http://localhost:3000/health"
RESPONSE=$(curl -s "$HEALTH_URL")
STATUS=$(echo "$RESPONSE" | jq -r '.status')

if [ "$STATUS" = "healthy" ]; then
    echo "Service is healthy"
    exit 0
else
    echo "Service is unhealthy: $RESPONSE"
    exit 1
fi
```

### Log Monitoring

```bash
# Monitor deployment logs
tail -f /var/log/deployer.log | grep -E "(FAILED|ERROR|ROLLBACK)"

# Count successful deployments today
grep "$(date +%Y-%m-%d)" /var/log/deployer.log | grep "SUCCESS" | wc -l
```

## Troubleshooting Examples

### Debug Failed Deployment

1. **Check logs:**
   ```bash
   tail -100 /var/log/deployer.log
   ```

2. **Test commands manually:**
   ```bash
   cd ~/apps/myapp
   git pull && npm ci && npm run build
   ```

3. **Check permissions:**
   ```bash
   ls -la ~/apps/myapp
   whoami
   ```

### Test Webhook Locally

```bash
# Simulate GitHub webhook
curl -X POST http://localhost:3000/webhook \
  -H "Content-Type: application/json" \
  -H "X-GitHub-Event: push" \
  -H "X-Hub-Signature-256: sha256=your_signature" \
  -d @webhook-payload.json
```

## Best Practices

1. **Always test commands manually first**
2. **Use specific versions in package.json/go.mod**
3. **Include health checks in deployment commands**
4. **Set appropriate timeouts for long-running builds**
5. **Use rollback commands that are fast and reliable**
6. **Monitor logs regularly**
7. **Test rollback procedures**
8. **Keep deployment commands idempotent**
