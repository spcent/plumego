# Card 0659

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P1
State: done
Primary Module: internal/nethttp
Owned Files: internal/nethttp/client.go, internal/nethttp/httpclient_test.go
Depends On: 0658

Goal:
Make internal nethttp request execution fail predictably for nil requests and normalize invalid retry counts.

Scope:
- Return a clear error from `Do` / internal dispatch when the request is nil.
- Clamp negative client-level and per-request retry counts to zero so one attempt is still made.
- Add focused tests for nil request and negative retry count behavior.

Non-goals:
- Do not change retry policy semantics for non-negative counts.
- Do not change high-level `Get` / `Post` method signatures.
- Do not alter SSRF, middleware order, or response body ownership.

Files:
- internal/nethttp/client.go
- internal/nethttp/httpclient_test.go

Tests:
- go test -timeout 20s ./internal/nethttp
- go test -timeout 20s ./internal/...
- go vet ./internal/...

Docs Sync:
Not required; internal defensive behavior only.

Done Definition:
- `Do(nil)` returns an error instead of panicking.
- Negative retry counts perform exactly one attempt.
- Focused and internal package validation pass.

Outcome:
Completed. `Do(nil)` now returns `ErrNilRequest`, and negative client-level or request-level retry counts are normalized to zero retries while preserving the initial attempt.

Validation:
- go test -timeout 20s ./internal/nethttp
- go test -timeout 20s ./internal/...
- go vet ./internal/...
