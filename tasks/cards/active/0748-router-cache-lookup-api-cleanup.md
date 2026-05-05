# Card 0748

Milestone: Router stable readiness
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P3
State: active
Primary Module: router
Owned Files: router/cache.go, router/dispatch.go, router/cache_coverage_test.go, tasks/cards/active/README.md
Depends On: 0747-router-reverse-url-failure-contract

Goal:
Remove duplicate internal match-cache lookup paths and the always-nil params
return.

Scope:
- Collapse `Get` and `Lookup` onto one internal implementation.
- Update dispatch to use a lookup shape that returns only the cached match and
  found flag.
- Update cache tests that asserted the obsolete nil params return.
- Preserve cache hit/miss accounting and LRU behavior.

Non-goals:
- Changing cache capacity or eviction policy.
- Exposing cache APIs.
- Reworking router matching.

Files:
- router/cache.go
- router/dispatch.go
- router/cache_coverage_test.go
- tasks/cards/active/README.md

Tests:
- go test -timeout 20s ./router/...
- go test -race -timeout 60s ./router/...
- go vet ./router/...

Docs Sync:
- Not required; internal-only cleanup.

Done Definition:
- No cache lookup method returns an always-nil params slice.
- Existing cache behavior tests pass.
- Router targeted tests, race tests, and vet pass.

Outcome:
