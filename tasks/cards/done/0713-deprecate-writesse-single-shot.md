# Card 0713

Priority: P2

Goal:
- Deprecate `Ctx.WriteSSE` because it creates a new `SSEWriter` on every call
  and encourages an inefficient usage pattern.

Problem:

`context_stream.go:521-528`:
```go
func (c *Ctx) WriteSSE(event SSEEvent) error {
    sw, err := NewSSEWriter(c.W)
    if err != nil {
        return err
    }
    return sw.Write(event)
}
```

Every call to `WriteSSE` creates a fresh `SSEWriter`, which in turn asserts
the Flusher interface on `c.W`. In the intended SSE usage pattern — writing
multiple events in a loop — each iteration pays this overhead needlessly.

More importantly, the "correct" API already exists: `RespondWithSSE()` returns
an `*SSEWriter` that the caller holds and invokes in a loop. `WriteSSE` teaches
the wrong pattern by making single-shot look idiomatic.

`WriteSSE` also cannot detect when it is called before `RespondWithSSE` has
written the SSE response headers. A caller could call `WriteSSE` on an
uninitialized response, bypassing the required header setup entirely.

Fix: Add a `// Deprecated:` doc comment pointing to `RespondWithSSE` +
`sw.Write()`. Do not remove yet (callers must be audited first).

If callers are found, provide a migration path in the deprecation comment:
```go
// Deprecated: WriteSSE creates a new SSEWriter on every call and bypasses
// SSE header initialization. Use RespondWithSSE to obtain an *SSEWriter and
// call sw.Write() for each event:
//
//   sw, err := c.RespondWithSSE()
//   if err != nil { ... }
//   for _, event := range events {
//       if err := sw.Write(event); err != nil { ... }
//   }
```

Scope:
- Add the `// Deprecated:` doc comment to `WriteSSE`.
- Grep all callers: `grep -rn '\.WriteSSE(' . --include='*.go'`
- Migrate any callers found.

Non-goals:
- Do not remove `WriteSSE` in this card (a follow-up can delete it once callers
  are confirmed absent).
- Do not change `RespondWithSSE` or `SSEWriter.Write`.

Files:
- `contract/context_stream.go`
- Any callers found by grep

Tests:
- `go test ./contract/...`
- `go vet ./...`
- `go build ./...`

Done Definition:
- `WriteSSE` carries a `// Deprecated:` doc comment referencing `RespondWithSSE`.
- All callers migrated to the `RespondWithSSE`+`sw.Write` pattern.
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
