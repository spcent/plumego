# Plan for M-023: AI Resilience Convergence

Milestone: `M-023`
Objective: Remove the highest-risk residual audit issue first by converging `x/ai/resilience` onto shared `x/resilience` primitives before any further manifest or control-plane cleanup.
Constraints: one primary module per card; max 5 files per card; max 3 validation commands per card; no new dependencies; remove dual first-class config paths instead of preserving them behind alternate legacy inputs.
Affected Modules: `x/ai`, `x/resilience`, `docs`, `specs`

## Phase Map

- Phase 1: converge rate-limit configuration onto shared primitives
- Phase 2: converge circuit-breaker configuration onto shared primitives
- Phase 3: validate the converged public surface and deprecation notes

## Card Inventory

| Card | Goal | Primary Module | Owned Files | Depends On | Quick Gates |
|------|------|----------------|-------------|------------|-------------|
| 2060 | Make shared `x/resilience/ratelimit` the only limiter input accepted by `x/ai/resilience`. | `x/ai` | `x/ai/resilience/provider.go`, `x/ai/resilience/provider_test.go`, `docs/modules/x-ai/README.md`, `specs/deprecation-inventory.yaml` | none | `go test ./x/ai/... ./x/resilience/...`, `deprecation-inventory -strict` |
| 2061 | Make shared `x/resilience/circuitbreaker` the only breaker input accepted by `x/ai/resilience`. | `x/ai` | `x/ai/resilience/provider.go`, `x/ai/resilience/provider_test.go`, `docs/modules/x-ai/README.md`, `specs/deprecation-inventory.yaml` | 2060 | `go test ./x/ai/... ./x/resilience/...`, `deprecation-inventory -strict` |

## Dependency Edges

- `2060 -> 2061`

## Parallel Groups

- Group A: `2060`
- Group B: `2061`

## Risk Register

- Risk: removing `x/ai/resilience` dual inputs turns into an unplanned breaking API migration.
  Mitigation: keep the config-path change isolated to this milestone and update tests, docs, and verification in the same change.
- Risk: the rate-limit and circuit-breaker halves diverge if implemented in isolation.
  Mitigation: keep both cards in the same milestone and reuse the same provider-level tests and docs path.

## Finding Disposition

- Verified and resolved: `x/ai/resilience/provider.go` now accepts only shared `x/resilience/ratelimit` and `x/resilience/circuitbreaker` inputs, so the former dual-stack configuration conflict is removed from the public config surface.
- Pulled ahead intentionally: this milestone is independent of the remaining control-plane cleanup so the runtime surface can be fixed first without waiting on docs/archive work.

## Verification Strategy

- Card-level checks: use the quick gates listed per card, plus focused package tests for `x/ai` and `x/resilience`.
- Milestone-level checks: `dependency-rules`, `deprecation-inventory -strict`, then focused race, normal, and vet runs across `x/ai` and `x/resilience`.

## Checkpoints

| Phase | Checkpoint Gate | Status |
|-------|-----------------|--------|
| Phase 1 | `go test -timeout 20s ./x/ai/... ./x/resilience/...` | passed |
| Phase 2 | `go test -race -timeout 60s ./x/ai/... ./x/resilience/...` | passed |
| Phase 3 | `go run ./internal/checks/deprecation-inventory -strict && go run ./internal/checks/dependency-rules` | passed |

## Exit Condition

- all planned cards completed or explicitly superseded
- all phase checkpoints recorded as passed
- verify report shows pass
- milestone acceptance criteria ready for PR packaging
