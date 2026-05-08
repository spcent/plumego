# Card 0918

Milestone:
Recipe: specs/change-recipes/middleware.yaml
Priority: P1
State: done
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
- Queued requests now return promptly when `r.Context()` is canceled before a
  worker slot is available.
- Added regression coverage proving the canceled waiter does not invoke the
  downstream handler or write a synthetic queue timeout response.
- Documented the queued cancellation contract.

Validation:
- `go test -timeout 20s ./middleware/concurrencylimit`
- `go test -timeout 20s ./middleware/...`
