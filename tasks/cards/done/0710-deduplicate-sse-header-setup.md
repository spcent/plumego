# Card 0710

Priority: P3

Goal:
- Eliminate the duplicated SSE header lines between `RespondWithSSE` and
  `SetSSEHeaders` by having `RespondWithSSE` delegate to `SetSSEHeaders`.

Problem:

`context_stream.go:500-506` — `RespondWithSSE`:
```go
func (c *Ctx) RespondWithSSE() (*SSEWriter, error) {
    c.W.Header().Set("Content-Type", "text/event-stream")
    c.W.Header().Set("Cache-Control", "no-cache")
    c.W.Header().Set("Connection", "keep-alive")
    c.W.WriteHeader(http.StatusOK)
    return NewSSEWriter(c.W)
}
```

`context_stream.go:514-519` — `SetSSEHeaders`:
```go
func (c *Ctx) SetSSEHeaders() {
    c.W.Header().Set("Content-Type", "text/event-stream")
    c.W.Header().Set("Cache-Control", "no-cache")
    c.W.Header().Set("Connection", "keep-alive")
}
```

Three identical `Header().Set` lines appear in both methods. If an SSE header
changes (e.g., adding `X-Accel-Buffering: no` for nginx), it must be updated
in two places.

Fix:
```go
func (c *Ctx) RespondWithSSE() (*SSEWriter, error) {
    c.SetSSEHeaders()
    c.W.WriteHeader(http.StatusOK)
    return NewSSEWriter(c.W)
}
```

Non-goals:
- Do not add or remove any SSE headers.
- Do not change `WriteSSE` or `IsSSESupported`.
- Do not change the behavior of either method.

Files:
- `contract/context_stream.go`

Tests:
- `go test ./contract/...`
- `go vet ./...`

Done Definition:
- `RespondWithSSE` body calls `c.SetSSEHeaders()` instead of the three inline
  `Header().Set` calls.
- Behavior is identical.
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
