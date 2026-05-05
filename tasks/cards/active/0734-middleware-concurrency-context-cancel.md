# Card 0734

Milestone:
Recipe: specs/change-recipes/middleware.yaml
Priority: P1
State: active
Primary Module: middleware
Owned Files:
- middleware/concurrencylimit/concurrency_limit.go
- middleware/concurrencylimit/concurrency_limit_test.go
- docs/modules/middleware/README.md
Depends On:
- 0733-middleware-timeout-safe-onpanic

Goal:
Respect request cancellation while a request is queued for an available
concurrency slot.

Scope:
- Add `r.Context().Done()` handling while waiting for `sem`.
- Return without writing a synthetic queue-timeout response after cancellation.
- Cover the cancellation path with tests.

Non-goals:
- Do not change queue-depth admission behavior.
- Do not add new exported configuration.
- Do not change normal queue timeout status or error shape.

Files:
- middleware/concurrencylimit/concurrency_limit.go
- middleware/concurrencylimit/concurrency_limit_test.go
- docs/modules/middleware/README.md

Tests:
- go test -timeout 20s ./middleware/concurrencylimit
- go test -timeout 20s ./middleware/...

Docs Sync:
- docs/modules/middleware/README.md

Done Definition:
- Queued canceled requests return promptly without invoking downstream handler.
- Existing busy and queue-timeout behavior remains unchanged.
- Targeted and middleware-wide tests pass.

Outcome:

