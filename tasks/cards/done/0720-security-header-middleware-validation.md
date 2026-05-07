# Card 0720

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P0
State: done
Primary Module: middleware
Owned Files:
- middleware/security/headers.go
- middleware/security/headers_test.go
- docs/modules/security/README.md
- docs/modules/middleware/README.md
Depends On: 0719

Goal:
Make the canonical security-header middleware fail closed when supplied an invalid header policy.

Scope:
- Validate custom `headers.Policy` values when `middleware/security.SecurityHeaders` is constructed.
- Preserve the existing middleware function signature.
- Return a clear canonical error response instead of silently applying a partially invalid policy.
- Add focused middleware tests for invalid policy behavior and sync docs.

Non-goals:
- Do not change `headers.Policy.Apply` skip behavior for direct low-level callers.
- Do not add non-stdlib dependencies.
- Do not move HTTP middleware behavior into `security/headers`.

Files:
- `middleware/security/headers.go`
- `middleware/security/headers_test.go`
- `docs/modules/security/README.md`
- `docs/modules/middleware/README.md`

Tests:
- `go test -race -timeout 60s ./middleware/security ./security/headers`
- `go test -timeout 20s ./middleware/... ./security/...`
- `go vet ./middleware/... ./security/...`

Docs Sync:
- Required for middleware fail-closed policy behavior.

Done Definition:
- Invalid configured policies do not reach downstream handlers.
- Direct `Policy.Apply` remains a primitive that skips unsafe values.
- Middleware and security docs describe the distinction.
- Targeted middleware/security tests and vet pass.

Outcome:
- `middleware/security.SecurityHeaders` now validates custom policies and fails closed with a canonical internal error when the policy is invalid.
- Added middleware coverage that invalid policy config does not call downstream handlers.
- Kept direct `headers.Policy.Apply` primitive behavior unchanged.
- Synced security and middleware module docs.
- Validation run: `go test -race -timeout 60s ./middleware/security ./security/headers`; `go test -timeout 20s ./middleware/... ./security/...`; `go vet ./middleware/... ./security/...`; `go run ./internal/checks/dependency-rules`.
