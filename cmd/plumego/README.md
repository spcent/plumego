# Plumego CLI

A code agent-friendly command-line tool for plumego projects. Designed for automation, CI/CD integration, and AI-assisted development workflows.

## Current Status

- Included in the Plumego planned v1 hardening scope
- Supported as a command-line tool, not as a Go library import surface
- Command behavior and generated output are part of the v1 hardening scope and must stay aligned with the repository's canonical docs

## Features

- **Machine-First Design**: JSON/YAML/Text output formats (default: JSON)
- **Non-Interactive**: All operations via flags, no prompts
- **Predictable**: Clear exit codes (0=success, 1=error, 2=warning)
- **Composable**: Works seamlessly with jq, grep, pipes
- **Automation-Ready**: Perfect for CI/CD and code agents

## Installation

### From Source

```bash
# Clone the repository
git clone https://github.com/spcent/plumego.git
cd plumego/cmd/plumego

# Build and install
go build -o plumego .
sudo mv plumego /usr/local/bin/

# Or just build locally
go build -o ../../bin/plumego .
```

### Verify Installation

```bash
plumego --help
```

## Quick Start

### Create a New Project

```bash
# Create a minimal project
plumego new myapp

# Create the canonical reference-style layout
plumego new myapp --template canonical

# Create an API server
plumego new myapi --template api

# Create a scenario-profile scaffold
plumego new rest-api --template rest-api
plumego new tenant-api --template tenant-api
plumego new edge-api --template gateway
plumego new realtime-api --template realtime
plumego new ai-api --template ai-service
plumego new ops-api --template ops-service

# Create with custom module path
plumego new myapp --template fullstack --module github.com/myorg/myapp
```

Supported templates are: `canonical`, `minimal`, `api`, `fullstack`,
`microservice`, `rest-api`, `tenant-api`, `gateway`, `realtime`, `ai-service`,
and `ops-service`.

The `canonical` template is aligned with `reference/standard-service`: stable
root imports only, explicit config loading, explicit route registration in
`internal/app/routes.go`, local handler DTOs, and no `x/*` capability wiring by
default.

The `api` template starts from the same canonical bootstrap and adds a minimal
`x/rest` users resource profile under `internal/resource`. It keeps resource
route wiring explicit in `internal/app/routes.go`; `x/rest` is not part of the
default `canonical` template.

Scenario profiles keep the same canonical bootstrap and add explicit optional
capability wiring. They include runnable scenario routes plus
`internal/scenario/profile.go` capability markers for the selected family:

| Template | Capability profile |
| --- | --- |
| `rest-api` | `x/rest` users resource under `/api/users` |
| `tenant-api` | `x/tenant/resolve`, `x/tenant/policy`, `x/tenant/quota`, `x/tenant/ratelimit` on `/api/models` |
| `gateway` | `x/gateway` loopback proxy under `/edge` |
| `realtime` | `x/websocket` hub metrics under `/realtime/metrics`, plus `x/messaging` marker |
| `ai-service` | offline `x/ai/provider`, `x/ai/session`, and `x/ai/tool` demo under `/ai/demo`, plus `x/ai/streaming` marker |
| `ops-service` | protected `x/observability` metrics under `/ops/metrics` and protected admin boundary summary under `/ops/admin` using `x/ops` DTOs |

These profiles do not install secrets, live provider credentials, hidden
globals, or default `x/devtools` routes.

### Generate Code

```bash
# Generate a component
plumego generate component Auth --with-tests

# Generate middleware
plumego generate middleware RateLimit

# Generate REST handlers
plumego generate handler User --methods GET,POST,PUT,DELETE --with-tests
```

### Development Server

```bash
# Start dev server with hot reload
plumego dev

# Custom port
plumego dev --addr :3000
```

The development dashboard APIs return structured Plumego error responses for
local tooling failures. Public error codes are uppercase stable identifiers, and
default messages avoid exposing local filesystem, parser, build, or `go list`
diagnostics.

## Commands

The v1 CLI surface currently targets these 9 commands:

1. **new** - Create projects from templates
2. **generate** - Generate middleware, handlers, models
3. **dev** - Development server with hot reload
4. **check** - Health and security validation
5. **config** - Configuration management
6. **routes** - Route analysis and inspection
7. **build** - Build with optimizations
8. **test** - Enhanced test runner
9. **inspect** - Runtime inspection

See module notes: [MODULE.md](./MODULE.md)

## CI/CD Integration

```yaml
# GitHub Actions example
- name: Install Plumego CLI
  run: |
    cd cmd/plumego
    go build -o $GITHUB_WORKSPACE/bin/plumego .

- name: Health Check
  run: plumego check --format json

- name: Run Tests
  run: plumego test --race --cover --format json
```

## Exit Codes

- `0` = Success
- `1` = General error
- `2` = Configuration error
- `3` = Resource conflict

## License

Same as plumego core - see [LICENSE](../../LICENSE).
