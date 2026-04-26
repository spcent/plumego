# Card 2215

Milestone:
Recipe: specs/change-recipes/http-endpoint-bugfix.yaml
Priority: P0
State: done
Primary Module: middleware
Owned Files: middleware/bodylimit/body_limit.go; middleware/bodylimit/body_limit_test.go; tasks/cards/done/2215-middleware-bodylimit-overrun-detection.md
Depends On: 2214

Goal:
Make request body limit enforcement detect one-byte overruns promptly and keep the middleware error path structured.

Scope:
- Fix the reader so a body with more than `maxBytes` bytes fails as soon as the overrun is observed, not only after a later caller happens to read again.
- Keep the structured `request_body_too_large` transport error path and redacted logging behavior.
- Add focused tests for exact-limit bodies, oversized bodies, and disabled limits.

Non-goals:
- Do not add streaming parsers, body buffering, or content-type-specific validation.
- Do not change handler decode conventions.
- Do not alter security/auth modules.

Files:
- `middleware/bodylimit/body_limit.go`
- `middleware/bodylimit/body_limit_test.go`

Tests:
- `go test -timeout 20s ./middleware/bodylimit`
- `go test -timeout 20s ./middleware/...`
- `go vet ./middleware/...`

Docs Sync:
- No docs sync required; behavior remains documented body limit enforcement.

Done Definition:
- Oversized bodies return the canonical JSON error without relying on a second read by downstream code.
- Exact-limit bodies still pass through unchanged.
- Targeted middleware tests and vet pass.

Outcome:
- Updated the limited reader to probe one byte past the configured maximum and fail immediately when overrun data exists.
- Kept exact-limit and disabled-limit requests passing through.
- Added tests for exact-limit, single-read overrun detection, and disabled limits.
- Validation run: `go test -timeout 20s ./middleware/bodylimit`; `go test -timeout 20s ./middleware/...`; `go vet ./middleware/...`.
