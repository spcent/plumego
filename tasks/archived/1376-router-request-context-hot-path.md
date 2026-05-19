# Card 1376

Milestone: Router stable readiness
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P1
State: done
Primary Module: router
Owned Files: router/dispatch.go, router/matcher.go, router/router.go, router/types.go, contract/context_core.go, contract/context_test.go, contract/module.yaml, docs/modules/contract/README.md, docs/stable-api/snapshots/contract-head.snapshot, tasks/cards/done/1376-router-request-context-hot-path.md
Depends On: 9075-router-remove-obsolete-broad-tests

Goal:
Reduce router request-dispatch allocation and lock overhead on the hot path
without changing route matching, params, groups, reverse routing, or response
semantics.

Scope:
- Build router-owned `contract.RequestContext` directly during dispatch instead
  of reading and cloning any stale context first.
- Carry route names through matched trie nodes and `matchResult` so successful
  dispatch does not need a per-request metadata map lookup.
- Add a contract-owned read-only route-param accessor that returns one param
  without cloning the entire params map.
- Make `router.Param` use the read-only param accessor.
- Record focused benchmark evidence after the change.

Non-goals:
- Changing `RequestContextFromContext` defensive-copy behavior.
- Changing route matching precedence, cache key format, LRU behavior, or static
  file serving.
- Adding non-stdlib dependencies.
- Adding mutable request bags, service locators, or router-owned contract
  internals.

Files:
- router/dispatch.go
- router/matcher.go
- router/router.go
- router/types.go
- contract/context_core.go
- contract/context_test.go
- contract/module.yaml
- docs/modules/contract/README.md
- docs/stable-api/snapshots/contract-head.snapshot

Tests:
- go test -timeout 20s ./router/... ./contract/...
- go vet ./router/... ./contract/...
- go test -race -timeout 60s ./router/... ./contract/...
- go run ./internal/checks/dependency-rules
- go run ./internal/checks/module-manifests
- go run ./internal/checks/agent-workflow
- go run ./internal/checks/reference-layout
- bash scripts/check-stable-api-snapshots.sh
- go build ./...
- go test -run '^$' -bench 'BenchmarkOpt(StaticRoute|ParamRoute|ParallelParam)' -benchmem -count=5 ./router

Docs Sync:
- Updated contract module manifest and primer for the new read-only route-param
  accessor.
- Regenerated `docs/stable-api/snapshots/contract-head.snapshot`.

Done Definition:
- Router dispatch no longer clones stale `RequestContext` before overwriting it.
- Successful route-name propagation no longer requires a router metadata lookup
  during request dispatch.
- `router.Param` reads one route param without cloning the full params map.
- Focused router and contract tests pass.
- Focused benchmark evidence is recorded in Outcome.

Outcome:
- Added `contract.RequestParamFromContext(ctx, name)` for single-param reads
  without returning or cloning the whole params map.
- Changed `router.Param` to use the single-param accessor.
- Changed router dispatch to build a fresh `contract.RequestContext` directly
  instead of reading the existing context snapshot and overwriting it.
- Carried route names through trie nodes and `matchResult`, removing the
  successful-dispatch metadata lookup for route names.
- Benchmark evidence from the same local machine:
  - Static route stayed allocation-neutral at 480 B/op and 5 allocs/op; latency
    remained in the same range, roughly 173-211 ns/op after the change.
  - Param route cache hit improved from roughly 739-784 ns/op, 1784 B/op, 12
    allocs/op to roughly 415-437 ns/op, 1112 B/op, 8 allocs/op.
  - Param route cache miss improved from roughly 1029-1099 ns/op, 1986 B/op, 16
    allocs/op to roughly 677-729 ns/op, 1329 B/op, 12 allocs/op.
  - Parallel param route improved from roughly 418-435 ns/op to roughly
    377-389 ns/op with the same 1104 B/op and 8 allocs/op.
- Validation passed:
  - `go test -timeout 20s ./router/... ./contract/...`
  - `go vet ./router/... ./contract/...`
  - `go test -race -timeout 60s ./router/... ./contract/...`
  - `go run ./internal/checks/dependency-rules`
  - `go run ./internal/checks/module-manifests`
  - `go run ./internal/checks/agent-workflow`
  - `go run ./internal/checks/reference-layout`
  - `bash scripts/check-stable-api-snapshots.sh`
  - `go build ./...`
