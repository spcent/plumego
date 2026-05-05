# Card 0739

Milestone:
Recipe: specs/change-recipes/middleware.yaml
Priority: P1
State: active
Primary Module: middleware
Owned Files:
- middleware/coalesce/coalesce.go
- middleware/coalesce/coalesce_test.go
- docs/modules/middleware/README.md
Depends On:
- 0738-middleware-stable-api-snapshot-and-conformance

Goal:
Make coalesced waiters return promptly when their request context is canceled.

Scope:
- Add `r.Context().Done()` handling while a waiter is waiting for an in-flight
  leader response.
- Decrement the waiter count on cancellation.
- Return without writing a synthetic timeout or upstream error response after
  cancellation.

Non-goals:
- Do not change leader request execution semantics.
- Do not change coalesce timeout status or response shape.
- Do not add new exported configuration.

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
- Canceled waiters return without invoking response replay.
- Existing timeout and successful replay behavior remain unchanged.
- Targeted and middleware-wide tests pass.

Outcome:

