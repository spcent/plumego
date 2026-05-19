# Card 0123

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P0
State: done
Primary Module: security
Owned Files:
- security/headers/headers.go
- security/headers/headers_test.go
- security/headers/headers_negative_matrix_test.go
- middleware/security/headers_test.go
Depends On:
- 0118

Goal:
Make security header policy validation reject semantically invalid standard header values instead of only checking header syntax.

Scope:
- Validate `X-Frame-Options`, `X-Content-Type-Options`, Referrer-Policy, COOP, CORP, and COEP against supported values.
- Keep custom `Additional` headers syntax-validated.
- Remove stale CSP nonce-generation claim if no nonce primitive exists.
- Add negative-path tests through both primitive validation and middleware fail-closed behavior.

Non-goals:
- Do not implement CSP parsing.
- Do not add nonce generation in this card.
- Do not change middleware response envelope shape.

Files:
- security/headers/headers.go
- security/headers/headers_test.go
- security/headers/headers_negative_matrix_test.go
- middleware/security/headers_test.go

Tests:
- go test -timeout 20s ./security/headers ./middleware/security
- go vet ./security/headers ./middleware/security

Docs Sync:
- Update package comments only if removing stale claims.

Done Definition:
- Invalid known header enum values fail `Policy.Validate`.
- Middleware keeps failing closed on invalid policies.
- Header syntax checks remain in place.

Outcome:
- Added semantic validation for `X-Frame-Options`, `X-Content-Type-Options`, Referrer-Policy, COOP, CORP, and COEP values.
- Kept CSP and Permissions-Policy on syntax validation only, so this card does not expand into policy parsing.
- Removed the stale package-level claim that the headers package generates CSP nonces.
- Added primitive and middleware tests proving invalid semantic policies fail closed.

Validation:
- go test -timeout 20s ./security/headers ./middleware/security
- go vet ./security/headers ./middleware/security
