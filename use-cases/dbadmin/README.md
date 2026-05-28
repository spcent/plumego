# dbadmin

Web-based MySQL and SQLite management tool built on [plumego](../../README.md),
comparable to PHP Adminer.

## Features

- **Connection management** — save and test MySQL / SQLite connections with
  optional AES-GCM encrypted password storage
- **Database / table browsing** — left-sidebar tree of connections → databases → tables
- **Table structure** — columns, indexes, foreign keys, DDL source
- **Data viewer** — paginated table data with column filter, sort
- **Row CRUD** — insert, edit, delete rows with parameterized queries
- **SQL Console** — Monaco-powered editor, Ctrl+Enter to run, query history
- **DDL operations** — create table, alter table (add/drop columns, rename), drop table
- **Export** — CSV or SQL INSERT dump per table
- **Import** — upload a `.sql` file to execute against any connection
- **Auth** — session-cookie login, configurable credentials, HttpOnly cookies

## Quick start

```bash
# 1. Build the frontend
cd web && npm install && npm run build && cd ..

# 2. Copy and edit config
cp env.example .env

# 3. Run
go run ./...
# → http://localhost:8080
```

## Configuration

| Variable | Default | Description |
|---|---|---|
| `APP_ADDR` | `:8080` | Listen address |
| `DBADMIN_DATA_DIR` | `./data` | KV store directory |
| `DBADMIN_USER` | `admin` | Login username |
| `DBADMIN_PASSWORD` | `admin` | Login password |
| `DBADMIN_ENCRYPTION_KEY` | (none) | 64-char hex AES-GCM key for passwords |

## Architecture

```
use-cases/dbadmin/
├── main.go
├── internal/
│   ├── config/           # env / flag loading
│   ├── app/              # core.App + middleware + routes
│   ├── handler/          # HTTP handlers (auth, connections, inspect, rows, query, ddl, export, import)
│   ├── dbmanager/        # connection pool + Inspector interface + mysql / sqlite implementations
│   └── domain/
│       ├── authn/        # session-cookie authenticator
│       ├── connection/   # saved connection CRUD (KV-backed, AES-GCM passwords)
│       ├── history/      # SQL query history (KV-backed)
│       └── session/      # session store (KV-backed, TTL)
└── web/                  # React 19 + Vite SPA (served via x/frontend)
```

## Boundary rules

This is a use-case app with its own `go.mod`.  It may import:
- All stable plumego roots (`core`, `contract`, `middleware/*`, `security`, `store`, `health`, `log`, `metrics`)
- Extensions: `x/frontend` only (for SPA serving)
- External drivers: `github.com/go-sql-driver/mysql`, `modernc.org/sqlite`
