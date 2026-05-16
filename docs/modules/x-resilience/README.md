# x/resilience

## Purpose

`x/resilience` contains reusable resilience components that are not part of the stable core and do not belong to a single feature family.

## v1 Status

- `experimental` in the Plumego v1 support matrix
- Included in repository release scope, but compatibility is not frozen

## Use this module when

- the task is a reusable circuit breaker or similar resilience primitive
- the task is a reusable rate limiter or keyed limiter primitive
- the behavior should be shared across multiple extension families

## Do not use this module for

- app bootstrap
- stable security policy
- feature-specific orchestration that belongs in `x/ai`, `x/gateway`, or another owning extension
- AI provider fallback, provider request keying, or AI error classification

## First files to read

- `x/resilience/module.yaml`
- `x/resilience/circuitbreaker`
- `specs/extension-taxonomy.yaml`

## Public entrypoints

- `circuitbreaker.New`
- `circuitbreaker.NewMiddleware`
- `ratelimit.New`
- `ratelimit.NewKeyed`

## Main risks when changing this module

- breaker threshold regression
- keyed limiter behavior drift
- adapter behavior regression
- hidden shared state between callers

## Canonical change shape

- keep reusable resilience primitives here instead of in stable roots
- keep HTTP or transport adapters explicit and colocated with the primitive when they are generic
- keep feature-specific orchestration in the owning extension package
- keep AI-provider wrappers in `x/ai/resilience`

## Boundary rules

- `x/resilience` owns reusable circuit breaker and rate limit primitives; do not add these to stable `security` or stable `middleware`
- keep resilience primitive state instance-scoped; do not introduce package-level global state or implicit registration
- keep HTTP or transport adapters local to the owning extension when they are generic; do not push them into stable roots
- feature-specific orchestration (retry strategies tied to business rules) belongs in the owning extension, not in `x/resilience`
- `x/ai/circuitbreaker` and `x/ai/ratelimit` remain AI compatibility surfaces; do not move their exported symbols here without a dedicated symbol-change card

## Validation commands

- `go test -race -timeout 60s ./x/resilience/...`
- `go test -timeout 20s ./x/resilience/...`
- `go vet ./x/resilience/...`

For the detailed AI boundary decision, see
`docs/architecture/x-ai-resilience-boundary.md`.
