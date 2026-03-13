# Plumego CLI - Quick Start Guide

## Installation

```bash
# Install from source
go install github.com/spcent/plumego/cmd/plumego@latest

# Or build locally
cd cmd/plumego
go build -o plumego
```

## Basic Usage

### Get Help

```bash
# General help
plumego --help

# Command-specific help
plumego new --help
plumego generate --help
```

### Create New Project

```bash
# Minimal project (default)
plumego new myapp

# API server
plumego new myapi --template api

# Full-stack application
plumego new webapp --template fullstack

# Custom module path
plumego new myapp --module github.com/myorg/myapp

# Preview without creating
plumego new myapp --dry-run --format json
```

### Output Formats

All commands support multiple output formats:

```bash
# JSON (default, best for code agents)
plumego new myapp --format json

# YAML
plumego new myapp --format yaml

# Plain text
plumego new myapp --format text
```

### Example JSON Output

```bash
$ plumego new myapp --format json
{
  "status": "success",
  "message": "Project created successfully",
  "data": {
    "project": "myapp",
    "path": "./myapp",
    "module": "myapp",
    "template": "minimal",
    "files_created": [
      "main.go",
      "go.mod",
      "env.example",
      ".gitignore",
      "README.md"
    ],
    "next_steps": [
      "cd myapp",
      "go mod tidy",
      "plumego dev"
    ]
  }
}
```

## Code Agent Integration

### Parsing Output

```bash
# Create project and extract path
PROJECT_PATH=$(plumego new myapp --format json | jq -r '.data.path')
cd "$PROJECT_PATH"

# Check exit code
if plumego check --format json; then
  echo "Project is healthy"
else
  echo "Health check failed with exit code $?"
fi
```

### Automation Example

```bash
#!/bin/bash
set -euo pipefail

# Create project
OUTPUT=$(plumego new myapp --template api --format json)
PROJECT=$(echo "$OUTPUT" | jq -r '.data.project')
PATH=$(echo "$OUTPUT" | jq -r '.data.path')

# Navigate to project
cd "$PATH"

# Validate
plumego check --format json > health.json

# Extract validation results
STATUS=$(jq -r '.status' health.json)
if [ "$STATUS" != "healthy" ]; then
  echo "Project validation failed"
  jq -r '.checks.security.issues[] | .message' health.json
  exit 1
fi

echo "Project $PROJECT created and validated successfully"
```

## Global Flags

| Flag | Description | Example |
|------|-------------|---------|
| `--format`, `-f` | Output format (json, yaml, text) | `plumego new myapp -f yaml` |
| `--quiet`, `-q` | Suppress non-essential output | `plumego new myapp -q` |
| `--verbose`, `-v` | Detailed logging | `plumego new myapp -v` |
| `--no-color` | Disable color output (text mode) | `plumego new myapp --no-color` |
| `--env-file` | Environment file path (used by `plumego config`) | `plumego --env-file .env.dev config show` |

## Exit Codes

| Code | Meaning | Action |
|------|---------|--------|
| 0 | Success | Continue |
| 1 | General error | Check error message |
| 2 | Configuration/validation error | Fix configuration |
| 3 | Resource conflict | Use --force or resolve |

## Environment File

Create `.env` (or use `--env-file`) to supply app settings to `plumego config`:

```bash
APP_ADDR=:8080
APP_DEBUG=false
APP_SHUTDOWN_TIMEOUT_MS=5000
WS_SECRET=your-websocket-secret-here-32-bytes-minimum
JWT_EXPIRY=15m
```

## Environment Variables

`plumego config` reads `APP_`, `WS_`, `JWT_`, `DB_`, and `REDIS_` from the process environment (and from the `.env` file when `--resolve` is set).

CLI global flags are not currently configurable via `PLUMEGO_*`.

## Next Steps

1. See full CLI design: [CLI_DESIGN.md](./CLI_DESIGN.md)
2. Read project documentation: [../README.md](../README.md)
3. Explore examples: [../examples/](../examples/)

## Tips for Code Agents

1. **Always use JSON format** for parsing output
2. **Check exit codes** to determine success/failure
3. **Use `--dry-run`** to preview operations
4. **Enable `--verbose`** for debugging
5. **Parse structured output** rather than text
6. **Chain commands** using pipes and jq
7. **Set `--quiet`** in CI/CD to reduce noise

## Example: CI/CD Integration

```yaml
# GitHub Actions example
name: Validate Project
on: [push]

jobs:
  validate:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - name: Install plumego CLI
        run: go install github.com/spcent/plumego/cmd/plumego@latest

      - name: Check project health
        run: |
          plumego check --format json > health.json
          cat health.json

      - name: Validate exit code
        run: |
          STATUS=$(jq -r '.status' health.json)
          if [ "$STATUS" != "healthy" ]; then
            echo "Health check failed"
            exit 1
          fi
```

## Support

- Documentation: https://github.com/spcent/plumego/tree/main/docs
- Issues: https://github.com/spcent/plumego/issues
- Examples: https://github.com/spcent/plumego/tree/main/examples
