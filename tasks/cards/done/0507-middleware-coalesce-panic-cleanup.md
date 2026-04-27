# Card 0507

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P0
State: done
Primary Module: middleware
Owned Files: middleware/coalesce/coalesce.go; middleware/coalesce/coalesce_test.go; tasks/cards/done/0507-middleware-coalesce-panic-cleanup.md
Depends On: 2216

Goal:
Prevent request coalescing from leaving stale in-flight entries when downstream handlers panic or return unusable captured responses.

Scope:
- Ensure `Coalescer.executeRequest` always removes the in-flight key and releases waiters even when `next.ServeHTTP` panics.
- Re-panic after cleanup so recovery middleware ordering semantics remain intact.
- Guard response replay against nil captured responses with a structured upstream failure.
- Add regression tests for panic cleanup and waiter release.

Non-goals:
- Do not add cache storage, background eviction, or non-HTTP coalescing primitives.
- Do not change the public coalesce key API.
- Do not change recovery middleware behavior.

Files:
- `middleware/coalesce/coalesce.go`
- `middleware/coalesce/coalesce_test.go`

Tests:
- `go test -timeout 20s ./middleware/coalesce`
- `go test -timeout 20s ./middleware/...`
- `go vet ./middleware/...`

Docs Sync:
- No docs sync required; this is a bug fix under existing coalescing semantics.

Done Definition:
- A panic in the primary in-flight request does not strand `Stats().InFlight`.
- Waiting requests are released to a canonical error path when no captured response is available.
- Targeted middleware tests and vet pass.

Outcome:
- Added panic-safe in-flight cleanup with re-panic behavior for recovery middleware compatibility.
- Added a nil captured-response guard for waiters, returning canonical `upstream_failed`.
- Added a regression test proving waiters are released and `Stats().InFlight` returns to zero after primary panic.
- Validation run: `go test -timeout 20s ./middleware/coalesce`; `go test -timeout 20s ./middleware/...`; `go vet ./middleware/...`.
