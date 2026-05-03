# Card 0123

Milestone:
Recipe: specs/change-recipes/stable-api.yaml
Priority: P1
State: active
Primary Module: middleware
Owned Files:
  - middleware/coalesce/coalesce.go
  - middleware/coalesce/coalesce_test.go
  - docs/modules/middleware/README.md
Depends On: 0122

Goal:
Bound coalesced response capture so request coalescing cannot grow memory
without limit when the upstream response is large.

Scope:
- Add a configurable response capture limit with a conservative default.
- If the leader response exceeds the limit, do not replay it to waiters; return a structured upstream failure to coalesced waiters.
- Clarify timeout behavior: waiting requests time out instead of executing their own fallback request.

Non-goals:
- Do not move `coalesce` to an extension package in this card.
- Do not change key generation semantics except as needed for tests.
- Do not introduce external dependencies.

Files:
- `middleware/coalesce/coalesce.go`
- `middleware/coalesce/coalesce_test.go`
- `docs/modules/middleware/README.md`

Tests:
- `go test -timeout 20s ./middleware/coalesce`
- `go test -race -timeout 60s ./middleware/coalesce`
- `go vet ./middleware/...`

Docs Sync:
- `docs/modules/middleware/README.md`

Done Definition:
- Coalesced response capture has a documented default bound.
- Oversized leader responses do not allocate unbounded waiter replay buffers.
- Targeted tests pass.

Outcome:

