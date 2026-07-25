# Deployment Guide

## Prerequisites

- PostgreSQL 16+
- Go 1.22+ (for building)
- Google OAuth credentials (Google Cloud Console → APIs & Services → Credentials → OAuth 2.0 Client ID)
- VAPID keys for push notifications (optional): `npx web-push generate-vapid-keys`

## Environment Variables

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `PORT` | No | `8080` | HTTP listen port |
| `DATABASE_URL` | Yes | — | PostgreSQL connection string |
| `SESSION_SECRET` | Yes | — | 32+ char random secret |
| `COOKIE_DOMAIN` | No | `""` | Cookie domain (empty for localhost) |
| `GOOGLE_CLIENT_ID` | Yes | — | Google OAuth client ID |
| `GOOGLE_CLIENT_SECRET` | Yes | — | Google OAuth client secret |
| `GOOGLE_REDIRECT_URL` | Yes | — | Must match Google Console exactly |
| `VAPID_PUBLIC_KEY` | Push | — | VAPID public key |
| `VAPID_PRIVATE_KEY` | Push | — | VAPID private key |
| `VAPID_SUBJECT` | Push | `mailto:admin@codrive.app` | VAPID contact |
| `DEV_LOGIN_ENABLED` | No | `false` | Set `true` for dev bypass |

Generate session secret: `openssl rand -hex 32`

## Railway

### Setup

```bash
# Install CLI
npm i -g @railway/cli

# Login
railway login

# Link project (or create new)
railway init
```

### PostgreSQL

```bash
# Add PostgreSQL plugin
railway plugins add postgresql

# Railway sets DATABASE_URL automatically
```

### Deploy

```bash
# Set secrets (one-time)
railway secrets set SESSION_SECRET="$(openssl rand -hex 32)"
railway secrets set GOOGLE_CLIENT_ID="..."
railway secrets set GOOGLE_CLIENT_SECRET="..."
railway secrets set GOOGLE_REDIRECT_URL="https://your-app.railway.app/auth/google/callback"
railway secrets set COOKIE_DOMAIN=".railway.app"

# Deploy
railway up

# Open
railway open
```

### Custom Domain

```bash
railway domain add your-domain.com
```

Update `GOOGLE_REDIRECT_URL` and `COOKIE_DOMAIN` after setting domain.

## Fly.io

### Setup

```bash
# Install flyctl
curl -fsSL https://fly.io/install.sh | sh

# Login
fly auth login

# Launch
fly launch --no-deploy
```

### PostgreSQL

```bash
fly postgres create --name codrive-db
fly postgres attach codrive-db
# Fly sets DATABASE_URL automatically
```

### Deploy

```bash
fly secrets set SESSION_SECRET="$(openssl rand -hex 32)"
fly secrets set GOOGLE_CLIENT_ID="..."
fly secrets set GOOGLE_CLIENT_SECRET="..."
fly secrets set GOOGLE_REDIRECT_URL="https://your-app.fly.dev/auth/google/callback"
fly secrets set COOKIE_DOMAIN=".fly.dev"
fly deploy

fly open
```

### fly.toml

```toml
app = "codrive"
primary_region = "iad"

[build]
  dockerfile = "Dockerfile"

[http_service]
  internal_port = 8080
  force_https = true
  auto_stop_machines = false
  auto_start_machines = true
  min_machines_running = 0
  processes = ["app"]

[[services]]
  port = 443
  protocol = "tcp"
  internal_port = 8080

  [[services.ports]]
    port = 80
    handlers = ["http"]
    force_https = true

  [[services.ports]]
    port = 443
    handlers = ["tls", "http"]
```

## Manual VPS

### Build

```bash
CGO_ENABLED=0 GOOS=linux go build -o codrive ./cmd/server
```

### Files

```
/opt/codrive/
├── codrive          # binary
├── static/          # static assets (must be alongside binary)
│   ├── css/
│   ├── js/
│   ├── icons/
│   ├── sw.js
│   └── site.webmanifest
└── .env             # environment variables
```

### Systemd Service

`/etc/systemd/system/codrive.service`:

```ini
[Unit]
Description=Co-Drive Fleet Management
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
User=codrive
Group=codrive
WorkingDirectory=/opt/codrive
EnvironmentFile=/opt/codrive/.env
ExecStart=/opt/codrive/codrive
Restart=always
RestartSec=5
LimitNOFILE=65536

[Install]
WantedBy=multi-user.target
```

```bash
sudo systemctl daemon-reload
sudo systemctl enable --now codrive
sudo systemctl status codrive
```

### Nginx Reverse Proxy

```nginx
server {
    listen 80;
    server_name codrive.example.com;
    return 301 https://$server_name$request_uri;
}

server {
    listen 443 ssl http2;
    server_name codrive.example.com;

    ssl_certificate /etc/letsencrypt/live/codrive.example.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/codrive.example.com/privkey.pem;

    location / {
        proxy_pass http://127.0.0.1:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

## Health Check

```bash
# Login page (public)
curl -s -o /dev/null -w "%{http_code}" http://localhost:8080/auth/login

# Should return 200
```

## Logs

The app logs structured JSON to stdout:

```json
{"time":"...","level":"INFO","msg":"Starting server","port":"8080"}
{"time":"...","level":"INFO","msg":"HTTP request","method":"GET","path":"/auth/login","status":200,"duration":466625,...}
```

Collect with any log shipper (Vector, Logtail, etc.) or platform built-in (Railway Logs, Fly Logs).

## Expiry Check

A goroutine runs every 24h querying PUC (7-day) and insurance (15-day) expiry windows. Errors are logged at ERROR level. No email/push sending is implemented yet — the goroutine is a placeholder for future notification integrations.

## Database

Migrations run automatically on startup. Idempotent (`CREATE TABLE IF NOT EXISTS`). To reset:

```bash
# Railway
railway plugins delete postgresql
railway plugins add postgresql

# Fly
fly postgres reset codrive-db

# Docker
docker compose down -v && docker compose up -d
```
