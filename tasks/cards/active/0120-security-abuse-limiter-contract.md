# Card 0120

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P1
State: active
Primary Module: security
Owned Files:
- security/abuse/limiter.go
- security/abuse/limiter_test.go
- security/abuse/limiter_negative_matrix_test.go
Depends On:
- 0119

Goal:
Make the abuse limiter constructor, zero-value behavior, and metrics surface clear and stable.

Scope:
- Align constructor behavior with `Config.Validate` without breaking `NewLimiter(Config{})`.
- Make zero-value limiter operations fail closed instead of panicking.
- Return metrics snapshots instead of exposing mutable internal counters.
- Add tests for invalid config, zero-value calls, stop idempotency, and metrics immutability.

Non-goals:
- Do not move limiter behavior to middleware.
- Do not add distributed or persistent rate limiting.
- Do not change the token bucket algorithm.

Files:
- security/abuse/limiter.go
- security/abuse/limiter_test.go
- security/abuse/limiter_negative_matrix_test.go

Tests:
- go test -timeout 20s ./security/abuse
- go test -race -timeout 60s ./security/abuse
- go vet ./security/abuse

Docs Sync:
- Update comments to describe zero-value and constructor behavior.

Done Definition:
- `NewLimiter(Config{})` still returns defaults.
- Explicit invalid non-zero config values fail through a new error-returning constructor or equivalent safe path.
- Zero-value `Limiter` methods do not panic.
- Metrics callers cannot mutate internal counters.

Outcome:

Validation:
