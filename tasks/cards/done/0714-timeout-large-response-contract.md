# Card 0714

Milestone: —
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P0
State: done
Primary Module: middleware
Owned Files:
- `middleware/timeout/timeout.go`
- `middleware/timeout/timeout_test.go`
- `docs/modules/middleware/README.md`
Depends On: —

Goal:
- Fix timeout middleware semantics for responses that exceed the buffering threshold.

Problem:
`middleware/timeout.Timeout` currently switches to bypass mode once buffered data exceeds `StreamingThreshold`, discards subsequent writes, and later writes a 500 response. This contradicts the package comment that says bypass mode avoids memory spikes for large or streaming responses, and it makes otherwise valid large responses fail.

Scope:
- Decide and implement one explicit contract:
  - either reject large buffered responses deterministically before partial success is observable, or
  - support a documented pass-through mode that does not discard body bytes.
- Keep the middleware transport-only and continue using `middleware.WriteTransportError` for timeout/error responses.
- Add regression tests for a response just below threshold, above threshold, and timeout after partial writes.
- Update middleware docs to describe the real large-response behavior.

Non-goals:
- Do not introduce streaming framework APIs.
- Do not change `core.AppConfig` timeout fields.
- Do not change unrelated middleware constructors.

Files:
- `middleware/timeout/timeout.go`
- `middleware/timeout/timeout_test.go`
- `docs/modules/middleware/README.md`

Tests:
- `go test -timeout 20s ./middleware/timeout`
- `go test -timeout 20s ./middleware/...`
- `go vet ./middleware/...`

Docs Sync:
- Required if the large-response contract changes or becomes more explicit.

Done Definition:
- Large responses no longer silently discard body data and then return an unexpected 500.
- Tests cover below-threshold, above-threshold, and timeout paths.
- Docs match implemented behavior.
- Middleware boundary remains transport-only.

Outcome:
- Implemented timeout response-size rejection instead of the previous bypass/discard behavior.
- Responses above `StreamingThreshold` now return a canonical internal transport error before any buffered body is committed.
- Added regression coverage for below-threshold replay, above-threshold rejection, and timeout after a partial buffered write.
- Documented that `StreamingThreshold` is a hard replay limit, not a streaming passthrough switch.
- Validation passed:
  - `go test -timeout 20s ./middleware/timeout`
  - `go test -timeout 20s ./middleware/...`
  - `go vet ./middleware/...`
