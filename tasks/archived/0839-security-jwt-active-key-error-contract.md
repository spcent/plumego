# Card 0839

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P0
State: done
Primary Module: security
Owned Files:
- security/jwt/jwt.go
- security/jwt/jwt_test.go
- docs/modules/security/README.md
Depends On:

Goal:
Make JWT startup fail closed on real active signing-key store errors while preserving recovery for a genuinely missing active key.

Scope:
- Distinguish missing active key from active-key read failures without importing store packages into security.
- Preserve recovery when the persisted active key points to missing signing material.
- Add negative tests for active key read failures.

Non-goals:
- Do not add a concrete key-store backend.
- Do not change token wire format.
- Do not add session revocation or tenant policy.

Files:
- security/jwt/jwt.go
- security/jwt/jwt_test.go
- docs/modules/security/README.md

Tests:
- go test -timeout 20s ./security/jwt
- go vet ./security/jwt

Docs Sync:
- Update security module README if startup key-store semantics are clarified.

Done Definition:
- Missing active key still creates a new active key.
- Active-key store errors fail manager startup.
- Stale active key recovery remains covered.

Outcome:
- Updated JWT manager startup to read the active signing-key marker only when the marker is present in the key-store key list.
- Preserved recovery for missing active markers and stale active key IDs while failing startup on active marker read errors.
- Added focused tests for absent active marker recovery and active marker read failures.
- Documented the recoverable versus fail-closed active-key startup behavior.

Validation:
- `go test -timeout 20s ./security/jwt`
- `go vet ./security/jwt`
