# Architecture Decision Records (ADRs) (`docs/Decisions.md`)

This document records the architectural and engineering decisions governing **Co-Drive**. All future AI agents and engineers proposing structural or boundary changes **MUST** register a new ADR prior to implementation.

---

## 📋 ADR Register

| ADR ID | Title | Status | Date |
| :--- | :--- | :--- | :--- |
| [ADR-001](#adr-001-pure-go-sqlite-architecture-with-wal--in-memory-derived-calculations) | Pure-Go SQLite Architecture with WAL & In-Memory Derived Calculations | **Superseded by ADR-007** | 2026-08-31 |
| [ADR-002](#adr-002-lock-free-database-snapshots--automated-google-drive-cloud-backups) | Lock-Free Database Snapshots & Automated Google Drive Cloud Backups | **Superseded by ADR-007** | 2026-08-31 |
| [ADR-003](#adr-003-multi-factor-authentication-suite--secure-cookie-sessions) | Multi-Factor Authentication Suite & Secure Cookie Sessions | **Accepted** | 2026-08-31 |
| [ADR-004](#adr-004-strict-sub-resource-ownership-enforcement--anti-probing) | Strict Sub-Resource Ownership Enforcement & Anti-Probing | **Accepted** | 2026-08-31 |
| [ADR-005](#adr-005-unified-codegraph-first-multi-agent-workflow--quality-gates) | Unified CodeGraph-First Multi-Agent Workflow & Quality Gates | **Accepted** | 2026-08-31 |
| [ADR-006](#adr-006-centralized-environment-configuration-via-configgo) | Centralized Environment Configuration via config.go | **Accepted** | 2026-09-04 |
| [ADR-007](#adr-007-migration-from-sqlite--google-drive-to-postgresql-neon-ready) | Migration from SQLite & Google Drive to PostgreSQL (Neon-Ready) | **Accepted** | 2026-09-04 |

---

## 📝 Standard ADR Template

Future ADR proposals must copy and fill the following template:

```markdown
## ADR-XXX: [Title]
- **Date**: YYYY-MM-DD
- **Status**: [Proposed | Accepted | Superseded by ADR-YYY | Deprecated]
- **Context**: [What problem are we solving? Why now?]
- **Decision**: [What did we decide? Include technical tradeoffs considered.]
- **Consequences**:
  - **Positive**: [Benefits and performance wins]
  - **Negative / Risks**: [Tradeoffs, operational overhead, migration requirements]
  - **Compliance Invariants**: [Invariants that future agents and code changes must not break]
```

---

## ADR-001: Pure-Go SQLite Architecture with WAL & In-Memory Derived Calculations

- **Date**: 2026-08-31
- **Status**: Superseded by ADR-007
- **Context**: Co-Drive required an ultra-lightweight storage backend without requiring an external database cluster daemon.
- **Decision**: Adopt pure-Go SQLite with Write-Ahead Logging and compute derived metrics dynamically in memory.
- **Superseded By**: ADR-007 migrated storage to PostgreSQL (Neon-ready) for scalable multi-instance deployment.

---

## ADR-002: Lock-Free Database Snapshots & Automated Google Drive Cloud Backups

- **Date**: 2026-08-31
- **Status**: Superseded by ADR-007
- **Context**: SQLite file durability required off-site backup.
- **Decision**: SQLite `VACUUM INTO` snapshots uploaded periodically to Google Drive.
- **Superseded By**: ADR-007 retired Google Drive snapshotting in favor of native managed PostgreSQL cloud durability (Neon point-in-time recovery and branchable databases).

---

## ADR-003: Multi-Factor Authentication Suite & Secure Cookie Sessions

- **Date**: 2026-08-31
- **Status**: Accepted
- **Context**: Modern authentication requires secure credential handling, email verification, and resilient password reset workflows.
- **Decision**: Implement Bcrypt (cost 12) for passwords, SHA-256 for one-time tokens (OTP and reset tokens), and cryptographically random session IDs stored in `HttpOnly`, `SameSite=Lax` cookies.
- **Consequences**:
  - **Positive**: Defense-in-depth against credential theft, session fixation, and brute-force attacks.
  - **Negative / Risks**: Requires working email configuration (SMTP) in production.
  - **Compliance Invariants**: Never store plain-text passwords or reset tokens. Always enforce session expiration.

---

## ADR-004: Strict Sub-Resource Ownership Enforcement & Anti-Probing

- **Date**: 2026-08-31
- **Status**: Accepted
- **Context**: Multi-tenant architectures risk horizontal privilege escalation if users can infer or access other tenants' vehicle records via brute-forcing resource IDs.
- **Decision**: Enforce ownership checks on all sub-resource routes (`ownsVehicle`). Return `404 Not Found` (never `403 Forbidden`) when an unowned resource is requested.
- **Consequences**:
  - **Positive**: Eliminates ID enumeration vulnerabilities (IDOR).
  - **Compliance Invariants**: Any handler taking a parent vehicle ID must verify the requesting user's ownership before performing reads or mutations.

---

## ADR-005: Unified CodeGraph-First Multi-Agent Workflow & Quality Gates

- **Date**: 2026-08-31
- **Status**: Accepted
- **Context**: AI agent operations across multiple handoffs require deterministic context gathering and rigorous verification to prevent regressions.
- **Decision**: Mandate the 6-phase engineering lifecycle and CodeGraph exploration protocol defined in `.agents/workflows/unified-agent-workflow.md`.
- **Consequences**:
  - **Positive**: Drastically reduced hallucination and context fragmentation. Zero broken builds across agent handoffs.
  - **Compliance Invariants**: Always run CodeGraph exploration before editing files; always pass `make check` before handoff.

---

## ADR-006: Centralized Environment Configuration via config.go

- **Date**: 2026-09-04
- **Status**: Accepted
- **Context**: Environment variables and configuration values were fragmented or risked being directly accessed via `os.Getenv` throughout different packages.
- **Decision**: Establish `config/config.go` (`co-drive/config`) as the single canonical location for loading, validating, and overriding application configuration.
- **Consequences**:
  - **Positive**: Centralized configuration management, deterministic testability, strict enforcement via AST unit tests preventing unauthorized `os.Getenv` calls.
  - **Compliance Invariants**: Never call `os.Getenv`, `os.LookupEnv`, or `os.Environ` outside `config/config.go`.

---

## ADR-007: Migration from SQLite & Google Drive to PostgreSQL (Neon-Ready)

- **Date**: 2026-09-04
- **Status**: Accepted
- **Context**: The project transitioned from single-node local SQLite storage to managed cloud PostgreSQL (Neon hosted database) while supporting local development against local PostgreSQL daemons. SQLite-specific snapshotting and Google Drive backup logic created unnecessary operational overhead and third-party cloud dependencies.
- **Decision**:
  1. Adopt `github.com/jackc/pgx/v5` (`stdlib`) as the pure-Go PostgreSQL driver with production connection pooling (`MaxOpenConns: 25`, `MaxIdleConns: 5`).
  2. Centralize database connectivity under a single standard `DATABASE_URL` (supporting Neon `sslmode=require`, local `sslmode=disable`, or any Postgres provider).
  3. Completely remove the Google Drive backup engine (`pkg/gdrive/`), related configuration keys, and heavy Google API dependencies.
  4. Migrate database schemas and SQL queries from SQLite syntax (`?`) to PostgreSQL syntax (`$1, $2...`, native `BOOLEAN`, `TIMESTAMPTZ`, `DOUBLE PRECISION`).
  5. Provide schema-isolated testing via temporary `search_path` namespaces so integration test suites run with 100% concurrency safety.
- **Consequences**:
  - **Positive**: Provider-agnostic PostgreSQL support (Neon, Supabase, RDS, local); scalable multi-connection pool; streamlined dependency footprint (`go.mod` reduced by ~30 packages); zero vendor lock-in.
  - **Negative / Risks**: Requires a running PostgreSQL instance for end-to-end integration tests and production deployment.
  - **Compliance Invariants**: All SQL queries must use PostgreSQL `$1, $2...` parameter notation. Connection configurations must come solely through `config.Config.Database.URL`.
