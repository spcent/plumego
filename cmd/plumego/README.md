# Plumego CLI

A code agent-friendly command-line tool for plumego projects. Designed for automation, CI/CD integration, and AI-assisted development workflows.

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

# Create an API server
plumego new myapi --template api

# Create with custom module path
plumego new myapp --template fullstack --module github.com/myorg/myapp
```

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

## Commands

All 9 commands are fully implemented:

1. **new** - Create projects from templates
2. **generate** - Generate components, middleware, handlers, models
3. **dev** - Development server with hot reload
4. **check** - Health and security validation
5. **config** - Configuration management
6. **routes** - Route analysis and inspection
7. **build** - Build with optimizations
8. **test** - Enhanced test runner
9. **inspect** - Runtime inspection

See full documentation: [docs/CLI_DESIGN.md](../../docs/CLI_DESIGN.md)

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
