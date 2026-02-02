# Plumego CLI - Independent Module

## Overview

The Plumego CLI (`cmd/plumego`) is an **independent Go module** with its own `go.mod` file. This separation ensures that CLI-specific dependencies (like YAML support) don't affect the core plumego library.

## Module Structure

```
plumego/                           # Main module
├── go.mod                         # Core plumego (no extra dependencies)
└── cmd/plumego/                   # CLI module
    ├── go.mod                     # Independent module with CLI dependencies
    ├── go.sum                     # CLI dependency checksums
    └── ...                        # CLI source code
```

## Why Independent Module?

### Benefits:
1. **Dependency Isolation**: CLI dependencies (like `gopkg.in/yaml.v3`) don't pollute the core plumego module
2. **Cleaner Core**: Projects using plumego as a library get zero extra dependencies
3. **Flexible Tooling**: CLI can use any dependencies without affecting core
4. **Better Compatibility**: Core library remains ultra-lightweight

### Dependencies:
- **Core plumego**: `go 1.24` - **zero external dependencies**
- **CLI**: `go 1.24.7` - `gopkg.in/yaml.v3` for YAML output support

## How It Works

### cmd/plumego/go.mod:
```go
module github.com/spcent/plumego/cmd/plumego

go 1.24.7

// Use local plumego package
replace github.com/spcent/plumego => ../..

require (
    github.com/spcent/plumego v0.0.0-00010101000000-000000000000
    gopkg.in/yaml.v3 v3.0.1
)
```

The `replace` directive tells Go to use the local plumego code instead of fetching it remotely.

## Building

### From Repository Root:
```bash
# Build CLI
cd cmd/plumego
go build -o ../../bin/plumego .

# Or use the project bin directory
go build -o /home/user/plumego/bin/plumego .
```

### Install Globally:
```bash
cd cmd/plumego
go install
```

This installs `plumego` to `$GOPATH/bin` (or `$HOME/go/bin`).

## Development

### Adding Dependencies to CLI:
```bash
cd cmd/plumego
go get <package>
go mod tidy
```

Dependencies are **only** added to `cmd/plumego/go.mod`, not the core.

### Adding Dependencies to Core:
```bash
cd /home/user/plumego
go get <package>
go mod tidy
```

Core dependencies affect both core and CLI (via replace directive).

## Testing

### Test CLI Builds:
```bash
cd cmd/plumego
go build .
```

### Test Core Builds (without CLI deps):
```bash
cd /home/user/plumego
go build ./...   # Should not include cmd/plumego
```

### Verify Independence:
```bash
# Check core has no YAML dependency
cd /home/user/plumego
grep yaml go.mod  # Should return nothing

# Check CLI has YAML dependency
cd cmd/plumego
grep yaml go.mod  # Should show gopkg.in/yaml.v3
```

## CI/CD Considerations

### Building in CI:
```yaml
# Build core library
- name: Build core
  run: |
    cd /home/user/plumego
    go build ./core
    go build ./router
    go build ./middleware
    # ... other packages

# Build CLI (separate step)
- name: Build CLI
  run: |
    cd /home/user/plumego/cmd/plumego
    go build -o plumego .
```

### Testing in CI:
```yaml
# Test core
- name: Test core
  run: |
    cd /home/user/plumego
    go test ./...

# Test CLI
- name: Test CLI
  run: |
    cd /home/user/plumego/cmd/plumego
    go test ./...
```

## Release Process

When releasing the CLI:

1. **Tag the main repository** (includes CLI):
   ```bash
   git tag v1.0.0
   git push origin v1.0.0
   ```

2. **Users install CLI** with:
   ```bash
   go install github.com/spcent/plumego/cmd/plumego@v1.0.0
   ```

3. **Projects use core** with:
   ```go
   import "github.com/spcent/plumego/core"
   ```
   ```bash
   go get github.com/spcent/plumego@v1.0.0
   ```

## FAQ

### Q: Why not use build tags?
**A**: Build tags would still require the dependencies to be listed in the main `go.mod`. This approach completely separates them.

### Q: Can I use the CLI as a library?
**A**: Yes, but you'll get the YAML dependency. If you only need CLI functionality without YAML output, we recommend copying the relevant code.

### Q: Does this affect go get?
**A**: No. Users can still `go get github.com/spcent/plumego` and get the core library without CLI dependencies.

### Q: What about the replace directive?
**A**: It only affects local development and building from source. When installed via `go install`, Go handles the versioning correctly.

## Verification Commands

```bash
# Verify core module has no extra dependencies
cd /home/user/plumego
go mod graph | grep -v "github.com/spcent/plumego"
# Should output nothing or only Go standard library

# Verify CLI module dependencies
cd /home/user/plumego/cmd/plumego
go mod graph | grep yaml
# Should show: gopkg.in/yaml.v3

# Build and test everything
cd /home/user/plumego
go build ./core ./router ./middleware  # Core packages
cd cmd/plumego
go build .  # CLI
./plumego --help  # Test it works
```

## Summary

This independent module approach ensures:
- ✅ Core plumego remains dependency-free
- ✅ CLI can use any tooling dependencies
- ✅ Clean separation of concerns
- ✅ No breaking changes for existing users
- ✅ Simple and maintainable structure

The `replace` directive makes local development seamless while keeping the modules logically separate.
