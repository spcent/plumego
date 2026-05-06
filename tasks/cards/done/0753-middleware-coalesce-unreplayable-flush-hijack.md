# Card 0753

Milestone:
Recipe: specs/change-recipes/middleware.yaml
Priority: P1
State: done
Primary Module: middleware
Owned Files:
- middleware/coalesce/coalesce.go
- middleware/coalesce/coalesce_test.go
- docs/modules/middleware/README.md
Depends On:
- 0752-middleware-concurrency-queue-depth-detail

Goal:
Make coalesce refuse waiter replay when the leader response uses transport
operations that cannot be safely captured.

Scope:
- Mark leader responses as unreplayable when the coalesce recorder observes
  `Flush` or `Hijack`.
- Return the existing structured upstream failure to waiters for unreplayable
  leader responses.
- Add regression tests for flushed and hijacked leader responses with waiters.
- Document the unreplayable response contract.

Non-goals:
- Do not add streaming support to coalesce.
- Do not change the leader response path for normal bounded responses.
- Do not change coalesce key semantics.

Files:
- middleware/coalesce/coalesce.go
- middleware/coalesce/coalesce_test.go
- docs/modules/middleware/README.md

Tests:
- go test -timeout 20s ./middleware/coalesce
- go test -timeout 20s ./middleware/...

Docs Sync:
- docs/modules/middleware/README.md

Done Definition:
- Flush and Hijack leader paths are never replayed to waiters.
- Waiters receive canonical upstream failure for unreplayable leaders.
- Middleware-wide tests pass.

Outcome:
- Added an unreplayable response state to the coalesce leader recorder.
- Marked leader responses unreplayable after `Flush` or successful `Hijack`.
- Waiters now receive the existing structured upstream failure instead of
  replaying flushed or hijacked leader responses.
- Added regression coverage for flushed and hijacked leader paths.
- Documented the unreplayable response contract.

Validation:
- go test -timeout 20s ./middleware/coalesce
- go test -timeout 20s ./middleware/...
