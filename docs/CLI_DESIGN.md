# Plumego CLI Design — Code Agent Friendly

## Overview

A command-line interface for plumego that prioritizes machine readability, automation, and code agent interaction. The CLI follows Unix philosophy: composable, predictable, and scriptable.

## Design Principles

1. **Machine-First Output**: Default to structured formats (JSON/YAML)
2. **Consistent Exit Codes**: Clear success/failure indicators
3. **Non-Interactive**: All operations can run without prompts
4. **Idempotent**: Safe to run multiple times
5. **Composable**: Commands do one thing well
6. **Discoverable**: Self-documenting with `--help`
7. **Configuration Transparency**: Flags are explicit; `.env` is used by `plumego config` for app settings (no global config file support yet)

---

## Installation

```bash
go install github.com/spcent/plumego/cmd/plumego@latest
```

---

## Command Structure

```
plumego [global-flags] <command> [command-flags] [args]
```

### Global Flags

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--format` | `-f` | string | `json` | Output format: json, yaml, text |
| `--quiet` | `-q` | bool | false | Suppress non-essential output |
| `--verbose` | `-v` | bool | false | Detailed logging |
| `--no-color` | | bool | false | Disable color output (text mode) |
| `--env-file` | | string | `.env` | Environment file path (used by `plumego config`) |

**Response Envelope:**
Most commands return a standard envelope:
`{"status":"success","message":"...","data":{...}}` or `{"status":"error","message":"...","exit_code":1,"data":{...}}`.

---

## Commands

### 1. `plumego new` — Create New Project

Creates a new plumego project from templates.

```bash
plumego new [flags] <project-name>
```

**Flags:**
- `--template <name>` - Template: minimal, api, fullstack, microservice (default: minimal)
- `--module <path>` - Go module path (default: inferred from project-name)
- `--dir <path>` - Output directory (default: ./project-name)
- `--force` - Overwrite existing directory
- `--no-git` - Skip git initialization
- `--dry-run` - Show what would be created without creating

**Output (JSON):**
```json
{
  "project": "myapp",
  "path": "/home/user/myapp",
  "module": "github.com/user/myapp",
  "template": "api",
  "files_created": [
    "main.go",
    "go.mod",
    "env.example",
    ".gitignore"
  ],
  "next_steps": [
    "cd myapp",
    "go mod tidy",
    "plumego dev"
  ]
}
```

**Exit Codes:**
- `0` - Success
- `1` - General error
- `2` - Directory exists (without --force)
- `3` - Invalid template

**Examples:**
```bash
# Minimal project
plumego new myapp

# API server with custom module path
plumego new myapi --template api --module github.com/acme/myapi

# Full-stack app with frontend
plumego new webapp --template fullstack

# Dry run to preview
plumego new myapp --dry-run --format yaml
```

---

### 2. `plumego generate` — Code Generation

Generates components, middleware, handlers, and other boilerplate.

```bash
plumego generate <type> [flags] <name>
```

**Types:**
- `component` - Component with full lifecycle
- `middleware` - HTTP middleware
- `handler` - HTTP handler
- `model` - Data model with validation

**Flags:**
- `--output <path>` - Output file path (default: auto-detect)
- `--package <name>` - Package name (default: inferred)
- `--methods <list>` - HTTP methods for handlers (GET,POST,PUT,DELETE)
- `--with-tests` - Generate test file
- `--with-validation` - Generate validation
- `--force` - Overwrite existing files

**Output (JSON):**
```json
{
  "type": "component",
  "name": "AuthComponent",
  "files": {
    "created": ["components/auth/auth.go"],
    "modified": ["main.go"]
  },
  "imports": [
    "github.com/spcent/plumego/core",
    "github.com/spcent/plumego/security/jwt"
  ]
}
```

**Exit Codes:**
- `0` - Success
- `1` - Generation failed
- `2` - Invalid type
- `3` - File exists (without --force)

**Examples:**
```bash
# Generate component
plumego generate component Auth --with-tests

# Generate middleware
plumego generate middleware RateLimit --output middleware/ratelimit.go

# Generate REST handler with tests
plumego generate handler User --methods GET,POST,PUT,DELETE --with-tests

# Generate in JSON for parsing
plumego generate component Cache --format json > output.json
```

---

### 3. `plumego dev` — Development Server

Starts development server with hot reload and enhanced logging.

```bash
plumego dev [flags]
```

**Flags:**
- `--dir <path>` - Project directory (default: .)
- `--addr <address>` - Listen address (default: :8080)
- `--dashboard-addr <address>` - Dashboard listen address (default: :9999)
- `--watch <patterns>` - Watch patterns (default: **/*.go)
- `--exclude <patterns>` - Exclude patterns
- `--debounce <duration>` - Debounce duration for file changes (default: 500ms)
- `--no-reload` - Disable hot reload (no file watcher)
- `--build-cmd <cmd>` - Custom build command (must output `.dev-server` unless `--run-cmd` is set)
- `--run-cmd <cmd>` - Custom run command

**Output (JSON streaming):**
```json
{"event": "starting", "message": "Plumego Dev Server starting", "data": {"project": "/abs/path", "app_addr": ":8080", "dashboard_addr": ":9999"}}
{"event": "dashboard_started", "message": "Dashboard started", "data": {"url": "http://localhost:9999"}}
{"event": "ready", "message": "Application ready", "data": {"url": "http://localhost:8080"}}
{"event": "file_changed", "message": "File changed", "data": {"path": "handler.go"}}
{"event": "reloading", "message": "Reloading application", "data": {"reason": "file_changed", "path": "handler.go"}}
{"event": "reload_complete", "message": "Reload complete"}
{"event": "stopped", "message": "Shutting down", "data": {"code": 0}}
```

Additional events may be emitted for build and app lifecycle/logs:
`dashboard_started`, `watching`, `reload_complete`, `reload_disabled`,
`build.start`, `build.success`, `build.fail`, `app.start`, `app.stop`, `app.log`, `app.error`.

Event field conventions (for CI parsing):
- `build.*`: `data.success` (bool), `data.duration_ms` (int, when available), `data.error` (string), `data.output` (string)
- `app.*`: `data.state` (string), `data.pid` (int, when available), `data.error` (string)
- `app.log` / `app.error`: `level` (info|warn|error|debug), `message` (string), `data.source` (stdout|stderr)

**Exit Codes:**
- `0` - Clean exit (including Ctrl+C)
- `1` - Build error

**Examples:**
```bash
# Standard dev server
plumego dev

# Custom port with specific watch patterns
plumego dev --addr :3000 --watch "internal/**/*.go,pkg/**/*.go"

# Without hot reload
plumego dev --no-reload

# Custom build and run commands
plumego dev --build-cmd "go build -o .dev-server ./cmd/api" --run-cmd "./.dev-server"
```

---

### 4. `plumego routes` — Route Inspection

Lists all registered routes with methods, paths, and middleware.

```bash
plumego routes [flags] [pattern]
```

**Flags:**
- `--group <name>` - Filter by route group
- `--method <method>` - Filter by HTTP method
- `--middleware` - Show middleware chain
- `--sort <field>` - Sort by: path, method, group (default: path)

**Output (JSON):**
```json
{
  "routes": [
    {
      "method": "GET",
      "path": "/api/v1/users/:id",
      "handler": "handlers.GetUser",
      "group": "api.v1",
      "middleware": [
        "RequestID",
        "Logging",
        "JWTAuth",
        "RateLimit"
      ]
    }
  ],
  "total": 1
}
```

**Exit Codes:**
- `0` - Success
- `1` - Failed to load routes

**Examples:**
```bash
# All routes in JSON
plumego routes --format json

# Filter by pattern
plumego routes "/api/*"

# Show middleware chains
plumego routes --middleware --format yaml

# Filter by method
plumego routes --method POST
```

---

### 5. `plumego check` — Project Health Check

Validates project structure, configuration, and dependencies.

```bash
plumego check [flags]
```

**Flags:**
- `--config-only` - Only check configuration
- `--deps-only` - Only check dependencies
- `--security` - Run security checks

**Output (JSON):**
```json
{
  "status": "healthy",
  "checks": {
    "config": {
      "status": "passed",
      "issues": []
    },
    "dependencies": {
      "status": "passed",
      "outdated": []
    },
    "security": {
      "status": "warning",
      "issues": [
        {
          "severity": "medium",
          "message": "WS_SECRET not set in environment",
          "fix": "Set WS_SECRET to 32+ byte secure random string"
        }
      ]
    }
  }
}
```

**Exit Codes:**
- `0` - All checks passed
- `1` - Errors found
- `2` - Warnings found (non-blocking)

**Examples:**
```bash
# Full health check
plumego check

# Config validation only
plumego check --config-only

# Security audit
plumego check --security --format json
```

---

### 6. `plumego migrate` — Database Migrations

Manages database migrations (when using plumego/store/db).

```bash
plumego migrate <command> [flags]
```

**Commands:**
- `up` - Apply pending migrations
- `down` - Rollback last migration
- `status` - Show migration status
- `create` - Create new migration file

**Flags:**
- `--steps <n>` - Number of migrations (default: all)
- `--db-url <url>` - Database connection string (DSN)
- `--driver <name>` - Database driver name (e.g., sqlite3, postgres, mysql)
- `--dir <path>` - Migrations directory (default: ./migrations)

**Output (JSON):**
```json
{
  "command": "up",
  "applied": [
    {
      "version": "20260201120000",
      "name": "create_users_table",
      "duration_ms": 45
    }
  ],
  "pending": [],
  "current_version": "20260201120000"
}
```

**Exit Codes:**
- `0` - Success
- `1` - Migration failed
- `2` - No migrations to apply

**Examples:**
```bash
# Apply all pending migrations
plumego migrate up --driver postgres --db-url "postgres://localhost/mydb"

# Create new migration
plumego migrate create add_users_email_index

# Check status
plumego migrate status --driver postgres --db-url "postgres://localhost/mydb" --format json
```

---

### 7. `plumego config` — Configuration Management

View, validate, and generate environment configuration.

Configuration is loaded from `.env` (or `--env-file`) and the current process environment. A global YAML config file is not used by the CLI yet.

```bash
plumego config <command> [flags]
```

**Commands:**
- `show` - Display current configuration
- `validate` - Validate configuration
- `init` - Generate `env.example`
- `env` - Show environment variables

**Flags:**
- `--resolve` - Resolve environment variables
- `--redact` - Redact sensitive values

**Output (JSON):**
```json
{
  "config": {
    "app": {
      "addr": ":8080",
      "debug": false,
      "shutdown_timeout_ms": 5000
    },
    "security": {
      "ws_secret": "***REDACTED***",
      "jwt_expiry": "15m"
    }
  },
  "source": {
    "app.addr": "env:APP_ADDR",
    "app.debug": "default",
    "security.ws_secret": "env:WS_SECRET"
  }
}
```

**Exit Codes:**
- `0` - Valid configuration
- `1` - Invalid configuration
- `2` - Valid with warnings

**Examples:**
```bash
# Show resolved config
plumego config show --resolve --redact

# Validate configuration
plumego config validate

# Generate default files
plumego config init

# List environment variables
plumego config env --format yaml
```

---

### 8. `plumego test` — Testing Utilities

Enhanced test running with plumego-specific helpers.

```bash
plumego test [flags] [packages...]
```

**Flags:**
- `--dir <path>` - Project directory (default: .)
- `--race` - Enable race detector
- `--cover` - Generate coverage report
- `--bench` - Run benchmarks
- `--timeout <duration>` - Test timeout (default: 20s)
- `--tags <tags>` - Build tags
- `--run <pattern>` - Run only tests matching pattern
- `--short` - Run short tests only

**Output (JSON):**
```json
{
  "status": "passed",
  "tests": 45,
  "passed": 43,
  "failed": 2,
  "skipped": 0,
  "duration_ms": 1234,
  "race_detector": false,
  "coverage_enabled": true,
  "benchmark": false,
  "coverage_percent": 78.5,
  "failures": [
    {
      "package": "github.com/user/myapp/handlers",
      "test": "TestUserCreate"
    }
  ]
}
```

**Exit Codes:**
- `0` - All tests passed
- `1` - Tests failed
- `2` - Build failed

**Examples:**
```bash
# Run all tests with race detector
plumego test --race ./...

# Generate coverage report
plumego test --cover --format json > coverage.json

# Run benchmarks
plumego test --bench --format text
```

---

### 9. `plumego build` — Build Utilities

Build application with plumego-specific optimizations.

```bash
plumego build [flags]
```

**Flags:**
- `--output <path>` - Output binary path (default: ./bin/app)
- `--ldflags <flags>` - Go linker flags
- `--tags <tags>` - Build tags
- `--embed-frontend` - Embed frontend assets
- `--compress` - Compress binary with UPX

**Output (JSON):**
```json
{
  "binary": "./bin/app",
  "size_bytes": 12582912,
  "build_time_ms": 3456,
  "go_version": "1.24.0",
  "git_commit": "a1b2c3d",
  "embedded": ["frontend/dist"]
}
```

**Exit Codes:**
- `0` - Build successful
- `1` - Build failed

**Examples:**
```bash
# Standard build
plumego build

# Production build with embedded frontend
plumego build --output ./bin/prod --embed-frontend --tags prod

# Get build info as JSON
plumego build --format json
```

---

### 10. `plumego inspect` — Runtime Inspection

Inspect running plumego application via HTTP endpoints.

```bash
plumego inspect <command> [flags]
```

`routes`, `config`, and `info` call `/_routes`, `/_config`, and `/_info` (core provides compatible aliases under `/_debug/*`).

**Commands:**
- `health` - Check health endpoints
- `metrics` - Fetch metrics
- `routes` - List active routes
- `config` - View runtime config
- `info` - General application info

**Flags:**
- `--url <url>` - Application URL (default: http://localhost:8080)
- `--auth <token>` - Authentication token

**Output (JSON):**
```json
{
  "status": "healthy",
  "checks": {
    "database": "healthy",
    "cache": "healthy"
  },
  "endpoint": "/health",
  "status_code": 200
}
```

**Exit Codes:**
- `0` - Application healthy
- `1` - Application unhealthy or unreachable

**Examples:**
```bash
# Check health
plumego inspect health --url http://localhost:8080

# Get metrics in JSON
plumego inspect metrics --format json

# View runtime routes
plumego inspect routes --auth "Bearer token"
```

---

## Configuration Sources

### .env (optional)

`plumego config` reads key/value pairs from `.env` (or `--env-file`) and merges them with the current process environment.

Example `.env`:

```bash
APP_ADDR=:8080
APP_DEBUG=false
APP_SHUTDOWN_TIMEOUT_MS=5000
WS_SECRET=your-websocket-secret-here-32-bytes-minimum
JWT_EXPIRY=15m
```

### Environment Variables

At runtime, the `config` command reads the following prefixes from the process environment (and, when `--resolve` is set, from the `.env` file): `APP_`, `WS_`, `JWT_`, `DB_`, `REDIS_`.

Note: CLI global flags are not currently configurable via `PLUMEGO_*` environment variables.

---

## Exit Code Reference

| Code | Meaning | Action |
|------|---------|--------|
| 0 | Success | Continue |
| 1 | General error | Check error message |
| 2 | Configuration error | Fix configuration |
| 3 | File/resource conflict | Use --force or resolve manually |
| 130 | Interrupted (SIGINT) | Normal Ctrl+C |
| 137 | Killed (SIGKILL) | System killed process |

---

## Code Agent Integration Examples

### Example 1: Project Scaffolding

```bash
# Agent creates new project
OUTPUT=$(plumego new myapp --template api --format json)

# Parse result
PROJECT_PATH=$(echo "$OUTPUT" | jq -r '.path')
cd "$PROJECT_PATH"

# Initialize dependencies
plumego config init
plumego migrate create init_schema
```

### Example 2: Health Check in CI

```bash
#!/bin/bash
set -euo pipefail

# Start app in background
plumego dev --addr :8080 --no-reload &
APP_PID=$!

# Wait for ready
sleep 2

# Run health check
if plumego inspect health --url http://localhost:8080 --format json > health.json; then
  echo "Health check passed"
  kill $APP_PID
  exit 0
else
  echo "Health check failed"
  kill $APP_PID
  exit 1
fi
```

### Example 3: Route Documentation Generation

```bash
# Extract routes as JSON
plumego routes --middleware --format json > routes.json

# Generate OpenAPI spec from routes
plumego generate openapi --input routes.json --output openapi.yaml
```

### Example 4: Configuration Validation in CI

```bash
#!/bin/bash

# Validate configuration
if ! plumego config validate --format json > validation.json; then
  echo "Configuration validation failed:"
  jq -r '.errors[] | "  - \(.field): \(.message)"' validation.json
  exit 1
fi

# Check for missing secrets
MISSING=$(jq -r '.warnings[] | select(.type == "missing_secret") | .field' validation.json)
if [ -n "$MISSING" ]; then
  echo "Missing secrets: $MISSING"
  exit 2
fi
```

---

## Implementation Roadmap

### Phase 1: Core Commands (Week 1-2)
- `plumego new` - Project scaffolding
- `plumego generate` - Code generation
- `plumego check` - Health checks
- `plumego config` - Configuration management

### Phase 2: Development Tools (Week 3-4)
- `plumego dev` - Development server with hot reload
- `plumego routes` - Route inspection
- `plumego test` - Test runner

### Phase 3: Advanced Features (Week 5-6)
- `plumego build` - Build utilities
- `plumego migrate` - Database migrations
- `plumego inspect` - Runtime inspection

### Phase 4: Extensions (Week 7+)
- Plugin system for custom commands
- Template marketplace
- Interactive TUI mode (optional)

---

## Technical Implementation Notes

### Technologies
- CLI Framework: `cobra` + `pflag`
- Output Formatting: `encoding/json`, `gopkg.in/yaml.v3`
- File Watching: `fsnotify`
- Terminal UI: `charm.sh/lipgloss` (for text mode)
- Code Generation: `text/template` + AST manipulation

### Project Structure
```
cmd/plumego/
├── main.go
├── commands/
│   ├── new.go
│   ├── generate.go
│   ├── dev.go
│   ├── routes.go
│   ├── check.go
│   ├── config.go
│   ├── migrate.go
│   ├── test.go
│   ├── build.go
│   └── inspect.go
├── internal/
│   ├── scaffold/      # Project scaffolding
│   ├── codegen/       # Code generation
│   ├── watcher/       # File watching
│   ├── inspector/     # Runtime inspection
│   └── templates/     # Template management
└── templates/
    ├── minimal/
    ├── api/
    ├── fullstack/
    └── microservice/
```

---

## Key Features for Code Agents

1. **Structured Output**: All commands support `--format json` for parsing
2. **Non-Interactive**: No prompts, all input via flags/env
3. **Predictable**: Same command always produces same result
4. **Composable**: Commands can be chained with pipes
5. **Exit Codes**: Clear success/failure indicators
6. **Verbose Mode**: Detailed logging for debugging
7. **Dry Run**: Preview operations without executing
8. **Force Mode**: Override safety checks when needed
9. **Validation**: Pre-flight checks before operations
10. **Documentation**: Every command has `--help` with examples

---

## Comparison with Existing Tools

| Feature | plumego | go | cobra | make |
|---------|---------|----|---------|----|
| Project scaffolding | ✅ | ❌ | ❌ | ❌ |
| Code generation | ✅ | ✅ (limited) | ❌ | ❌ |
| Hot reload | ✅ | ❌ | ❌ | ✅ (manual) |
| Route inspection | ✅ | ❌ | ❌ | ❌ |
| Config validation | ✅ | ❌ | ❌ | ❌ |
| JSON output | ✅ | ❌ | ❌ | ❌ |
| Agent-friendly | ✅ | ⚠️ | ⚠️ | ✅ |

---

## Future Enhancements

1. **AI Integration**: `plumego ask "how to add rate limiting?"`
2. **Plugin System**: Custom commands via Go plugins
3. **Remote Inspect**: Inspect deployed applications
4. **Performance Profiling**: Built-in pprof integration
5. **Deployment Helpers**: Docker, Kubernetes manifests
6. **Security Scanning**: Automated vulnerability checks
7. **Dependency Updates**: Automated dep management
8. **Telemetry**: Optional usage analytics
9. **Interactive Mode**: TUI for non-agent users
10. **Cloud Integration**: Deploy to cloud providers

---

## Summary

This CLI design prioritizes **machine readability**, **automation**, and **composability**. Every command:

- Outputs structured data (JSON/YAML)
- Has clear exit codes
- Supports non-interactive operation
- Provides detailed help
- Can be chained with other tools

This makes it ideal for:
- Code agents (Claude Code, GitHub Copilot, etc.)
- CI/CD pipelines
- Automation scripts
- Testing frameworks
- DevOps tools

The CLI extends plumego's philosophy of being **explicit**, **composable**, and **standard library-first** to command-line operations.
