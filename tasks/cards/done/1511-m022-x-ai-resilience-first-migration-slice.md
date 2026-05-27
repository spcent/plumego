# Card 1511

Milestone: M-022
Recipe: specs/change-recipes/symbol-change.yaml
Context Package: implementation
Priority: P1
State: done
Primary Module: x/ai
Owned Files:
- `x/ai/resilience/provider.go`
- `x/ai/resilience/provider_test.go`
- `x/resilience/ratelimit/ratelimit.go`
- `docs/modules/x/ai/README.md`
Depends On: 1507

Goal:
- Land the first verified convergence slice from `x/ai`-local resilience
  wrappers toward shared `x/resilience/*` primitives without migrating or
  deleting existing exported `x/ai/circuitbreaker` or `x/ai/ratelimit` types.

Scope:
- Let `x/ai/resilience.Config` accept shared `x/resilience/circuitbreaker` and
  `x/resilience/ratelimit` primitives through explicit adapter fields.
- Keep existing `x/ai` compatibility fields working unchanged.
- Add the minimum `x/resilience/ratelimit` capability needed by the adapter.
- Document the new preferred composition path in `docs/modules/x/ai/README.md`.

Non-goals:
- Do not remove `x/ai/circuitbreaker` or `x/ai/ratelimit`.
- Do not rename existing exported AI resilience fields or constructors.
- Do not migrate public symbol ownership across packages in this card.

Files:
- `x/ai/resilience/provider.go`
- `x/ai/resilience/provider_test.go`
- `x/resilience/ratelimit/ratelimit.go`
- `docs/modules/x/ai/README.md`

Acceptance Tests:
- `go test -timeout 20s ./x/ai/resilience ./x/resilience/...`

Tests:
- `go test -timeout 20s ./x/ai/... ./x/resilience/...`
- `go run ./internal/checks/dependency-rules`

Docs Sync:
- `docs/modules/x/ai/README.md`

Validation:
- `go test -timeout 20s ./x/ai/... ./x/resilience/...`
- `go run ./internal/checks/dependency-rules`
- `gofmt -l .`

Done Definition:
- [x] Acceptance Tests pass.
- [x] All Validation commands exit 0.
- [x] gofmt -l . produces no output.
- [x] Docs Sync targets updated (if applicable).

Outcome:
- `x/ai/resilience.Config` now supports direct composition with shared
  `x/resilience/circuitbreaker.CircuitBreaker` and
  `x/resilience/ratelimit.KeyedBuckets` through explicit `Shared*` fields while
  retaining the older AI-local compatibility fields.
- The resilient provider now resolves explicit adapter paths, rejects
  conflicting dual configuration, maps shared circuit-open failures back onto
  the AI compatibility error, and preserves the existing `CircuitBreakerState`,
  `CircuitBreakerStats`, and rate-limit query helpers.
- `x/resilience/ratelimit` gained a minimal `Remaining` API so AI wrappers can
  query shared keyed limiter capacity without reintroducing a second token
  bucket implementation.
- Validation:
  - `go test -timeout 20s ./x/ai/... ./x/resilience/...`
  - `go run ./internal/checks/dependency-rules`
  - `gofmt -l .`
