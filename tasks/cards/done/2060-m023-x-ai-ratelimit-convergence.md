# Card 2060

Milestone: M-023
Recipe: specs/change-recipes/symbol-change.yaml
Context Package: implementation
Priority: P0
State: done
Primary Module: x/ai
Owned Files:
- `x/ai/resilience/provider.go`
- `x/ai/resilience/provider_test.go`
- `x/ai/ratelimit/adapter.go`
- `docs/modules/x/ai/README.md`
- `specs/deprecation-inventory.yaml`
Depends On:

## Goal

Make shared `x/resilience/ratelimit` the canonical limiter input for
`x/ai/resilience` and move legacy `x/ai/ratelimit` support behind an explicit
compatibility adapter path instead of a parallel first-class config surface.

## Scope

Converge the rate-limit side of `x/ai/resilience.Config` so callers no longer
have to choose between sibling `RateLimiter` and `SharedRateLimiter` fields.
Keep any retained `x/ai/ratelimit` public API clearly marked as compatibility
surface only.

## Non-goals

- Do not remove `x/ai/ratelimit` exported symbols without same-change caller migration.
- Do not change circuit-breaker behavior in this card.
- Do not add new generic limiter algorithms outside `x/resilience/ratelimit`.

## Files

- `x/ai/resilience/provider.go`
- `x/ai/resilience/provider_test.go`
- `x/ai/ratelimit/adapter.go`
- `docs/modules/x/ai/README.md`
- `specs/deprecation-inventory.yaml`

## Acceptance Tests

- `x/ai/resilience/provider_test.go: TestNewResilientProviderE_UsesSharedRateLimiterAsCanonicalInput`
- `x/ai/resilience/provider_test.go: TestNewResilientProviderE_LegacyRateLimiterUsesCompatibilityAdapter`

## Tests

- `go test -timeout 20s ./x/ai/... ./x/resilience/...`
- `go run ./internal/checks/deprecation-inventory -strict`

## Docs Sync

- `docs/modules/x/ai/README.md`

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

- `x/ai/resilience.Config.RateLimiter` is now the canonical shared
  `x/resilience/ratelimit.KeyedBuckets` input.
- Legacy AI-local limiter wiring moved behind the explicit
  `LegacyRateLimiter` field plus `x/ai/ratelimit.NewCompatibilityAdapter(...)`.
- `ErrMultipleRateLimiters` now guards canonical-vs-legacy composition instead
  of shared-vs-shared duplication.
- Added `x-ai-ratelimit-compatibility-adapter` to the deprecation inventory and
  updated the `x/ai` module primer to document the migration path.
- Validation:
  - `go test -timeout 20s ./x/ai/... ./x/resilience/...`
  - `go run ./internal/checks/deprecation-inventory -strict`
  - `gofmt -l .`
