# Co-Drive

Lightweight, high-performance Go web backend powered by **PostgreSQL** (Neon-ready, Supabase, or local) with passwordless **Email OTP** authentication and driver profile management.

---

## Architecture & Design

Co-Drive follows standard Go clean architecture practices:

- **Database**: Pure-Go PostgreSQL (`github.com/jackc/pgx/v5`, zero CGO dependency required) with production connection pooling (`MaxOpenConns: 25`, `MaxIdleConns: 5`), native `BOOLEAN`/`TIMESTAMPTZ` types, and foreign key cascades.
- **Provider Independence**: Connect to Neon, Supabase, AWS RDS, GCP Cloud SQL, or local PostgreSQL simply by providing `DATABASE_URL` in `.env`.
- **Authentication & Profile System**:
  - **Passwordless Email OTP**: 6-digit numeric OTP generation, SHA-256 token hashing, email dispatch, and single-use validation. Automatic display name derivation from email address.
  - **Driver Profile & Details**: Dedicated driver profile management page (`/profile`) backed by the relational `user_details` table (`GET/PUT /api/users/me`).
  - **Session & CSRF**: PostgreSQL session storage, secure HTTP-only cookies, and double-submit CSRF cookie protection middleware.
- **Middlewares**: Centralized request logging, sliding-window rate limiting, CSRF protection, and session authentication under `internal/middleware/`.
- **Reusable Packages (`pkg/`)**: Standalone, decoupled utility libraries (`database`, `email`, `response`).

---

## How to Build & Run

### Available `make` Targets

| Target | Description |
| :--- | :--- |
| `make build` | Compiles `bin/server` and `bin/migrate` binaries |
| `make migrate` | Applies embedded database schema migrations up |
| `make migrate-down` | Rolls back database schema migrations down |
| `make dev` | Applies migrations up and starts dev server (`http://localhost:8080`) |
| `make run` | Starts the compiled server binary |
| `make test` | Executes all unit and integration tests across packages |
| `make check` | Full verification gate (`go vet` + `go test` + binary builds) |
| `make clean` | Removes compiled binaries (`bin/`) |

### Quickstart Command

```bash
# 1. Ensure DATABASE_URL is set in .env (defaults to postgres://localhost:5432/co-drive?sslmode=disable)
cp .env.example .env

# 2. Apply migrations & launch dev server in one step
make dev
```

---

## Directory Layout

```text
co-drive/
├── bin/                       # Compiled application binaries (server, migrate)
├── cmd/
│   ├── migrate/main.go        # Database migration CLI entrypoint
│   └── server/main.go         # HTTP server entrypoint
├── config/
│   ├── config.go              # Canonical configuration loader and ENV reader
│   └── config_test.go         # Configuration unit tests and AST isolation check
├── docs/
│   ├── Architecture.md        # Technical specifications & topology
│   ├── Decisions.md           # Architecture Decision Records (ADRs)
│   ├── PRD.md                 # Product scope & requirements
│   ├── Runbook.md             # Operations, deployment & troubleshooting
│   └── openapi.yaml           # OpenAPI 3.0 API specification
├── internal/
│   ├── auth/                  # Domain authentication (Password, OTP, Session)
│   ├── middleware/            # HTTP middlewares (Logging, Rate Limiting, CSRF, Auth)
│   ├── session/               # Session context extraction helpers
│   ├── users/                 # User repository & profile model
│   └── vehicles/              # Vehicles, mileage, documents, and derived calculations
├── pkg/
│   ├── database/              # PostgreSQL driver, connection pool & migrations runner
│   ├── email/                 # Transactional SMTP email sender & HTML templates
│   └── response/              # Standardized JSON response helpers
├── Dockerfile                 # Multi-stage production Docker build
├── docker-compose.yml         # Local Docker setup with PostgreSQL
├── Makefile                   # Build & migration shortcuts
└── README.md
```

---

## API Documentation & Endpoints

See **[docs/openapi.yaml](docs/openapi.yaml)** for the full OpenAPI 3.0 specification.

| Endpoint | Method | Purpose |
| :--- | :--- | :--- |
| `/health` | `GET` | Server healthcheck & CSRF token cookie initialization |
| `/profile` | `GET` | Driver Profile UI management page |
| `/api/auth/otp/request` | `POST` | Request 6-digit Email OTP code |
| `/api/auth/otp/verify` | `POST` | Verify Email OTP code & issue session |
| `/api/users/me` | `GET`, `PUT` | Get & update logged-in user profile and details (Protected) |
| `/api/vehicles` | `GET`, `POST` | List and create vehicles (Protected) |
| `/api/vehicles/{id}` | `GET`, `DELETE` | View and delete vehicle (Protected) |
| `/api/vehicles/{id}/mileage` | `POST` | Add mileage entry (Protected) |
| `/api/vehicles/{id}/puc` | `POST` | Upsert PUC certificate (Protected) |
| `/api/vehicles/{id}/insurance`| `POST` | Upsert insurance policy (Protected) |
| `/api/auth/logout` | `POST` | Revoke session & clear cookies (Protected) |

---

## Deployment

A multi-stage `Dockerfile` is provided for containerized deployments:
```bash
docker build -t co-drive:latest .
docker run -p 8080:8080 -e DATABASE_URL="postgresql://user:pass@ep-xyz.neon.tech/neondb?sslmode=require" co-drive:latest
```

---

## License

MIT
