# Card 0674

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P0
State: active
Primary Module: router
Owned Files: router/router.go, router/registration.go, router/router_contract_test.go
Depends On:

Goal:
Make route registration normalize route paths consistently and invalidate stale route-match cache entries after successful registration.

Scope:
- Normalize route paths so root and grouped routes store leading-slash patterns consistently.
- Fix grouped registrations with child paths that omit a leading slash.
- Clear matcher caches after successful route registration so newly added more-specific routes are not hidden by earlier wildcard or param cache entries.
- Reject nil route handlers through the public `AddRoute` error path.

Findings:
- `Group("/api").AddRoute(GET, "users", ...)` currently registers `/apiusers`, which is not the caller-visible route shape.
- A request that populates a wildcard or parameter pattern cache can keep serving that older route after a more-specific route is registered later.
- Route metadata and `Routes()` snapshots can store paths without a leading slash when `AddRoute` is called with a relative path.
- Nil handlers are accepted during registration and can panic later during dispatch.

Non-goals:
- Do not add new route helper aliases.
- Do not expose cache controls as public API.
- Do not change named-route collision semantics.
- Do not change method matching or 405 behavior in this card.

Files:
- router/router.go
- router/registration.go
- router/router_contract_test.go

Tests:
- go test -race -timeout 60s ./router/...
- go test -timeout 20s ./router/...
- go vet ./router/...

Docs Sync:
Not required; this card preserves the intended router contract and closes implementation gaps.

Done Definition:
- Root, grouped, and relative child route registrations produce canonical leading-slash stored paths.
- New route registrations clear stale exact and pattern cache entries.
- Nil route handlers are rejected with an `AddRoute` error.
