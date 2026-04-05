# Card 0706

Priority: P3

Goal:
- Rename `RequestContextFrom` to `RequestContextFromContext` to follow the
  `*FromContext` naming pattern used by every other context accessor in the
  package.

Problem:

All context accessor functions in `contract` follow the `<Thing>FromContext`
naming convention:

| Function | File |
|---|---|
| `TraceContextFromContext` | trace.go:48 |
| `TraceIDFromContext` | trace.go:58 |
| `ParamsFromContext` | context_core.go:165 |
| `PrincipalFromContext` | auth.go:49 |

The one exception is:

`context_core.go:182`:
```go
func RequestContextFrom(ctx context.Context) RequestContext { ... }
```

This function is missing the `Context` suffix. It was likely named before the
convention solidified. The inconsistency creates noise when browsing the API or
using IDE autocomplete.

Scope:
- Rename `RequestContextFrom` → `RequestContextFromContext`.
- Grep all callers: `grep -rn 'RequestContextFrom[^C]' . --include='*.go'`
- Update `AdaptCtxHandler` (context_core.go:416) which calls `RequestContextFrom`.

Non-goals:
- Do not rename `WithRequestContext` (its name is already consistent).
- Do not change the function body.

Files:
- `contract/context_core.go`
- All callers of `RequestContextFrom`

Tests:
- `go test ./...`
- `go vet ./...`
- `go build ./...`

Done Definition:
- `RequestContextFrom` does not exist; `RequestContextFromContext` does.
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
