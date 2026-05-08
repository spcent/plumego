# Card 1027

Milestone:
Recipe: specs/change-recipes/middleware.yaml
Priority: P2
State: done
Primary Module: middleware
Owned Files:
- middleware/ratelimit/abuse_guard.go
- middleware/ratelimit/abuse_guard_test.go
- docs/modules/middleware/README.md
Depends On:
- 0742-middleware-accesslog-redaction

Goal:
Define deterministic behavior when a custom rate-limit key function returns a
blank key.

Scope:
- Fall back to the direct peer IP when `KeyFunc` returns an empty or all-blank
  key.
- Add regression coverage proving blank custom keys do not share one global
  limiter bucket.
- Document the blank-key fallback.

Non-goals:
- Do not change default direct peer-IP behavior.
- Do not add trusted proxy parsing to defaults.
- Do not change rate limit response schema.

Files:
- middleware/ratelimit/abuse_guard.go
- middleware/ratelimit/abuse_guard_test.go
- docs/modules/middleware/README.md

Tests:
- go test -timeout 20s ./middleware/ratelimit
- go test -timeout 20s ./middleware/...

Docs Sync:
- docs/modules/middleware/README.md

Done Definition:
- Blank custom keys fall back to direct peer IP.
- Existing rate-limit behavior remains unchanged for nonblank keys.
- Targeted and middleware-wide tests pass.

Outcome:
- Blank custom rate-limit keys now fall back to the direct peer IP.
- Added regression coverage proving different fallback peers do not share one
  empty-key limiter bucket.
- Documented the fallback contract.

Validation:
- `go test -timeout 20s ./middleware/ratelimit`
- `go test -timeout 20s ./middleware/...`
