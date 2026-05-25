# Card 2061

Milestone: M-023
Recipe: specs/change-recipes/symbol-change.yaml
Context Package: implementation
Priority: P0
State: done
Primary Module: x/ai
Owned Files:
- `x/ai/resilience/provider.go`
- `x/ai/resilience/provider_test.go`
- `x/ai/circuitbreaker/adapter.go`
- `docs/modules/x-ai/README.md`
- `specs/deprecation-inventory.yaml`
Depends On: 2060

## Goal

Make shared `x/resilience/circuitbreaker` the canonical breaker input for
`x/ai/resilience` and move legacy `x/ai/circuitbreaker` support behind an
explicit compatibility adapter path.

## Scope

Converge the circuit-breaker side of `x/ai/resilience.Config` so callers no
longer face sibling `CircuitBreaker` and `SharedCircuitBreaker` public config
paths for the same behavior.

## Non-goals

- Do not remove `x/ai/circuitbreaker` exported symbols without same-change caller migration.
- Do not change rate-limit behavior in this card.
- Do not add new generic circuit-breaker algorithms outside `x/resilience/circuitbreaker`.

## Files

- `x/ai/resilience/provider.go`
- `x/ai/resilience/provider_test.go`
- `x/ai/circuitbreaker/adapter.go`
- `docs/modules/x-ai/README.md`
- `specs/deprecation-inventory.yaml`

## Acceptance Tests

- `x/ai/resilience/provider_test.go: TestNewResilientProviderE_UsesSharedCircuitBreakerAsCanonicalInput`
- `x/ai/resilience/provider_test.go: TestNewResilientProviderE_LegacyCircuitBreakerUsesCompatibilityAdapter`

## Tests

- `go test -timeout 20s ./x/ai/... ./x/resilience/...`
- `go run ./internal/checks/deprecation-inventory -strict`

## Docs Sync

- `docs/modules/x-ai/README.md`

## Validation

- `go test -timeout 20s ./x/ai/... ./x/resilience/...`
- `go run ./internal/checks/deprecation-inventory -strict`
- `gofmt -l .`

## Done Definition

- [x] Acceptance Tests pass.
- [x] All Validation commands exit 0.
- [x] gofmt -l . produces no output.
- [x] Docs Sync targets updated (if applicable).

## Outcome

- `x/ai/resilience.Config.CircuitBreaker` is now the canonical shared
  `x/resilience/circuitbreaker.CircuitBreaker` input.
- Legacy AI-local breaker wiring moved behind the explicit
  `LegacyCircuitBreaker` field plus
  `x/ai/circuitbreaker.NewCompatibilityAdapter(...)`.
- `ErrMultipleCircuitBreakers` now guards canonical-vs-legacy composition
  instead of AI-local-vs-shared duplication.
- Added `x-ai-circuitbreaker-compatibility-adapter` to the deprecation
  inventory and updated the `x/ai` module primer to document the breaker
  migration path.
- Validation:
  - `go test -timeout 20s ./x/ai/... ./x/resilience/...`
  - `go run ./internal/checks/deprecation-inventory -strict`
  - `gofmt -l .`
