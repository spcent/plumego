# Card 0101

Priority: P2

Goal:
- Make `Ctx.Errors` unexported (or read-only from the outside) so that the
  `Ctx.Error(err)` method is the only write path, preventing callers from
  directly mutating or clearing the collected errors.

Problem:

`context_core.go:53-56`:
```go
// Errors collects non-fatal errors encountered during request processing.
Errors []error
```

Because `Errors` is a public field, any caller can:
```go
ctx.Errors = nil                          // silently drops all collected errors
ctx.Errors = append(ctx.Errors, err)      // bypasses Ctx.Error(), skips return value
ctx.Errors[0] = differentErr             // direct element mutation
```

The intended write path is `Ctx.Error(err)` (context_core.go:328-333), which
appends and returns the error for inline use. But this method provides no
protection when callers write to the field directly.

A middleware that clears `ctx.Errors = nil` by accident would silently drop all
errors accumulated by earlier middleware in the chain, with no indication in
logs or responses.

Fix: unexport the field and add a read accessor.

```go
type Ctx struct {
    // ...
    errors []error   // unexported
}

// Errors returns a snapshot of the non-fatal errors collected so far.
// The returned slice is a copy; modifying it does not affect the context.
func (c *Ctx) CollectedErrors() []error {
    return append([]error(nil), c.errors...)
}
```

Update `Ctx.Error` to append to the unexported field. Update any code that
reads `ctx.Errors` directly.

Note on naming: `Errors` is taken as a method name by the pattern `error.Errors()`
(used in Go 1.20+ multi-error), so `CollectedErrors()` avoids a collision.

Migration:
- Run `grep -rn '\.Errors' . --include='*.go'` to find all read/write sites.
- Reads become `ctx.CollectedErrors()`.
- Direct appends become `ctx.Error(err)`.
- Direct clears (`ctx.Errors = nil`) must be evaluated case-by-case.

Non-goals:
- Do not change the behavior of `Ctx.Error(err)`.
- Do not add thread-safety to error collection (it is already not safe for
  concurrent use; that is an existing, separate limitation).

Files:
- `contract/context_core.go`
- All callers that access `.Errors`

Tests:
- Add a test: `ctx.CollectedErrors()` returns a copy; mutating the copy does not
  change `ctx.CollectedErrors()` in subsequent calls.
- `go test ./contract/...`
- `go vet ./...`
- `go build ./...`

Done Definition:
- `Ctx.Errors []error` field is unexported.
- `Ctx.CollectedErrors() []error` is the read accessor.
- `Ctx.Error(err)` is the write path.
- All callers updated.
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
