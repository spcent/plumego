# Card 0608

Priority: P2

Goal:
- Remove the exported mutable pointer `DefaultRequestConfig` and replace it
  with an unexported default value, exposing a read-only accessor function
  so the global cannot be mutated.

Problem:
- `contract/context_core.go:224`:
  ```go
  var DefaultRequestConfig = &RequestConfig{
      MaxBodySize:    defaultMaxBodySize,
      MaxHeaderBytes: defaultMaxHeaderBytes,
      Timeout:        defaultTimeout,
  }
  ```
- This is an exported mutable pointer. Any package (including tests) can write
  `contract.DefaultRequestConfig.MaxBodySize = 0` and permanently alter the
  global default for the process lifetime.
- Tests currently mutate it (context_extended_test.go, context_test.go) causing
  potential test pollution if tests run in parallel.
- AGENTS.md §2 forbids hidden globals and mutable shared state at package level.

Scope:
- Change to an unexported constant-style value:
  ```go
  var defaultRequestConfig = RequestConfig{
      MaxBodySize:    defaultMaxBodySize,
      MaxHeaderBytes: defaultMaxHeaderBytes,
      Timeout:        defaultTimeout,
  }
  ```
- Export a read-only accessor (returns a copy, not a pointer):
  ```go
  // DefaultConfig returns a copy of the default request processing configuration.
  // Modify the returned value freely; it does not affect the package default.
  func DefaultConfig() RequestConfig {
      return defaultRequestConfig
  }
  ```
- Update `newCtxWithLogger` (context_core.go:232-235) to call `defaultRequestConfig`
  instead of `DefaultRequestConfig`.
- Update all test files that assign to `DefaultRequestConfig` to instead pass
  a custom config to `NewCtxWithConfig` (or whatever the constructor is called).

Non-goals:
- Do not add a `SetDefaultConfig` mutator; callers needing a custom default
  should inject config at construction time.
- Do not change `RequestConfig` struct fields.

Files:
- `contract/context_core.go`
- `contract/context_test.go` (tests that mutate DefaultRequestConfig)
- `contract/context_extended_test.go` (tests that mutate DefaultRequestConfig)
- Any other caller of `contract.DefaultRequestConfig`

Prerequisite:
```bash
grep -rn 'DefaultRequestConfig' . --include='*.go'
```
Confirm the list before editing.

Tests:
- `go test -race ./contract/...`
- `go build ./...`
- `grep -rn 'contract\.DefaultRequestConfig\b'` must return empty after migration

Done Definition:
- `contract.DefaultRequestConfig` does not exist as an exported symbol.
- `contract.DefaultConfig()` returns a copy of the default configuration.
- No test mutates global state; each test that needs custom config passes it
  explicitly.
- `go test -race ./contract/...` passes with no data race warnings.
