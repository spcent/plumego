# Plumego CLI - Summary

## Overview

A code agent-friendly command-line interface for the plumego HTTP toolkit. Designed for automation, CI/CD integration, and programmatic interaction.

## Key Features

### 1. **Machine-First Design**
- Default JSON output for all commands
- Structured data that's easily parseable
- Consistent schema across all operations

### 2. **Non-Interactive Operation**
- No prompts or user input required
- All configuration via flags, env vars, or config files
- Perfect for CI/CD and automation

### 3. **Predictable Exit Codes**
```
0 = Success
1 = General error
2 = Configuration/validation error
3 = Resource conflict
130 = User interrupt (Ctrl+C)
```

### 4. **Multiple Output Formats**
- JSON (default, for machines)
- YAML (human-readable structured)
- Text (simple output)

### 5. **Composable Commands**
Each command does one thing well and can be chained:
```bash
plumego new myapp --format json | jq -r '.data.path' | xargs cd
```

## Quick Comparison

### Traditional CLI
```bash
$ some-tool create myapp
? What template? (Use arrow keys)
  ▸ minimal
    api
    fullstack
? Initialize git? (y/n)
✓ Project created!
```

### Plumego CLI (Agent-Friendly)
```bash
$ plumego new myapp --template api --format json
{
  "status": "success",
  "message": "Project created successfully",
  "data": {
    "project": "myapp",
    "path": "./myapp",
    "module": "myapp",
    "template": "api",
    "files_created": ["main.go", "go.mod", "..."],
    "next_steps": ["cd myapp", "go mod tidy", "plumego dev"]
  }
}
```

## Command Overview

| Command | Purpose | Output | Status |
|---------|---------|--------|--------|
| `new` | Create project from template | Project metadata | ✅ Implemented |
| `generate` | Generate code (components, handlers) | Generated files | ✅ Implemented |
| `dev` | Development server with hot reload | Real-time events | ✅ Implemented |
| `routes` | Inspect registered routes | Route list | ✅ Implemented |
| `check` | Health and security checks | Validation report | ✅ Implemented |
| `config` | Configuration management | Config tree | ✅ Implemented |
| `migrate` | Database migrations | Migration status | 🚧 Stub |
| `test` | Enhanced test runner | Test results | ✅ Implemented |
| `build` | Build with optimizations | Build metadata | ✅ Implemented |
| `inspect` | Inspect running app | Runtime status | ✅ Implemented |

## Architecture

```
cmd/plumego/
├── main.go                    # Entry point
├── commands/
│   ├── root.go               # Command dispatcher
│   ├── new.go                # Project scaffolding (implemented)
│   ├── generate.go           # Code generation
│   ├── dev.go                # Development server
│   ├── routes.go             # Route inspection
│   ├── check.go              # Health checks
│   ├── config.go             # Config management
│   ├── test.go               # Test runner
│   ├── build.go              # Build utilities
│   ├── inspect.go            # Runtime inspection
│   └── stubs.go              # Stub registry (legacy placeholder)
└── internal/
    ├── output/
    │   └── formatter.go      # Output formatting (JSON/YAML/Text)
    ├── scaffold/
    │   └── scaffold.go       # Project scaffolding
    ├── codegen/              # Code generation
    ├── routeanalyzer/        # Route inspection analysis
    └── watcher/              # File watching
```

## Design Principles

### 1. Unix Philosophy
- Do one thing well
- Work together with other tools
- Handle text streams (JSON)

### 2. Configuration Layering
```
Flags > Environment Variables > Config File > Defaults
```

### 3. Idempotency
Safe to run commands multiple times:
```bash
plumego new myapp        # Creates project
plumego new myapp        # Error: directory exists
plumego new myapp --force # Overwrites
```

### 4. Dry Run Support
Preview operations without executing:
```bash
plumego new myapp --dry-run --format json
{
  "dry_run": true,
  "files_created": ["main.go", "..."],
  ...
}
```

### 5. Verbose Logging
```bash
plumego new myapp --verbose
[VERBOSE] Creating project: myapp
[VERBOSE] Template: minimal
[VERBOSE] Module: myapp
[VERBOSE] Directory: ./myapp
[INFO] Writing main.go
[INFO] Writing go.mod
...
```

## Code Agent Integration Patterns

### Pattern 1: Create and Validate
```bash
#!/bin/bash
set -euo pipefail

# Create project
OUTPUT=$(plumego new myapp --template api --format json)
PROJECT_PATH=$(echo "$OUTPUT" | jq -r '.data.path')

# Navigate
cd "$PROJECT_PATH"

# Validate
if ! plumego check --format json > health.json; then
  echo "Validation failed:"
  jq -r '.errors[]' health.json
  exit 1
fi
```

### Pattern 2: Extract Route Information
```bash
# Get all API routes
plumego routes --format json | \
  jq -r '.routes[] | select(.path | startswith("/api")) | .path'
```

### Pattern 3: CI/CD Health Check
```bash
# Start server in background
plumego dev --addr :8080 &
PID=$!

# Wait for ready
sleep 2

# Check health
if plumego inspect health --url http://localhost:8080; then
  echo "✓ Health check passed"
else
  echo "✗ Health check failed"
  exit 1
fi

# Cleanup
kill $PID
```

### Pattern 4: Configuration Validation
```bash
# Validate before deployment
plumego config validate --format json > validation.json

# Check for errors
ERRORS=$(jq -r '.errors | length' validation.json)
if [ "$ERRORS" -gt 0 ]; then
  echo "Configuration errors found:"
  jq -r '.errors[] | "  - \(.field): \(.message)"' validation.json
  exit 1
fi
```

## Why This Benefits Code Agents

### 1. **Deterministic Output**
Same input = same output = predictable parsing

### 2. **Structured Data**
No need to parse human-readable text with regex

### 3. **Clear Success/Failure**
Exit codes immediately tell if operation succeeded

### 4. **Self-Documenting**
JSON output includes field names and types

### 5. **Composable**
Can be combined with jq, grep, awk, etc.

### 6. **Automation-Ready**
No interactive prompts to break scripts

## Example: Full Workflow Automation

```bash
#!/bin/bash
# Automated project setup for code agents

set -euo pipefail

PROJECT_NAME="$1"
TEMPLATE="${2:-api}"

# 1. Create project
echo "Creating project $PROJECT_NAME..."
OUTPUT=$(plumego new "$PROJECT_NAME" \
  --template "$TEMPLATE" \
  --module "github.com/myorg/$PROJECT_NAME" \
  --format json)

# 2. Extract information
PROJECT_PATH=$(echo "$OUTPUT" | jq -r '.data.path')
echo "Project created at: $PROJECT_PATH"

# 3. Navigate
cd "$PROJECT_PATH"

# 4. Initialize dependencies
go mod tidy

# 5. Generate additional components
plumego generate component Auth --with-tests --format json
plumego generate middleware RateLimit --format json
plumego generate handler User --methods GET,POST,PUT,DELETE --format json

# 6. Run health check
plumego check --security --format json > health.json

# 7. Validate
STATUS=$(jq -r '.status' health.json)
if [ "$STATUS" != "healthy" ]; then
  echo "Health check failed:"
  jq -r '.checks | to_entries[] | "\(.key): \(.value.status)"' health.json
  exit 1
fi

# 8. Run tests
plumego test --race --cover --format json > test-results.json

# 9. Build
plumego build --output "./bin/$PROJECT_NAME" --format json

echo "✓ Project $PROJECT_NAME setup complete!"
echo "  Path: $PROJECT_PATH"
echo "  Binary: ./bin/$PROJECT_NAME"
echo "  Tests: $(jq -r '.passed' test-results.json)/$(jq -r '.tests' test-results.json) passed"
echo "  Coverage: $(jq -r '.coverage' test-results.json)%"
```

## Current Status

### Implemented ✅
- CLI framework with command routing
- Global flags (format, quiet, verbose, etc.)
- Output formatter (JSON, YAML, text)
- `plumego new` command with templates
- `plumego generate` command for components, handlers, middleware, and models
- `plumego dev` command with file watching and hot reload
- `plumego routes` command for route discovery
- `plumego check` command for health and security checks
- `plumego config` command for configuration management
- `plumego test` command with structured test output
- `plumego build` command with build metadata
- `plumego inspect` command for runtime inspection
- Project scaffolding system
- Exit code management
- Help system

### Next Steps 🚧
1. Implement `plumego migrate` for database migrations
2. Extend `plumego inspect` with richer endpoint adapters (per-app integration)
3. Add additional scaffolds/templates for specialized project types

### Future Enhancements 💡
- Plugin system for custom commands
- Template marketplace
- AI integration (`plumego ask`)
- Cloud deployment helpers
- Performance profiling
- Security scanning

## Documentation

- **Design Document**: [CLI_DESIGN.md](./CLI_DESIGN.md) - Complete specification
- **Quick Start**: [CLI_QUICK_START.md](./CLI_QUICK_START.md) - Getting started guide
- **This Document**: Summary and overview

## Installation

```bash
# From source
go install github.com/spcent/plumego/cmd/plumego@latest

# Build locally
cd /home/user/plumego
go build -o bin/plumego ./cmd/plumego

# Add to PATH
export PATH="$PATH:/home/user/plumego/bin"
```

## Testing

```bash
# Help
plumego --help

# Create project (dry run)
plumego new myapp --dry-run --format json

# Create actual project
plumego new myapp --template api --format yaml

# Verbose mode
plumego new myapp --verbose --format text
```

## Conclusion

This CLI design puts **code agents first** while remaining useful for human developers. Every decision prioritizes:

1. **Machine readability** over human aesthetics
2. **Automation** over interactive workflows
3. **Structured data** over prose
4. **Predictability** over flexibility
5. **Composability** over monolithic solutions

The result is a CLI that works seamlessly with:
- AI code assistants (Claude Code, GitHub Copilot)
- CI/CD pipelines (GitHub Actions, GitLab CI)
- Automation scripts (bash, Python, Node.js)
- DevOps tools (Terraform, Ansible)
- Testing frameworks (pytest, jest)

Making it a first-class tool for modern, AI-assisted development workflows.
