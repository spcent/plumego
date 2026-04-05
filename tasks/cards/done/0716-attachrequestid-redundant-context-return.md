# Card 0716

Priority: P3

Goal:
- Remove the redundant `context.Context` return value from
  `AttachRequestID` since it is always equal to the returned request's context.

Problem:

`observability_policy.go:59-71`:
```go
func (p ObservabilityPolicy) AttachRequestID(
    ctx context.Context,
    w http.ResponseWriter,
    r *http.Request,
    id string,
    includeInRequest bool,
) (*http.Request, context.Context) {
    if r == nil {
        return r, ctx
    }
    ctx = WithTraceIDString(ctx, id)     // ctx updated
    // ...
    return r.WithContext(ctx), ctx       // returned ctx == returned r.Context()
}
```

On line 70, `r.WithContext(ctx)` creates a new request whose `.Context()` method
returns exactly `ctx`. The caller therefore receives two values that point to the
same context — one via the `*http.Request` and one as a bare `context.Context`.

Callers must destructure both values:
```go
r, ctx = policy.AttachRequestID(ctx, w, r, id, true)
// ctx == r.Context() always; the standalone ctx is redundant
```

Any code that uses the returned `context.Context` directly (instead of `r.Context()`)
is bypassing the idiomatic Go pattern where context flows through the request.
The redundant return encourages this anti-pattern.

Fix: Return only `*http.Request`.
```go
func (p ObservabilityPolicy) AttachRequestID(
    ctx context.Context,
    w http.ResponseWriter,
    r *http.Request,
    id string,
    includeInRequest bool,
) *http.Request
```

Migration: Callers that destructure `r, ctx = ...` must be updated to
`r = ...` and then use `r.Context()` where ctx was used.

Scope:
- Change the return type of `AttachRequestID`.
- Grep all callers: `grep -rn 'AttachRequestID' . --include='*.go'`
- Update each caller.

Non-goals:
- Do not change the behavior of the function (what headers are set, etc.).
- Do not change `RequestIDFromRequest` or `MiddlewareLogFields`.

Files:
- `contract/observability_policy.go`
- All callers of `AttachRequestID`

Tests:
- `go test ./...`
- `go vet ./...`
- `go build ./...`

Done Definition:
- `AttachRequestID` returns `*http.Request` only.
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
