# Card 0513

Milestone:
Recipe: specs/change-recipes/http-endpoint-bugfix.yaml
Priority: P1
State: done
Primary Module: middleware
Owned Files: middleware/debug/debug_errors.go; middleware/debug/debug_errors_test.go; tasks/cards/done/0513-middleware-debug-streaming-boundary.md
Depends On: 2222

Goal:
Keep debug error replacement from rewriting streaming-style responses that declare streaming content after the request is accepted.

Scope:
- Treat response `Content-Type: text/event-stream` and stream-like content types as pass-through.
- Preserve JSON pass-through and plain-text error replacement behavior.
- Add tests for downstream SSE content type without relying only on request `Accept`.

Non-goals:
- Do not redesign debug middleware buffering.
- Do not add production error monitoring or logging behavior.
- Do not change canonical contract error shape.

Files:
- `middleware/debug/debug_errors.go`
- `middleware/debug/debug_errors_test.go`

Tests:
- `go test -timeout 20s ./middleware/debug`
- `go test -timeout 20s ./middleware/...`
- `go vet ./middleware/...`

Docs Sync:
- No docs sync required; this tightens an existing skip rule.

Done Definition:
- Debug errors no longer replace response-declared streaming content.
- Existing JSON/plain-text replacement tests pass.
- Targeted middleware tests and vet pass.

Outcome:
- Added a response content-type streaming guard to `shouldReplaceError`.
- Added a regression test for downstream `text/event-stream` error responses without request `Accept: text/event-stream`.
- Validation run: `go test -timeout 20s ./middleware/debug`; `go test -timeout 20s ./middleware/...`; `go vet ./middleware/...`.
