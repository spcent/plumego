# Card 0714

Milestone: Router stable readiness
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P0
State: active
Primary Module: router
Owned Files: router/cache.go, router/dispatch.go, router/cache_coverage_test.go, router/router_contract_test.go
Depends On:

Goal:
Keep cached route dispatch identical to cold trie dispatch for overlapping static, param, and wildcard routes.

Scope:
- Remove or replace parameterized pattern-cache behavior that can bypass trie precedence.
- Keep exact-path route cache behavior for already resolved request paths.
- Add regression tests proving warm cache does not change route selection.

Non-goals:
- Public API changes.
- New router configuration knobs.
- Cross-module routing changes.

Files:
- router/cache.go
- router/dispatch.go
- router/cache_coverage_test.go
- router/router_contract_test.go

Tests:
- go test -timeout 20s ./router/...
- go test -race -timeout 60s ./router/...
- go vet ./router/...

Docs Sync:
- Not required unless cache semantics are documented.

Done Definition:
- Warm-cache dispatch returns the same handler and params as cold dispatch for static/param/wildcard overlap.
- Router tests and vet pass.

Outcome:

