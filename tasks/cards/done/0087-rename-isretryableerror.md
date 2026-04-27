# Card 0087

Priority: P2

Goal:
- Rename `IsRetryableError(APIError) bool` to resolve the naming collision with
  `IsRetryable(error) bool` and clarify which function to call in each context.

Problem:

Two functions exist for checking retryability:

`errors.go:416`:
```go
// IsRetryableError checks if the error is retryable based on its status code.
func IsRetryableError(err APIError) bool { ... }
```

`error_wrap.go:113`:
```go
// IsRetryable reports whether err represents a transient condition that a caller
// may safely retry.
func IsRetryable(err error) bool { ... }
```

The name `IsRetryableError` implies "is this error retryable?" accepting the
`error` interface. Instead it accepts a concrete `APIError` struct value.
`IsRetryable` (without "Error") is the generic interface version. The names are
inverted relative to Go convention where the more specific variant carries the
qualifier.

A new caller reading both names will naturally reach for `IsRetryableError` for
any `error` value — and get a compile error because it doesn't accept `error`.
`IsRetryable` is the right function for `error` values, but its name suggests a
more limited scope.

Internally, `IsRetryable` already delegates to `IsRetryableError` when it
encounters an `APIError`.

Fix: Rename `IsRetryableError` to `IsAPIErrorRetryable` to make clear it
operates on the concrete `APIError` type. Update `IsRetryable` to call
`IsAPIErrorRetryable` at the delegation site.

Scope:
- Rename `IsRetryableError` → `IsAPIErrorRetryable` in `errors.go`.
- Update the call site in `error_wrap.go:119` and `error_wrap.go:123`.
- Grep all other callers: `grep -rn 'IsRetryableError' . --include='*.go'`

Non-goals:
- Do not change the behavior of either function.
- Do not merge the two functions.

Files:
- `contract/errors.go`
- `contract/error_wrap.go`
- All callers of `IsRetryableError`

Tests:
- `go test ./contract/...`
- `go vet ./...`
- `go build ./...`

Done Definition:
- `IsRetryableError` does not exist; `IsAPIErrorRetryable(APIError) bool` does.
- `IsRetryable(error)` delegates to `IsAPIErrorRetryable`.
- All tests pass.

Outcome:
- Completed in the 2026-04-05 contract cleanup batch.
- Verified as part of the shared contract/task-card completion pass.

Validation Run:
- `go test -timeout 20s ./contract/...`
- `go vet ./contract/...`
- `go run ./internal/checks/dependency-rules`
- `go run ./internal/checks/agent-workflow`
- `go run ./internal/checks/module-manifests`
- `go run ./internal/checks/reference-layout`
- `go build ./...`
- `go test -timeout 20s ./...`
- `go test -race -timeout 60s ./...`
- `go vet ./...`
