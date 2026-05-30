# CLAUDE.md

This file is the Claude Code entry point for the guardus use-case. It delegates
to AGENTS.md via the @-import below so that the same guidance applies to all
agent types (Claude Code, Codex, Cursor, etc.) without duplicating content.
AGENTS.md is the authoritative source; edit it, not this file.

---

## guardus — Gatus v1 reimplementation on plumego

**What it is:** Feature-equivalent v1 backend of [Gatus](https://github.com/TwiN/gatus),
rebuilt on plumego stable roots + extensions. Sits alongside `use-cases/gatus`
to validate plumego in monitoring/observability product scenarios.

**Module:** `guardus` (own `go.mod`, `replace` to root plumego)

**Status:** v1 complete — HTTP/TCP/ICMP/DNS/TLS/gRPC/WebSocket probes, SQLite+MySQL+memory
storage (dialect-based), Basic Auth, 5 alert providers, maintenance windows,
Prometheus metrics, embedded React SPA with lazy-loaded routes.

---

## Quick start

```bash
# Build (includes frontend)
make build

# Run
make run

# Test (SQLite + memory)
make test

# Test with MySQL (opt-in)
GUARDUS_TEST_MYSQL_DSN="guardus:test@tcp(127.0.0.1:3306)/guardus_test?parseTime=true" \
  go test ./internal/storage -v

# Frontend dev
cd web/app && npm run dev
```

## Key commands

```bash
make build              # Build Go binary + embedded SPA
make run                # Build and run
make test               # go test ./... -race
make e2e                # End-to-end route tests
make frontend-build     # Build React SPA into web/static/
make lint               # go vet ./...

# From repo root (before merge)
make validate-diff      # Minimal gate for current changes
make gates              # Full CI: vet, race tests, boundary checks
```

## Architecture highlights

### Storage layer (dialect pattern)

`internal/storage/sql/` uses a **dialect interface** to abstract SQL differences
between SQLite and MySQL:

- `dialect.go` — Interface + implementations (placeholder, RETURNING vs LastInsertId,
  ON CONFLICT vs ON DUPLICATE KEY UPDATE, PRAGMA)
- `sql.go` — Backend-agnostic business logic using dialect methods
- `schema_sqlite.sql` / `schema_mysql.sql` — Embedded DDL

Adding a new backend (PostgreSQL, etc.): implement dialect, create schema file,
add driver import, update factory.

### Frontend (React + Vite)

- Route-level lazy loading: `React.lazy()` + `Suspense` in `App.jsx`
- Manual chunk splitting: vendor (React/Router), charts (Chart.js), icons (lucide-react)
- Initial load: ~217 KB (vs 663 KB single bundle)

### Wire-format compatibility

guardus mirrors upstream gatus JSON format exactly (no envelope, plain-text errors)
so the SPA works unmodified. This is a documented deviation from plumego's
`contract.WriteResponse` pattern.

## Configuration

All runtime config via `GUARDUS_*` environment variables. See `env.example` for
the complete surface. Key settings:

```bash
# Storage
GUARDUS_STORAGE_TYPE=sqlite|mysql|memory
GUARDUS_STORAGE_PATH=./guardus.db  # or MySQL DSN

# Web
GUARDUS_WEB_PORT=8080

# Bootstrap (optional)
GUARDUS_BOOTSTRAP_FILE=./endpoints.json
```

## Where to work

| Task | Start here |
|------|------------|
| Add probe type | `internal/client/` + `internal/domain/endpoint/` |
| Add alert provider | `internal/alerting/provider/` + `internal/alerting/config.go` |
| Add storage backend | `internal/storage/sql/dialect.go` (see AGENTS.md §4) |
| Modify SPA | `web/app/src/` (React + Vite) |
| Change HTTP handler | `internal/handler/` (use `WriteRawJSON`, not `contract.WriteResponse`) |
| Add env var | `internal/config/env.go` + `env.example` |

## Hard rules

- **No YAML** — env vars + JSON only
- **No contract envelope** — use `handler.WriteRawJSON` / `WriteErrorString`
- **No reflection/init()** — explicit typed fields + switches
- **No globals** — everything injected via `app.New`
- **Module-local deps** — guardus deps stay in this `go.mod`, never leak to root

## Validation

Before merging:

```bash
go run ./internal/checks/dependency-rules
go run ./internal/checks/cross-extension-deps
go run ./internal/checks/module-manifests
go run ./internal/checks/reference-layout
make gates
```

Verify no deps leak: `grep modernc ../../go.mod` (must return nothing).

---

@AGENTS.md
