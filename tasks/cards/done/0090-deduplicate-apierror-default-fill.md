# Card 0090

Priority: P3

Goal:
- Remove the hand-rolled default-fill logic inside `toAPIError` by using
  `ErrorBuilder` consistently, eliminating a third copy of the same pattern
  that already exists in `WriteError` and `ErrorBuilder.Build()`.

Problem:

The logic "if Status is 0 set 500, if Code is empty set status text, if Category
is empty derive from status" appears in three places:

**1. `errors.go:167-181`** — `WriteError`:
```go
if err.Status == 0 { err.Status = http.StatusInternalServerError }
if err.Code == "" { err.Code = http.StatusText(err.Status) }
if err.Category == "" { ... }
```

**2. `errors.go:319-333`** — `ErrorBuilder.Build()`:
```go
if b.err.Status == 0 { b.err.Status = http.StatusInternalServerError }
if b.err.Code == "" { b.err.Code = http.StatusText(b.err.Status) }
if b.err.Category == "" { ... }
```

**3. `error_utils.go:43-56`** — `toAPIError`:
```go
if apiErr, ok := err.(APIError); ok {
    if apiErr.Status == 0 { apiErr.Status = http.StatusInternalServerError }
    if apiErr.Code == "" { apiErr.Code = http.StatusText(apiErr.Status) }
    if apiErr.Category == "" { ... }
    return apiErr
}
```

`toAPIError` also manually builds fallback `APIError` structs in two other
branches (lines 66-73 and 75-82) instead of using `NewErrorBuilder()`.

Any future change to the defaults must be applied in all three locations. The
`WriteError` copy is intentional (backward-compatibility guard at the boundary);
the `toAPIError` copy is not.

Fix:
- In `toAPIError`, when converting an `APIError`, pass it through
  `NewErrorBuilder()` and call `.Build()` to normalize it, instead of the
  manual three-field check.
- The two fallback `WrappedErrorWithContext` branches already use
  `NewErrorBuilder()` — ensure consistency by verifying they don't duplicate
  the pattern.

Non-goals:
- Do not remove the default-fill from `WriteError` (it is a public API
  boundary guard).
- Do not change `ErrorBuilder.Build()`.
- Do not change the observable behavior of `toAPIError`.

Files:
- `contract/error_utils.go`

Tests:
- `go test ./contract/...`
- `go vet ./...`

Done Definition:
- `toAPIError` does not contain the three-field default check for `APIError` values.
- `toAPIError` uses `ErrorBuilder` to normalize received `APIError` values.
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
