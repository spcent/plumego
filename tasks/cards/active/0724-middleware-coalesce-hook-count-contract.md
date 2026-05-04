# Card 0724

Milestone:
Recipe: specs/change-recipes/middleware.yaml
Priority: P2
State: active
Primary Module: middleware
Owned Files:
- middleware/coalesce/coalesce.go
- middleware/coalesce/coalesce_test.go
- docs/modules/middleware/README.md
Depends On:
- 0723-middleware-gzip-panic-finalization

Goal:
Make coalesce callback count semantics clear and deterministic.

Scope:
- Define `OnCoalesced` count as a stable per-waiter event count.
- Avoid reporting the mutable total waiter count on each callback.
- Document key collision and callback semantics.
- Add tests covering multiple waiters and waiter timeout interaction.

Non-goals:
- Do not change the `OnCoalesced` function signature.
- Do not add a new metrics interface.
- Do not replace the default key hash algorithm.

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
- `OnCoalesced` reports deterministic per-callback count semantics.
- Mixed timeout/success waiter behavior is tested.
- Coalesce package and middleware-wide tests pass.

Outcome:

