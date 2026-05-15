# Card 1422

Milestone: v1-breaking-normalization
Recipe: specs/change-recipes/stable-root-boundary-review.yaml
Priority: P0
State: active
Primary Module: store
Owned Files:
- store/cache/*
- store/kv/*
- store/file/*
- docs/modules/store/README.md
Depends On:
- 1421

Goal:
- Remove store compatibility aliases and wrappers so v1 exposes one canonical
  cache, kv, and file contract surface.

Scope:
- Enumerate all cache and kv aliases/wrappers before editing.
- Remove compatibility aliases such as legacy cache-miss names when a canonical
  sentinel exists.
- Remove kv compatibility wrappers and migrate every caller to canonical
  options/interfaces.
- Keep backend implementations out of stable `store`.
- Update tests and docs for the final v1 contract.

Non-goals:
- Do not import `x/*`.
- Do not move persistence backends into `store`.
- Do not add new storage capabilities.

Files:
- store/cache/cache.go
- store/kv/compat.go
- store/kv/kv.go
- store/kv/options.go
- store/**/*_test.go

Tests:
- go test -timeout 20s ./store/...
- go vet ./store/...
- go run ./internal/checks/dependency-rules

Docs Sync:
- Update store docs for removed aliases and canonical v1 names.

Done Definition:
- Removed aliases/wrappers have no remaining Go callers.
- Store tests pass.
- Stable store packages remain independent of `x/*`.

Outcome:
- Removed `store/cache.ErrCacheMiss`; `ErrNotFound` is now the single cache miss
  sentinel.
- Removed `store/kv` no-error compatibility methods: `Exists`, `Keys`, `Size`,
  and `GetStats`.
- Migrated store tests, scheduler KV persistence, and JWT key management to
  context-aware KV APIs.
- Updated store docs to describe the v1 breaking normalized surface.
