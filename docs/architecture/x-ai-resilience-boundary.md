# x/ai and x/resilience Boundary

This document records the v1 cleanup decision for resilience-related extension
ownership. It is a boundary decision only; it does not migrate public types.

Read this before changing `x/ai/resilience`, `x/ai/circuitbreaker`,
`x/ai/ratelimit`, or `x/resilience`.

---

## Decision

`x/resilience` owns reusable resilience primitives for extension-layer use.
`x/ai/resilience` owns AI-provider resilience orchestration.

This means:

- generic circuit breakers, rate limiters, keyed limiters, and generic HTTP
  middleware adapters start in `x/resilience`
- AI provider wrappers, provider fallback composition, AI request keying,
  model/provider-aware policy, and AI error classification stay in `x/ai`
- stable roots such as `contract`, `middleware`, `security`, and `core` do not
  absorb extension resilience ownership

The current v1 cleanup path is documentation and internal convergence first.
Public type migration is out of scope until a later compatibility card defines
aliases, deprecations, tests, and release notes.

---

## Package Ownership

### `x/resilience`

Owns cross-extension primitives:

- circuit breaker state machines and options that do not know about AI providers
- token bucket and keyed rate limiter primitives
- generic middleware adapters colocated with the primitive
- instance-scoped state with explicit constructors

Does not own:

- provider fallback strategy
- prompt, model, session, tenant, or tool policy
- AI-specific error classification
- app bootstrap or hidden package-level policy

### `x/ai/resilience`

Owns AI-provider composition:

- wrapping `provider.Provider` implementations with retry, breaker, and rate
  limit behavior
- deciding how AI request fields map to limiter keys
- preserving provider stream and completion contracts
- handling AI-specific fallback and error classification

Does not own:

- new generic breaker or limiter algorithms
- reusable HTTP middleware for non-AI packages
- stable-root middleware or security policy

### `x/ai/circuitbreaker` and `x/ai/ratelimit`

These packages remain AI-local packages, but they are no longer alternate
configuration inputs to `x/ai/resilience`. They are not the landing zone for
new cross-family resilience features.

New generic work should start in `x/resilience`. New AI-provider wrapping may
continue in `x/ai/resilience`, but that wrapper should compose only shared
`x/resilience/*` primitives.

---

## Change Rules

- Do not migrate exported types between these packages without a dedicated
  symbol-change card.
- Do not require stable roots to import or understand AI resilience internals.
- Do not introduce hidden globals, `init()` registration, or implicit provider
  registries.
- Keep constructors explicit and error-returning for dynamic composition.
- Add feature-specific retry or fallback behavior to the owning extension, not
  to `x/resilience`.

---

## Examples

Use `x/resilience` when the same primitive can be reused by gateway, messaging,
data, or AI code without knowing feature semantics.

Use `x/ai/resilience` when the behavior depends on `provider.Provider`,
completion requests, streaming, model names, provider names, or AI error
classes.

---

## Follow-up Path

The safe migration path is:

1. keep shared `x/resilience/*` as the only resilience composition input to
   `x/ai/resilience`
2. add new generic breaker or limiter algorithms in `x/resilience`
3. keep any AI-local limiter or breaker usage explicit at the direct package
   call site instead of routing it back through `x/ai/resilience`
4. document future symbol removals or deprecations before removing any
   remaining exported AI-local symbols
