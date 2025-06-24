# Deployment Examples

This document provides real-world examples of how to configure CI/CD Thing for different types of applications.

## Node.js Applications

### Basic Node.js App with PM2

**.env configuration:**
```bash
REPO_MAP=myorg/webapp:~/apps/webapp
COMMANDS_webapp=git pull && npm ci && npm run build && pm2 restart webapp
ROLLBACK_COMMANDS_webapp=git reset --hard HEAD~1 && npm ci && npm run build && pm2 restart webapp
```

### Next.js Application

**.env configuration:**
```bash
REPO_MAP=myorg/nextjs-app:~/apps/nextjs-app
COMMANDS_nextjs-app=git pull && npm ci && npm run build && pm2 restart nextjs-app
ROLLBACK_COMMANDS_nextjs-app=git reset --hard HEAD~1 && npm ci && npm run build && pm2 restart nextjs-app
```

### Node.js API with Database Migrations

**.env configuration:**
```bash
REPO_MAP=myorg/api:~/apps/api
COMMANDS_api=git pull && npm ci && npm run migrate && npm run build && pm2 restart api
ROLLBACK_COMMANDS_api=git reset --hard HEAD~1 && npm ci && npm run migrate:rollback && npm run build && pm2 restart api
```

## Go Applications

### Simple Go Web Server

**.env configuration:**
```bash
REPO_MAP=myorg/go-api:~/apps/go-api
COMMANDS_go-api=git pull && go build -o api . && systemctl restart go-api
ROLLBACK_COMMANDS_go-api=git reset --hard HEAD~1 && go build -o api . && systemctl restart go-api
```

### Go Application with Tests

**.env configuration:**
```bash
REPO_MAP=myorg/go-service:~/apps/go-service
COMMANDS_go-service=git pull && go test ./... && go build -o service . && systemctl restart go-service
ROLLBACK_COMMANDS_go-service=git reset --hard HEAD~1 && go build -o service . && systemctl restart go-service
```

## Python Applications

### Django Application

**.env configuration:**
```bash
REPO_MAP=myorg/django-app:~/apps/django-app
COMMANDS_django-app=git pull && pip install -r requirements.txt && python manage.py migrate && python manage.py collectstatic --noinput && systemctl restart django-app
ROLLBACK_COMMANDS_django-app=git reset --hard HEAD~1 && pip install -r requirements.txt && python manage.py migrate && systemctl restart django-app
```

### Flask API with Gunicorn

**.env configuration:**
```bash
REPO_MAP=myorg/flask-api:~/apps/flask-api
COMMANDS_flask-api=git pull && pip install -r requirements.txt && systemctl restart flask-api
ROLLBACK_COMMANDS_flask-api=git reset --hard HEAD~1 && pip install -r requirements.txt && systemctl restart flask-api
```

## Docker Applications

### Single Container Application

**.env configuration:**
```bash
REPO_MAP=myorg/docker-app:~/apps/docker-app
COMMANDS_docker-app=git pull && docker build -t myapp:latest . && docker stop myapp || true && docker run -d --name myapp -p 8080:8080 myapp:latest
ROLLBACK_COMMANDS_docker-app=git reset --hard HEAD~1 && docker build -t myapp:latest . && docker stop myapp || true && docker run -d --name myapp -p 8080:8080 myapp:latest
```

### Docker Compose Application

**.env configuration:**
```bash
REPO_MAP=myorg/compose-app:~/apps/compose-app
COMMANDS_compose-app=git pull && docker-compose down && docker-compose build && docker-compose up -d
ROLLBACK_COMMANDS_compose-app=git reset --hard HEAD~1 && docker-compose down && docker-compose build && docker-compose up -d
```

## Static Sites

### Hugo Static Site

**.env configuration:**
```bash
REPO_MAP=myorg/hugo-site:~/sites/hugo-site
COMMANDS_hugo-site=git pull && hugo --minify && rsync -av --delete public/ /var/www/html/
ROLLBACK_COMMANDS_hugo-site=git reset --hard HEAD~1 && hugo --minify && rsync -av --delete public/ /var/www/html/
```

### Jekyll Site

**.env configuration:**
```bash
REPO_MAP=myorg/jekyll-site:~/sites/jekyll-site
COMMANDS_jekyll-site=git pull && bundle install && bundle exec jekyll build && rsync -av --delete _site/ /var/www/html/
ROLLBACK_COMMANDS_jekyll-site=git reset --hard HEAD~1 && bundle install && bundle exec jekyll build && rsync -av --delete _site/ /var/www/html/
```

## Multi-Environment Setup

### Staging and Production

**Staging .env:**
```bash
PORT=3001
REPO_MAP=myorg/app:~/staging/app
COMMANDS_app=git pull && npm ci && npm run build:staging && pm2 restart app-staging
BRANCH_FILTER=develop
LOG_FILE=/var/log/deployer-staging.log
```

**Production .env:**
```bash
PORT=3000
REPO_MAP=myorg/app:~/production/app
COMMANDS_app=git pull && npm ci && npm run build:production && pm2 restart app-production
BRANCH_FILTER=main
LOG_FILE=/var/log/deployer-production.log
```

## Advanced Configurations

### Multiple Repositories

**.env configuration:**
```bash
REPO_MAP=myorg/frontend:~/apps/frontend,myorg/backend:~/apps/backend,myorg/docs:~/sites/docs

# Frontend (React)
COMMANDS_frontend=git pull && npm ci && npm run build && pm2 restart frontend

# Backend (Go API)
COMMANDS_backend=git pull && go test ./... && go build -o api . && systemctl restart backend

# Documentation (Hugo)
COMMANDS_docs=git pull && hugo --minify && rsync -av --delete public/ /var/www/docs/

# Rollback commands
ROLLBACK_COMMANDS_frontend=git reset --hard HEAD~1 && npm ci && npm run build && pm2 restart frontend
ROLLBACK_COMMANDS_backend=git reset --hard HEAD~1 && go build -o api . && systemctl restart backend
ROLLBACK_COMMANDS_docs=git reset --hard HEAD~1 && hugo --minify && rsync -av --delete public/ /var/www/docs/
```

### Complex Deployment Pipeline

**.env configuration:**
```bash
REPO_MAP=myorg/complex-app:~/apps/complex-app
COMMANDS_complex-app=git pull && npm ci && npm run test && npm run lint && npm run build && docker build -t complex-app:latest . && docker-compose down && docker-compose up -d && npm run smoke-test
ROLLBACK_COMMANDS_complex-app=git reset --hard HEAD~1 && docker build -t complex-app:latest . && docker-compose down && docker-compose up -d
TIMEOUT_SECONDS=600
```

## Security Examples

### IP Allowlisting

```bash
# Allow only specific IPs
IP_ALLOWLIST=192.168.1.100,10.0.0.50

# Allow IP ranges
IP_ALLOWLIST=192.168.1.0/24,10.0.0.0/8

# GitHub webhook IPs (example)
IP_ALLOWLIST=140.82.112.0/20,185.199.108.0/22,192.30.252.0/22
```

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
