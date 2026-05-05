# Card 0733

Milestone:
Recipe: specs/change-recipes/refactor.yaml
Priority: P2
State: active
Primary Module: security
Owned Files:
- security/jwt/jwt.go
- security/jwt/policy.go
- security/jwt/context.go
- security/jwt/jwt_test.go
- docs/modules/security/README.md
Depends On:
- 0732

Goal:
Reduce JWT maintenance risk by isolating policy/context helpers and making time-dependent behavior easier to test.

Scope:
- Move JWT authorization policy helpers into a focused file without changing exported APIs.
- Move token-claims context helpers into a focused file without changing exported APIs.
- Add an internal clock hook for manager time-dependent logic and tests for rotation/verification boundaries.

Non-goals:
- Do not change JWT wire format.
- Do not add public clock configuration unless needed for stable API.
- Do not change middleware adapters.

Files:
- security/jwt/jwt.go
- security/jwt/policy.go
- security/jwt/context.go
- security/jwt/jwt_test.go
- docs/modules/security/README.md

Tests:
- go test -timeout 20s ./security/jwt
- go vet ./security/jwt

Docs Sync:
- Update docs only if the time behavior is clarified.

Done Definition:
- `security/jwt/jwt.go` is smaller and focused on manager/signing/verification.
- Policy and context helpers keep existing API and tests.
- Time-dependent JWT paths have deterministic focused tests.

Outcome:

Validation:
