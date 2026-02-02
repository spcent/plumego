# Plumego CLI

A code agent-friendly command-line interface for the plumego HTTP toolkit.

## Features

- **JSON-first output**: Default structured output for easy parsing
- **Non-interactive**: All operations via flags and config
- **Predictable exit codes**: Clear success/failure indicators
- **Multiple formats**: JSON, YAML, or plain text output
- **Composable commands**: Unix philosophy for automation

## Installation

```bash
# Install from source
go install github.com/spcent/plumego/cmd/plumego@latest

# Or build locally
go build -o plumego .
```

## Quick Start

```bash
# Create new project
plumego new myapp --template api

# Get JSON output
plumego new myapp --format json

# Preview without creating
plumego new myapp --dry-run
```

## Commands

- `new` - Create project from template ✅
- `generate` - Generate code (🚧 coming soon)
- `dev` - Development server with hot reload (🚧 coming soon)
- `routes` - Inspect routes (🚧 coming soon)
- `check` - Health validation (🚧 coming soon)
- `config` - Configuration management (🚧 coming soon)
- `test` - Test runner (🚧 coming soon)
- `build` - Build utilities (🚧 coming soon)
- `inspect` - Runtime inspection (🚧 coming soon)

## Documentation

See [docs/](../../docs/) for detailed documentation:
- [CLI_DESIGN.md](../../docs/CLI_DESIGN.md) - Complete design specification
- [CLI_QUICK_START.md](../../docs/CLI_QUICK_START.md) - Getting started guide
- [CLI_SUMMARY.md](../../docs/CLI_SUMMARY.md) - Overview and summary

## Examples

### Create API Server
```bash
plumego new myapi --template api --module github.com/myorg/myapi
cd myapi
go mod tidy
go run main.go
```

### Parse JSON Output
```bash
PROJECT_PATH=$(plumego new myapp --format json | jq -r '.data.path')
cd "$PROJECT_PATH"
```

### CI/CD Integration
```bash
#!/bin/bash
if plumego check --format json > health.json; then
  echo "✓ Validation passed"
else
  jq -r '.errors[]' health.json
  exit 1
fi
```

## Architecture

```
cmd/plumego/
├── main.go              # Entry point
├── commands/            # Command implementations
│   ├── root.go         # Command dispatcher
│   ├── new.go          # Project scaffolding
│   └── stubs.go        # Stub implementations
└── internal/            # Internal packages
    ├── output/         # Output formatting
    └── scaffold/       # Project templates
```

## Development

```bash
# Build
go build -o plumego .

# Test
./plumego --help
./plumego new testapp --dry-run --format json

# Install locally
go install
```

## Design Principles

1. **Machine-first**: Default to JSON for agent parsing
2. **Non-interactive**: No prompts, all via flags
3. **Idempotent**: Safe to run multiple times
4. **Composable**: Works with pipes and tools
5. **Predictable**: Clear exit codes and structure

## Status

- ✅ Core framework implemented
- ✅ `plumego new` command working
- ✅ JSON/YAML/Text output formats
- ✅ Project scaffolding system
- 🚧 Other commands (stubs in place)

## Contributing

This CLI follows plumego's standard library-first philosophy:
- Minimal dependencies (only gopkg.in/yaml.v3 for YAML)
- Standard library patterns
- Clear, testable code

## License

Same as plumego main project.
