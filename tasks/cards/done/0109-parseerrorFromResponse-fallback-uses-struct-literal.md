# Card 0109

Priority: P3

Goal:
- Replace the raw `APIError{}` struct literal in `ParseErrorFromResponse`'s
  decode-failure branch with `NewErrorBuilder()` to match the package pattern.

Problem:

`errors.go:393-399`:
```go
if err := json.NewDecoder(resp.Body).Decode(&errorResp); err != nil {
    return APIError{
        Status:   resp.StatusCode,
        Code:     http.StatusText(resp.StatusCode),
        Message:  fmt.Sprintf("failed to parse error response: %v", err),
        Category: CategoryServer,
    }, nil
}
```

This is one of the few remaining places in the package that constructs an
`APIError` via a raw struct literal instead of `NewErrorBuilder()`. The
inconsistency has two consequences:

1. If `ErrorBuilder.Build()` defaults ever change (e.g., a new required field
   is added), this branch won't pick up the change automatically.
2. `Type` and `Severity` fields (once card 0703 is complete) would be absent
   from errors created here.
3. The literal construction bypasses the `ValidateError` checks that callers
   using the builder get for free.

Fix:
```go
if err := json.NewDecoder(resp.Body).Decode(&errorResp); err != nil {
    return NewErrorBuilder().
        Status(resp.StatusCode).
        Code(http.StatusText(resp.StatusCode)).
        Category(CategoryServer).
        Type(ErrTypeInternal).
        Message(fmt.Sprintf("failed to parse error response: %v", err)).
        Build(), nil
}
```

Non-goals:
- Do not change the function signature or when it is called.
- Do not change the success path.

Files:
- `contract/errors.go`

Tests:
- Existing tests must pass unchanged.
- `go test ./contract/...`
- `go vet ./...`

Done Definition:
- `ParseErrorFromResponse` fallback branch uses `NewErrorBuilder()`.
- No raw `APIError{}` struct literal remains in the file for error construction
  (audit with `grep -n 'APIError{' errors.go`).
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
