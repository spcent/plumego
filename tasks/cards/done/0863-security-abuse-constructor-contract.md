# Card 0863

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P1
State: done
Primary Module: security
Owned Files:
- security/abuse/limiter.go
- security/abuse/limiter_test.go
- docs/modules/security/README.md
Depends On:
- 0728

Goal:
Clarify abuse limiter constructor behavior so production callers use the error-returning constructor.

Scope:
- Keep `NewLimiter` compatibility behavior but document it as lenient.
- Make `NewLimiterWithConfig` the canonical strict constructor in comments and docs.
- Add tests proving strict constructor rejects explicit invalid values and lenient constructor falls back.

Non-goals:
- Do not change middleware wiring in this card.
- Do not change limiter algorithm or metrics.
- Do not remove `NewLimiter`.

Files:
- security/abuse/limiter.go
- security/abuse/limiter_test.go
- docs/modules/security/README.md

Tests:
- go test -timeout 20s ./security/abuse
- go vet ./security/abuse

Docs Sync:
- Update security README constructor guidance.

Done Definition:
- Strict and lenient constructor semantics are explicit.
- Production guidance prefers `NewLimiterWithConfig`.
- Existing callers remain compatible.

Outcome:
- Documented `NewLimiterWithConfig` as the canonical strict production constructor.
- Clarified `NewLimiter` as the lenient compatibility constructor that falls back to defaults on invalid explicit values.
- Added a focused regression test covering lenient fallback behavior.

Validation:
- `gofmt -w security/abuse/limiter.go security/abuse/limiter_test.go`
- `go test -timeout 20s ./security/abuse`
- `go vet ./security/abuse`
