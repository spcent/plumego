# Plumego v1.0 GA Production Scope

Status: Approved  
Date: 2026-03-10  
Decision: Option A (Core GA + Experimental Isolation)

## Scope

v1.0 GA production commitments apply to:

- `core/`
- `router/`
- `middleware/`
- `contract/`
- `security/`
- `store/`
- top-level `plumego` convenience exports for the above stable APIs

These modules are covered by v1.x compatibility guarantees when used through
their canonical APIs.

## Explicitly Out of GA Scope (Experimental)

- `tenant/*`
- `net/mq/*`

These modules are available but remain experimental in v1.0. They are not
covered by v1.0 GA compatibility guarantees and may evolve before stabilization.

## Production Policy

1. No new canonical handler/route/error styles in v1.x.
2. Stable modules must pass full quality gates on every release candidate:
   - `go test -race -timeout 60s ./...`
   - `go test -timeout 20s ./...`
   - `go vet ./...`
   - `gofmt -l .` (must be empty)
3. Any behavior/config/security change in stable modules requires synchronized
   updates to:
   - `README.md`
   - `README_CN.md`
   - `AGENTS.md`
   - `CLAUDE.md`
   - `env.example`

## Graduation Criteria For Experimental Modules

Before `tenant/*` or `net/mq/*` can be promoted from experimental:

1. End-to-end integration coverage across routing + middleware + persistence
   paths.
2. Negative-path security and isolation tests.
3. Stable API declaration with migration notes.
4. Production cookbook and runnable reference example.
5. Release-cycle soak period with no unresolved P0/P1 regressions.
