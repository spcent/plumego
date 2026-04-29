# Card 0695

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P1
State: done
Primary Module: security
Owned Files:
- security/jwt/jwt.go
- security/jwt/jwt_test.go
- docs/modules/security/README.md
Depends On: 0694

Goal:
Harden JWT verification against malformed compact token envelopes before expensive decode or semantic work.

Scope:
- Reject compact JWTs with empty header, payload, or signature segments.
- Add conservative maximum token and segment sizes for verification input.
- Verify the signature before unmarshalling payload claims.
- Add focused regression tests for empty segments and oversized token input.

Non-goals:
- Do not add revocation, session lifecycle, or tenant policy behavior.
- Do not change generated token payload shape.
- Do not introduce a third-party JWT dependency.

Files:
- `security/jwt/jwt.go`
- `security/jwt/jwt_test.go`
- `docs/modules/security/README.md`

Tests:
- `go test -race -timeout 60s ./security/jwt`
- `go test -timeout 20s ./security/...`
- `go vet ./security/...`

Docs Sync:
- Required for clarified JWT verification input boundaries.

Done Definition:
- Malformed compact JWT envelopes fail closed before payload decode.
- Verification rejects oversized token input.
- Targeted JWT tests and vet pass.

Outcome:
- Added compact JWT envelope size limits and empty-segment rejection before decode work.
- Moved payload decoding until after signature verification.
- Added malformed and oversized compact token regression cases.
- Synced the security module primer with JWT input-boundary behavior.
- Validation run: `go test -race -timeout 60s ./security/jwt`; `go test -timeout 20s ./security/...`; `go vet ./security/...`; `go run ./internal/checks/dependency-rules`; `go run ./internal/checks/module-manifests`; `go run ./internal/checks/agent-workflow`.
