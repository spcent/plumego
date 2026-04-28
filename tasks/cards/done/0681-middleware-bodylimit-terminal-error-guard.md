# Card 0681

Milestone:
Recipe: specs/change-recipes/http-endpoint-bugfix.yaml
Priority: P0
State: done
Primary Module: middleware
Owned Files:
- middleware/bodylimit/body_limit.go
- middleware/bodylimit/body_limit_test.go
Depends On:

Goal:
Make `bodylimit.BodyLimit` terminal after it emits the canonical 413 response so downstream handler error handling cannot append a second plaintext or partial response body.

Scope:
- Wrap the response writer used by `BodyLimit` enough to suppress downstream writes after the limit reader has emitted the structured error.
- Keep exact-limit reads and disabled limits unchanged.
- Preserve optional response-writer capabilities already safe to forward.
- Add coverage proving a handler that reacts to the read error cannot pollute the canonical 413 envelope.

Non-goals:
- Do not replace request-body reading with binding or DTO behavior.
- Do not introduce a second body-limit constructor family.
- Do not change `contract` error response shape.

Files:
- `middleware/bodylimit/body_limit.go`
- `middleware/bodylimit/body_limit_test.go`

Tests:
- `go test -timeout 20s ./middleware/bodylimit`
- `go test -timeout 20s ./middleware/...`
- `go vet ./middleware/...`

Docs Sync:
- No docs sync required; this tightens an existing transport error path.

Done Definition:
- Over-limit reads still return status 413 and `REQUEST_BODY_TOO_LARGE`.
- Downstream `http.Error` or writes after the body-limit error are suppressed.
- Targeted middleware tests and vet pass.

Outcome:
- Added a body-limit response writer that suppresses downstream writes after the middleware emits the canonical 413.
- Preserved `Flusher` and `Hijacker` forwarding for handlers that depend on optional response-writer capabilities.
- Added regression coverage for handlers that call `http.Error` after seeing the body-limit read error.
- Validation run: `go test -timeout 20s ./middleware/bodylimit`; `go test -timeout 20s ./middleware/...`; `go vet ./middleware/...`.
