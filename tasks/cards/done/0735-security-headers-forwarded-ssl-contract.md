# Card 0735

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P2
State: done
Primary Module: security
Owned Files:
- security/headers/headers.go
- security/headers/headers_test.go
- docs/modules/security/README.md

Goal:
Clarify and tighten the HTTPS detection trust model used for security headers.

Scope:
- Stop treating `X-Forwarded-Ssl: on` as sufficient HTTPS evidence in the stable primitive.
- Keep direct TLS, `X-Forwarded-Proto`, and RFC `Forwarded` proto handling.
- Add tests proving `X-Forwarded-Ssl` alone does not enable HSTS.
- Sync docs with the supported forwarded HTTPS signals.

Non-goals:
- Do not add trusted proxy configuration.
- Do not add middleware-specific proxy policy.
- Do not change CSP, COOP, CORP, or other header policy behavior.

Files:
- security/headers/headers.go
- security/headers/headers_test.go
- docs/modules/security/README.md

Tests:
- go test -timeout 20s ./security/headers
- go vet ./security/headers
- go run ./internal/checks/dependency-rules

Docs Sync:
- Update security module docs with the HTTPS detection contract.

Done Definition:
- HSTS is not emitted solely because of `X-Forwarded-Ssl: on`.
- Existing supported forwarded proto tests still pass.
- Targeted tests, vet, and dependency checks pass.

Outcome:
- Removed `X-Forwarded-Ssl: on` from the stable HTTPS detection primitive.
- Kept direct TLS, all-HTTPS `X-Forwarded-Proto`, and all-HTTPS RFC `Forwarded` proto support.
- Added HSTS coverage proving `X-Forwarded-Ssl` alone is ignored.
- Updated security module docs with the supported HTTPS detection signals.

Validation:
- `gofmt -w security/headers/headers.go security/headers/headers_test.go security/headers/headers_negative_matrix_test.go`
- `go test -timeout 20s ./security/headers`
- `go vet ./security/headers`
- `go run ./internal/checks/dependency-rules`
