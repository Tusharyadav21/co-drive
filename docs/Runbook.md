# Operations & Runbook (`docs/Runbook.md`)

> **Application**: Co-Drive  
> **Environment**: Development, Staging, Production  
> **Target Audience**: SREs, System Operators, DevOps Engineers, Autonomous AI Agents

---

## 1. Quickstart & Local Development

### Prerequisites
- Go 1.25+
- PostgreSQL (Local daemon, Docker container, or Neon/Supabase cloud instance)
- Make

```bash
# 1. Run migrations and launch the dev server on localhost:8080
make dev

# 2. Execute verification gate (vet + tests + build)
make check
```

---

## 2. Environment Variables & Configuration

Co-Drive reads configuration from `config/config.go` with `.env` and system environment variable support:

| Environment Variable | Required | Default Value | Description |
| :--- | :--- | :--- | :--- |
| `SERVER_HOST` / `HOST` | No | `0.0.0.0` | HTTP listen host interface |
| `SERVER_PORT` / `PORT` | No | `8080` | HTTP listen port |
| `DATABASE_URL` | No | `postgres://localhost:5432/co-drive?sslmode=disable` | Standard PostgreSQL connection URL (Neon, Supabase, RDS, local) |
| `OTP_TTL_MINUTES` | No | `10` | Email OTP expiration window in minutes |
| `PASSWORD_RESET_TTL_MINUTES` | No | `15` | Password reset token expiration window in minutes |
| `SMTP_HOST` | Email | `""` | SMTP relay server host (e.g. `smtp.gmail.com`) |
| `SMTP_PORT` | Email | `587` | SMTP relay server port |
| `SMTP_USER` | Email | `""` | SMTP username |
| `SMTP_PASS` | Email | `""` | SMTP password |
| `SMTP_FROM` | Email | `no-reply@co-drive.app` | Outgoing email sender address |

---

## 3. Database Operations & Migrations

### Connecting to Neon or Remote Hosted PostgreSQL
To use a Neon serverless PostgreSQL database, paste the connection string into `.env`:
```env
DATABASE_URL=postgresql://[user]:[password]@[endpoint].neon.tech/neondb?sslmode=require
```

### Applying Migrations
Migrations are embedded into the Go binary and automatically executed up on application startup. To run migrations explicitly via CLI:

```bash
# Apply pending schema migrations
make migrate-up

# Roll back last schema migration
make migrate-down
```

### Direct PostgreSQL Inspection
To inspect the database using `psql`:
```bash
psql -d "co-drive" -c "\dt"
psql -d "co-drive" -c "SELECT id, email, full_name FROM users;"
```

---

## 4. Production Deployment

### 4.1 Docker Container Deployment
Co-Drive includes a `docker-compose.yml` that provisions the Go app along with an optional local PostgreSQL 16 database:

```bash
# Launch app + PostgreSQL in background
docker compose up -d

# Check container logs
docker compose logs -f app
```

Or deploy as a standalone container connected to an external cloud database (e.g. Neon):
```bash
docker run -d \
  --name codrive \
  -p 8080:8080 \
  -e DATABASE_URL="postgresql://user:pass@ep-xyz.neon.tech/neondb?sslmode=require" \
  co-drive:latest
```

### 4.2 Systemd VPS Service
`/etc/systemd/system/codrive.service`:
```ini
[Unit]
Description=Co-Drive Fleet Management
After=network.target

[Service]
Type=simple
User=codrive
Group=codrive
WorkingDirectory=/opt/codrive
EnvironmentFile=/opt/codrive/.env
ExecStart=/opt/codrive/bin/server
Restart=always
RestartSec=5
LimitNOFILE=65536

[Install]
WantedBy=multi-user.target
```

---

## 5. Observability & Troubleshooting

### Health Check Endpoint
```bash
curl -I http://localhost:8080/health
# Expected: HTTP/1.1 200 OK
```

### Common Failure Modes
| Symptom | Root Cause | Remediation |
| :--- | :--- | :--- |
| `failed to ping postgresql database` | PostgreSQL daemon down or bad `DATABASE_URL` | Verify PostgreSQL is running on port 5432 or check credentials in `DATABASE_URL`. |
| `pq: SSL is not enabled on the server` / SSL error | Wrong SSL mode in connection string | Use `sslmode=disable` for local dev; use `sslmode=require` for Neon/cloud. |
| `CSRF token mismatch` | Client request missing `X-CSRF-Token` header on POST | Ensure frontend client reads `csrf_token` cookie and attaches `X-CSRF-Token` header. |
