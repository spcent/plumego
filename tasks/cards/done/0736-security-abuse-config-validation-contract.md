# Card 0736

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P2
State: done
Primary Module: security
Owned Files:
- security/abuse/limiter.go
- security/abuse/limiter_test.go
- docs/modules/security/README.md

Goal:
Align abuse limiter `Config.Validate` with the canonical strict constructor semantics.

Scope:
- Make `Config.Validate` accept zero fields as omitted defaults and reject explicitly invalid negative values.
- Keep `NewLimiterWithConfig` as the canonical strict constructor.
- Keep `NewLimiter` compatibility fallback documented.
- Add tests proving `Validate` and `NewLimiterWithConfig` agree on zero and negative values.

Non-goals:
- Do not change limiter token bucket behavior.
- Do not remove `NewLimiter`.
- Do not change middleware wiring.

Files:
- security/abuse/limiter.go
- security/abuse/limiter_test.go
- docs/modules/security/README.md

Tests:
- go test -timeout 20s ./security/abuse
- go vet ./security/abuse
- go run ./internal/checks/dependency-rules

Docs Sync:
- Update security module docs if validation guidance changes.

Done Definition:
- `Config.Validate` and `NewLimiterWithConfig` agree on zero-as-omitted semantics.
- Negative explicit values fail closed.
- Targeted tests, vet, and dependency checks pass.

Outcome:
- Changed `Config.Validate` to use the same normalization path as `NewLimiterWithConfig`.
- Documented zero fields as omitted defaults and negative values as invalid.
- Added tests proving `Validate` and `NewLimiterWithConfig` agree on zero and negative values.

Validation:
- `gofmt -w security/abuse/limiter.go security/abuse/limiter_test.go`
- `go test -timeout 20s ./security/abuse`
- `go vet ./security/abuse`
- `go run ./internal/checks/dependency-rules`
