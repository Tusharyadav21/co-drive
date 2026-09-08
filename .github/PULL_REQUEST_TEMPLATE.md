## Description

Provide a clear and concise description of the changes introduced in this pull request.

Fixes / Closes #(issue)

---

## Type of Change

- [ ] 🐛 Bug fix (non-breaking change which fixes an issue)
- [ ] ✨ New feature (non-breaking change which adds functionality)
- [ ] ♻️ Refactoring / Code quality (no functional change)
- [ ] ⚡ Performance improvement
- [ ] 📝 Documentation update
- [ ] 🔒 Security fix
- [ ] 🔧 Build / CI / Tooling

---

## Key Subsystems Affected

- [ ] Authentication & Sessions (`internal/auth`, `internal/session`)
- [ ] Vehicles, Documents & Mileage (`internal/vehicles`)
- [ ] Users & Driver Profile (`internal/users`)
- [ ] Database Schema & Migrations (`cmd/migrate`, `pkg/database`)
- [ ] Frontend UI & Styling (`static/`)
- [ ] HTTP Routing & Middleware (`cmd/server`, `internal/middleware`)

---

## Verification & Quality Checklist

Before submitting this PR, please verify the following:

- [ ] **Verification Gate**: `make check` passes locally with exit code 0 (`go vet`, `go test -v ./...`, and `make build`).
- [ ] **Pure-Go PostgreSQL**: No CGO dependencies introduced; pure `pgx/v5` driver used.
- [ ] **Sub-Resource Ownership**: Authorization verified so users can only access and modify their own assets.
- [ ] **Standard JSON Envelope**: All JSON API responses adhere to the standard envelope (`data` / `error`).
- [ ] **Security**: No secrets or private environment configs committed (`.env` guarded).
- [ ] **Documentation**: Updated relevant documentation in `docs/` or `CONTEXT.md` if domain models or invariants changed.
