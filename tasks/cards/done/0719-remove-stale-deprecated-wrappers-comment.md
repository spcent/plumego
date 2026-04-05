# Card 0719

Priority: P3

Goal:
- Remove the stale "deprecated wrappers" comment from `context_stream.go:167`.

Problem:

`context_stream.go:167`:
```go
// --- shared slice/chunked helpers called by Stream and deprecated wrappers ---
```

The deprecated stream wrapper methods this comment refers to were removed in an
earlier cleanup wave. No deprecated wrappers remain in the file; the helpers
below this comment (`streamBinaryReader`, `streamJSONSliceChunked`, etc.) are
called only by `Stream` (the canonical dispatcher).

The stale comment misleads readers into looking for deprecated callers that no
longer exist, and may discourage removing the helpers if someone thinks they
still serve two purposes.

Scope:
- Replace the comment with one that accurately describes the section:
```go
// --- internal streaming helpers called by Stream ---
```

Non-goals:
- Do not change any function.
- Do not move or reorganize the helpers.

Files:
- `contract/context_stream.go`

Tests:
- `go build ./...`

Done Definition:
- Line 167 no longer references "deprecated wrappers".
- The replacement comment is accurate.

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
