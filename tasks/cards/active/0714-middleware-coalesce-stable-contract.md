# Card 0714

Milestone:
Recipe: specs/change-recipes/middleware.yaml
Priority: P0
State: active
Primary Module: middleware
Owned Files:
- middleware/coalesce/coalesce.go
- middleware/coalesce/coalesce_test.go
- docs/modules/middleware/README.md
Depends On:

Goal:
Make coalesce safe enough to remain in stable middleware by clarifying its transport-only contract and hardening replay semantics.

Scope:
- Remove application-wiring examples from the package documentation.
- Clarify that coalesce is response coalescing, not business caching.
- Guard HEAD replay semantics so coalesced HEAD requests do not receive a body.
- Add focused tests for HEAD replay, response-size overflow, timeout waiters, and replay header/status behavior.

Non-goals:
- Do not move coalesce to x/* in this card.
- Do not change the public constructor surface.
- Do not add cache storage, persistence, or business key logic.

Files:
- middleware/coalesce/coalesce.go
- middleware/coalesce/coalesce_test.go
- docs/modules/middleware/README.md

Tests:
- go test ./middleware/coalesce
- go test ./middleware/...

Docs Sync:
- docs/modules/middleware/README.md

Done Definition:
- Coalesce docs describe only transport response coalescing.
- HEAD replay is explicitly covered and stable.
- Overflow and waiter timeout behavior have regression tests.
- Middleware tests pass.

Outcome:

