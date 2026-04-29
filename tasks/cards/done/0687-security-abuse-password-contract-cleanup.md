# Card 0687

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P2
State: done
Primary Module: security
Owned Files:
- security/abuse/limiter.go
- security/abuse/limiter_test.go
- security/password/password.go
- security/password/password_test.go
- security/password/password_strength_test.go
- docs/modules/security/README.md
Depends On: 0686

Goal:
Clean up limiter accounting and password error contracts so security primitives report state consistently without leaking secrets.

Scope:
- Fix shard-local eviction accounting so bucket counts decrement on the shard that owns the removed key.
- Keep limiter bucket metrics aligned with the tracked bucket count after creation, cleanup, and eviction.
- Remove unused limiter helpers that obscure the actual bucket lifecycle.
- Add exported password sentinel errors for invalid cost, invalid hash format, and password mismatch.
- Wrap password parse/verification failures with sentinels so callers can use `errors.Is`.
- Add focused tests for limiter bucket metrics and password sentinel errors.

Non-goals:
- Do not change rate-limit algorithm semantics.
- Do not add persistence, distributed quotas, or resilience primitives.
- Do not log or return password material.
- Do not enforce `MinimumCost` as a breaking behavior change in this card.

Files:
- `security/abuse/limiter.go`
- `security/abuse/limiter_test.go`
- `security/password/password.go`
- `security/password/password_test.go`
- `security/password/password_strength_test.go`
- `docs/modules/security/README.md`

Tests:
- `go test -race -timeout 60s ./security/abuse ./security/password`
- `go test -timeout 20s ./security/...`
- `go vet ./security/...`

Docs Sync:
- Required if password error contract or limiter metrics guidance changes.

Done Definition:
- Limiter bucket count and bucket metrics stay consistent after create, cleanup, and eviction.
- Password errors are classifiable with `errors.Is` without exposing password material.
- Targeted security tests and vet pass.

Outcome:
- Removed unused limiter bucket/time-cache helpers so the live bucket lifecycle is explicit.
- Fixed shard-local eviction to decrement the owning shard counter and keep bucket metrics synchronized after create, eviction, and cleanup.
- Added password sentinel errors for invalid cost, invalid stored hash, and password mismatch.
- Updated password parsing and verification errors to support `errors.Is` without exposing password material.
- Synced the security module primer with limiter metrics and password error classification guidance.
- Validation run: `go test -race -timeout 60s ./security/abuse ./security/password`; `go test -timeout 20s ./security/...`; `go vet ./security/...`.
