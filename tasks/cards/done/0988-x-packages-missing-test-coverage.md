# Card 0988: Add Missing Test Coverage for Untested x Packages

Priority: P2
State: done
Recipe: specs/change-recipes/fix-bug.yaml
Primary Module: x/cache, x/tenant, x/devtools/pubsubdebug, x/observability/recordbuffer

## Goal

Several x package directories have Go source files but zero test files.  The required gate
`go test -timeout 20s ./...` passes vacuously for these packages — it reports no tests, not a
failure.  This leaves entire packages with no behavioral contract enforced by CI.

Packages with source files but no `*_test.go` files (confirmed by directory scan):

| Package | Note |
|---|---|
| `x/cache` | Distributed cache top-level package; only internal subpackage has tests |
| `x/tenant` | Tenant entry-point package; subpackages have tests, top-level has none |
| `x/devtools/pubsubdebug` | PubSub debug handler; no tests |
| `x/observability/recordbuffer` | Record buffer utility; no tests |
| `x/ai/semanticcache/cachemanager` | Semantic cache manager; no tests |

## Scope Per Package

### `x/cache`

`x/cache/cache.go` exports `Cache` interface and `New` / `NewDistributed` factory functions.
Write smoke tests that:
- Call `New()` with a nil store and assert the returned implementation satisfies `Cache`.
- Confirm `Get`, `Set`, `Delete` round-trips on the in-memory implementation.

### `x/tenant`

`x/tenant/tenant.go` (or equivalent entry point) exports tenant type aliases and convenience
constructors.  Write tests that:
- Confirm exported type aliases are non-nil / correctly forwarded.
- Confirm any constructor functions return non-zero values.

### `x/devtools/pubsubdebug`

`x/devtools/pubsubdebug/component.go` exports `Handler` with HTTP endpoints.  Write tests
using `httptest` that:
- POST to the publish endpoint with a valid payload → 200.
- POST with a missing topic → 400.
- GET status endpoint → 200 with JSON body.

### `x/observability/recordbuffer`

Write tests that cover the buffer's accumulation and flush semantics:
- Records appended before flush appear in the flushed batch.
- Buffer respects any configured capacity limit.
- Concurrent appends do not race (run with `-race`).

### `x/ai/semanticcache/cachemanager`

Write unit tests for the cache manager's core operations (store, lookup, eviction if
applicable).  Use a stub embedding store.

## Non-goals

- Do not add integration tests requiring external services (Redis, databases, etc.).
- Do not increase coverage inside subpackages that already have tests.
- Do not add benchmarks — unit tests only.

## Tests

```bash
go test -timeout 20s -race \
  ./x/cache/... \
  ./x/tenant/... \
  ./x/devtools/pubsubdebug/... \
  ./x/observability/recordbuffer/... \
  ./x/ai/semanticcache/cachemanager/...
go vet ./x/cache/... ./x/tenant/... ./x/devtools/... ./x/observability/... ./x/ai/...
```

## Done Definition

- Each listed package directory contains at least one `*_test.go` file with ≥1 passing test.
- `go test -race` passes for all listed packages.
- `go vet` clean.
- No test file uses `t.Skip` as its only body.

## Outcome

Completed. Three test files added (x/cache and x/tenant skipped — doc-only packages):

- `x/observability/recordbuffer/collector_test.go`: 7 tests covering append,
  timestamp backfill, capacity cap (WithMaxRecords), clear, copy isolation,
  concurrent safety, and ObserveHTTP.
- `x/devtools/pubsubdebug/component_test.go`: 7 tests covering New,
  RegisterRoutes disabled/enabled, default/custom path, snapshot handler
  200/500 paths, and Health state.
- `x/ai/semanticcache/cachemanager/manager_test.go`: 7 tests covering
  NewManager, Stats on empty backend, Cleanup (all/unknown), NewWarmer,
  WarmFromQueries (empty/populated) using stub provider + MemoryBackend.

`go test -race` passes for all three packages.
