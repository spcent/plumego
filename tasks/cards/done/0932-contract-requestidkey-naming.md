# Card 0932

Priority: P1
State: done
Primary Module: contract
Owned Files:
- `contract/request_id.go`
Depends On: 0822

Goal:
- Rename `requestIDKey` to `requestIDContextKey` so that all context key type names in the package follow the same `*ContextKey` suffix convention.

Problem:
- The `contract` package defines four unexported context key types:
  - `requestContextKey` (context_core.go:28) — ✓ correct pattern
  - `traceContextKey` (trace.go:47) — ✓ correct pattern
  - `principalContextKey` (auth.go:41) — ✓ correct pattern
  - `requestIDKey` (request_id.go:8) — ✗ breaks the pattern (missing `Context` infix)
- The odd one out, `requestIDKey`, requires a reader to do a double-take to verify it is indeed a context key. The other three make their purpose explicit in the name.
- This is the only file in the package that names a context key type without the `Context` infix.

Scope:
- Rename `requestIDKey` → `requestIDContextKey` in `request_id.go`.
- No behavioral change; no exported symbols are affected.

Non-goals:
- Do not rename any exported functions or the `RequestIDHeader` constant.
- Do not change context key values or semantics.
- If card 0822 touches `request_id.go` first, fold this rename into that change instead of doing it separately.

Files:
- `contract/request_id.go`

Tests:
- `go test -timeout 20s ./contract/...`
- `go vet ./contract/...`

Docs Sync:
- None required.

Done Definition:
- No type named `requestIDKey` exists in the package.
- `requestIDContextKey` is used consistently in `request_id.go`.
- All tests pass.

Outcome:
- Pending.
