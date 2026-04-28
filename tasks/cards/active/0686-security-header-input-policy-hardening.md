# Card 0686

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P1
State: active
Primary Module: security
Owned Files:
- security/headers/headers.go
- security/headers/headers_test.go
- security/headers/headers_negative_matrix_test.go
- security/input/input.go
- security/input/input_test.go
- docs/modules/security/README.md
Depends On: 0685

Goal:
Make security header application and input validation fail closed for malformed proxy/header values while removing stale public documentation references.

Scope:
- Tighten HTTPS detection from proxy headers so mixed or malformed forwarded chains do not enable HSTS.
- Parse `Forwarded` header parameters exactly instead of using substring matching.
- Reject ASCII control characters in header values, not only CR/LF/NUL.
- Fix stale package documentation that references non-existent input APIs.
- Add negative tests for mixed `X-Forwarded-Proto`, malformed `Forwarded`, and unsafe header controls.

Non-goals:
- Do not add trusted-proxy configuration or app bootstrap policy to stable security.
- Do not change middleware transport wiring.
- Do not replace the basic HTML/SQL sanitizers with a non-stdlib sanitizer.

Files:
- `security/headers/headers.go`
- `security/headers/headers_test.go`
- `security/headers/headers_negative_matrix_test.go`
- `security/input/input.go`
- `security/input/input_test.go`
- `docs/modules/security/README.md`

Tests:
- `go test -race -timeout 60s ./security/headers ./security/input`
- `go test -timeout 20s ./security/...`
- `go vet ./security/...`

Docs Sync:
- Required for changed header trust semantics and corrected input package guidance.

Done Definition:
- HSTS is not emitted for mixed or malformed forwarded-proto chains.
- `Forwarded` only enables HTTPS when each element explicitly has `proto=https`.
- Header value validation rejects unsafe ASCII controls.
- Targeted security tests and vet pass.

Outcome:
