# Card 0515

Milestone:
Recipe: specs/change-recipes/http-endpoint-bugfix.yaml
Priority: P0
State: done
Primary Module: middleware
Owned Files: middleware/coalesce/coalesce.go; middleware/coalesce/coalesce_test.go; tasks/cards/done/0515-middleware-coalesce-key-and-timeout.md
Depends On: 2224

Goal:
Prevent coalescing from merging requests across virtual hosts and remove avoidable timeout timer churn.

Scope:
- Include `r.Host` in the default coalescing key.
- Include `r.Host` in header-aware coalescing keys.
- Replace `time.After` in waiter paths with an explicit timer that is stopped.
- Add tests proving same path on different hosts does not coalesce and same host still does.

Non-goals:
- Do not add cache storage or background eviction.
- Do not change public key function names.
- Do not add tenant- or route-aware key policy.

Files:
- `middleware/coalesce/coalesce.go`
- `middleware/coalesce/coalesce_test.go`

Tests:
- `go test -timeout 20s ./middleware/coalesce`
- `go test -timeout 20s ./middleware/...`
- `go vet ./middleware/...`

Docs Sync:
- No docs sync required; this is a correctness fix under existing URL-based key semantics.

Done Definition:
- Requests with different hosts no longer share in-flight responses.
- Waiter timeout path uses explicit timer cleanup.
- Targeted middleware tests and vet pass.

Outcome:
- Added `r.Host` to default and header-aware coalescing key generation.
- Replaced waiter `time.After` with `time.NewTimer` plus `Stop`.
- Added coverage for same path on different hosts and host-sensitive default keys.
- Validation run: `go test -timeout 20s ./middleware/coalesce`; `go test -timeout 20s ./middleware/...`; `go vet ./middleware/...`.
