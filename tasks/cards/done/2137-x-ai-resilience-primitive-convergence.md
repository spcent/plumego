# Card 2137: x/ai Resilience Primitive Convergence

Milestone: none
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P1
State: done
Primary Module: x/ai
Owned Files:
- `x/ai/resilience/provider.go`
- `x/ai/resilience/provider_test.go`
- `x/ai/circuitbreaker/circuitbreaker.go`
- `x/ai/ratelimit/ratelimit.go`
- `docs/modules/x-ai/README.md`
Depends On: none

Goal:
Converge AI-specific resilience wrappers so `x/ai` does not maintain a second
unclear copy of generic circuit breaker and rate-limit primitives beside
`x/resilience`.

Problem:
`x/ai/resilience` imports local `x/ai/circuitbreaker` and `x/ai/ratelimit`
packages, while the repository already has canonical reusable primitives under
`x/resilience/circuitbreaker` and `x/resilience/ratelimit`. The AI-specific
copies use different configuration names, state models, lifecycle behavior, and
test expectations. `x/ai/ratelimit.NewTokenBucketLimiterWithConfig` can also
start a cleanup goroutine without an explicit stop path. This makes future
resilience behavior hard to reason about and contradicts the `x/resilience`
family role.

Scope:
- Decide whether the AI wrapper should adapt the canonical `x/resilience`
  primitives or explicitly document why the local AI primitives remain
  feature-specific.
- If adapting, migrate `x/ai/resilience/provider.go` and its tests to use
  adapter interfaces over `x/resilience` primitives without changing the
  stable-tier `provider.Provider` contract.
- If local primitives remain, add narrow docs and tests that explain their
  feature-specific differences and add explicit lifecycle coverage for cleanup
  behavior.
- Add nil-provider and nil-request guard coverage for `NewResilientProvider`,
  `Complete`, and `CompleteStream` so invalid composition fails predictably.

Non-goals:
- Do not change provider request or response wire shapes.
- Do not move generic resilience primitives into stable roots.
- Do not redesign `x/resilience` itself in this card.
- Do not add third-party dependencies.

Files:
- `x/ai/resilience/provider.go`
- `x/ai/resilience/provider_test.go`
- `x/ai/circuitbreaker/circuitbreaker.go`
- `x/ai/ratelimit/ratelimit.go`
- `docs/modules/x-ai/README.md`

Tests:
- `go test -race -timeout 60s ./x/ai/resilience ./x/ai/circuitbreaker ./x/ai/ratelimit`
- `go test -timeout 20s ./x/ai/resilience ./x/ai/circuitbreaker ./x/ai/ratelimit`
- `go vet ./x/ai/resilience ./x/ai/circuitbreaker ./x/ai/ratelimit`

Docs Sync:
Update `docs/modules/x-ai/README.md` to state the canonical relationship between
AI-specific resilience wrappers and the reusable `x/resilience` primitives.

Done Definition:
- There is one documented resilience ownership story for AI provider wrapping.
- Invalid resilient-provider composition does not panic through nil dereference.
- Background cleanup behavior is either removed, explicitly stoppable, or
  documented as intentionally absent from the canonical path.
- The listed validation commands pass.

Outcome:
- Added `NewResilientProviderE` as the explicit error-returning construction
  path and kept `NewResilientProvider` as the compatibility wrapper.
- Added `ErrNilProvider` and `ErrNilRequest` coverage so invalid provider
  composition and nil completion requests fail predictably instead of nil
  dereferencing.
- Added `TokenBucketLimiter.Close` and idempotency coverage for cleanup-enabled
  limiters.
- Documented the compatibility relationship between `x/ai` resilience wrappers
  and canonical reusable `x/resilience` primitives.

Validation:
- `go test -race -timeout 60s ./x/ai/resilience ./x/ai/circuitbreaker ./x/ai/ratelimit`
- `go test -timeout 20s ./x/ai/resilience ./x/ai/circuitbreaker ./x/ai/ratelimit`
- `go vet ./x/ai/resilience ./x/ai/circuitbreaker ./x/ai/ratelimit`
