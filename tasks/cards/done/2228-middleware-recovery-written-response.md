# Card 2228

Milestone:
Recipe: specs/change-recipes/http-endpoint-bugfix.yaml
Priority: P0
State: done
Primary Module: middleware
Owned Files: middleware/recovery/recover.go; middleware/recovery/recover_test.go; tasks/cards/done/2228-middleware-recovery-written-response.md
Depends On: 2227

Goal:
Avoid writing a second canonical error body when recovery catches a panic after downstream code has already started the response.

Scope:
- Wrap downstream writes enough to know whether headers/body were already sent.
- On panic before any response is written, keep the existing structured 500 error behavior.
- On panic after response start, log the panic but do not append another error response.
- Preserve recovery panic propagation boundaries and nil logger validation.

Non-goals:
- Do not buffer all recovery responses.
- Do not change logger interfaces.
- Do not redesign recovery ordering semantics.

Files:
- `middleware/recovery/recover.go`
- `middleware/recovery/recover_test.go`

Tests:
- `go test -timeout 20s ./middleware/recovery`
- `go test -timeout 20s ./middleware/...`
- `go vet ./middleware/...`

Docs Sync:
- No docs sync required; this tightens error-path behavior.

Done Definition:
- Panic before write still returns canonical 500.
- Panic after a response write does not append a second error body.
- Targeted middleware tests and vet pass.

Outcome:
- Added a small recovery response writer to track whether the response has started.
- Recovery now writes the canonical 500 only when no response has been sent.
- Added tests for early panic error writing and late panic no-append behavior.
- Validation run: `go test -timeout 20s ./middleware/recovery`; `go test -timeout 20s ./middleware/...`; `go vet ./middleware/...`.
