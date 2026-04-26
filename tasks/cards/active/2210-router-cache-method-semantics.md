# Card 2210: Router Cache Method Semantics

Milestone: none
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P0
State: active
Primary Module: router
Owned Files:
- `router/dispatch.go`
- `router/cache.go`
- `router/router_contract_test.go`
Depends On: 2209

Goal:
Keep cached dispatch behavior identical to uncached dispatch for HEAD fallback,
ANY fallback, route metadata, and 405 checks.

Problem:
The matcher cache is split between exact and parameterized routes. The exact
path cache is keyed by request method, but parameterized pattern lookup is also
method-scoped and does not account for HEAD-to-GET fallback or ANY fallback.
That makes cache behavior harder to reason about than the trie matcher.

Scope:
- Ensure parameterized cache lookup follows the same method fallback order as
  trie matching.
- Preserve HEAD response body suppression when a cached GET route handles HEAD.
- Preserve route pattern and route name context values on cached fallback
  matches.
- Add focused tests for cached HEAD and ANY parameterized routes.

Non-goals:
- Do not replace the cache implementation.
- Do not add public cache configuration.
- Do not change 404/405 response shapes.

Files:
- `router/dispatch.go`
- `router/cache.go`
- `router/router_contract_test.go`

Tests:
- `go test -race -timeout 60s ./router/...`
- `go test -timeout 20s ./router/...`
- `go vet ./router/...`

Docs Sync:
No docs change required; this aligns existing cache internals with documented
dispatch behavior.

Done Definition:
- Cached parameterized HEAD requests use GET fallback consistently.
- Cached parameterized ANY fallback requests use the ANY route metadata
  consistently.
- Existing router cache tests pass.

Outcome:

