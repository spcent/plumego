# Card 6203

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P1
State: active
Primary Module: internal/pool
Owned Files: internal/pool/pool.go, internal/pool/pool_test.go
Depends On: tasks/cards/done/6202-internal-nethttp-retry-context-backoff.md

Goal:
Prevent internal pools from retaining oversized slices, buffers, and map
references after temporary large use.

Scope:
- Drop oversized buffers and slices instead of returning them to `sync.Pool`.
- Clear map entries held inside `[]map[string]any` before pooling the slice.
- Add focused tests for capacity guards and map-reference clearing.

Non-goals:
- Do not change JSON marshal newline behavior.
- Do not change exported function names.
- Do not refactor extraction helpers in this card.

Files:
- internal/pool/pool.go
- internal/pool/pool_test.go

Tests:
- go test ./internal/pool
- go test ./internal/...

Docs Sync:
- None; internal retention guard only.

Done Definition:
- Oversized temporary values are not retained by pools.
- Map slice pools do not keep stale map contents reachable.

Outcome:
