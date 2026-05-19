# Card 1135

Milestone:
Recipe: specs/change-recipes/middleware.yaml
Priority: P3
State: done
Primary Module: middleware
Owned Files:
- middleware/concurrencylimit/concurrency_limit.go
- middleware/concurrencylimit/concurrency_limit_test.go
- docs/modules/middleware/README.md
Depends On:
- 0751-middleware-recovery-conformance-matrix

Goal:
Make concurrency-limit queue timeout details describe what is actually being
reported.

Scope:
- Replace ambiguous `queue_depth` timeout detail with a name/value that matches
  the current implementation.
- Add or update tests asserting timeout detail fields.
- Document the timeout detail semantics.

Non-goals:
- Do not redesign the semaphore/queue algorithm.
- Do not add new public metrics surfaces.
- Do not change status code or error code.

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
- Timeout detail naming is accurate.
- Tests assert the detail field.
- Middleware-wide tests pass.

Outcome:
- Replaced the ambiguous `queue_depth` timeout detail with
  `queue_occupancy` and `queue_capacity`.
- Documented that queue occupancy reflects the internal queue channel
  occupancy, including active and waiting requests.
- Added regression coverage for the timeout error detail payload.

Validation:
- go test -timeout 20s ./middleware/concurrencylimit
- go test -timeout 20s ./middleware/...
