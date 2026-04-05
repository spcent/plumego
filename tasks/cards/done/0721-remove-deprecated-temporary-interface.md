# Card 0721

Priority: P2

Goal:
- Remove the use of the deprecated `Temporary() bool` interface from `IsRetryable`
  and replace it with a check that is valid in Go 1.18+.

Problem:

`error_wrap.go:129-131`:
```go
if netErr, ok := err.(interface{ Temporary() bool }); ok {
    return netErr.Temporary()
}
```

`net.Error.Temporary()` was deprecated in Go 1.18 (see golang.org/issue/45729).
The Go standard library's own implementations (e.g., `*net.OpError`) now always
return `false` from `Temporary()`, making this check a no-op for all stdlib
errors. Keeping the check:

1. Misleads readers into thinking it has meaningful effect on net errors.
2. Will continue to mislead as third-party libraries also drop `Temporary()`.
3. Couples the error-retryability logic to a deprecated interface contract.

The `Timeout() bool` check below it (lines 132-134) is still valid — timeouts
are retryable, and `net.Error.Timeout()` is not deprecated.

Fix:
- Remove lines 129-131 entirely.
- The `Timeout()` check covers the main net-error retryable case.
- If detecting generic transient conditions is needed, use `errors.As` against
  specific error types or a non-deprecated interface.

Non-goals:
- Do not change `IsRetryableError` (operates on `APIError`, unrelated).
- Do not change the `Timeout()` check.
- Do not add new retryability heuristics in this card.

Files:
- `contract/error_wrap.go`

Tests:
- Existing tests must pass unchanged.
- If a test asserts that `IsRetryable` returns true for an error implementing
  `Temporary()`, update the test to use `Timeout()` or a specific error type.
- `go test ./contract/...`
- `go vet ./...`

Done Definition:
- `IsRetryable` does not reference the `Temporary()` interface.
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
