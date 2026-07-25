# Co-Drive

Server-rendered vehicle fleet management — Go + HTMX + PostgreSQL.

## Quick Start

```bash
cp .env.example .env      # then edit with your values
docker compose up -d db    # start PostgreSQL
go mod download
go run ./cmd/server        # start server (http://localhost:8080)
```

Dev login: set `DEV_LOGIN_ENABLED=true` in `.env`, then click the dev bypass button on the login page.

## Project Structure

```
cmd/server/main.go          # Entry point
internal/
├── auth/                   # Session & OAuth2
├── db/                     # SQL queries + migrations
├── expiry/                 # Expiry status logic
├── handler/                # HTTP handlers + templates (embedded)
├── middleware/             # Auth, security, logging
└── svg/                    # Server-rendered charts
static/                     # CSS, JS, icons, manifest
migrations/                 # SQL migrations (idempotent)
```

> **Note:** Templates live in `internal/handler/templates/` and are embedded into the binary at compile time via `//go:embed`. Edit templates there — there is no separate `templates/` directory at the root.

## Dependencies (3)

| Library | Purpose |
|---------|---------|
| `pgx/v5` | PostgreSQL driver |
| `oauth2` | Google OAuth |
| `webpush-go` | VAPID push notifications |

No ORM, framework, or query builder.

## Features

- Google OAuth login, session management
- Dashboard with fleet stats + alerts
- Vehicle CRUD, mileage/fuel logging
- PUC & insurance expiry tracking
- Vehicle sharing (HTMX partials)
- Dark mode, PWA, push notifications

## Deployment

See [deployment.md](deployment.md) for Railway, Fly.io, and VPS guides.

## License

MIT
