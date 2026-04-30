# Card 0711

Milestone: M-002
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P0
State: done
Primary Module: security
Owned Files: security/authn/token.go, security/jwt/jwt.go, security/headers/headers.go, security/input/input.go, security/**/*_test.go
Depends On: 0706-stable-root-api-inventory

Goal:
Strengthen security negative-path coverage for fail-closed auth/token/header/input behavior.

Scope:
Add or tighten tests around invalid tokens, unsafe request/header inputs, security header policy edge cases, and timing-safe secret checks using existing APIs.

Non-goals:
Do not introduce tenant session lifecycle behavior.
Do not add provider globals or hidden registration.
Do not weaken validation or logging safety.

Files:
security/authn/token.go
security/jwt/jwt.go
security/headers/headers.go
security/input/input.go
security/**/*_test.go

Tests:
go test -race -timeout 60s ./security/...
go test -timeout 20s ./security/...
go run ./internal/checks/dependency-rules

Docs Sync:
None unless behavior changes. Security semantics changes must be documented before this card is marked done.

Done Definition:
Invalid-token and unsafe-input paths fail closed.
Negative-path tests cover the touched security subpackage.
Security tests and dependency boundary check pass.

Outcome:
Completed.

Changes:

- Fixed `StaticToken.Authenticate(nil)` to fail closed with `ErrUnauthenticated` instead of panicking during `X-Token` fallback extraction.
- Added a nil-request regression test for static token authentication.

Validation:

- `go test -race -timeout 60s ./security/...` passed.
- `go test -timeout 20s ./security/...` passed.
- `go run ./internal/checks/dependency-rules` passed.

