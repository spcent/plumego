# Card 0567

Milestone:
Recipe: specs/change-recipes/refactor.yaml
Priority: P2
State: done
Primary Module: core
Owned Files:
- core/app_helpers.go
- core/routing.go
- core/app_test.go
Depends On: 2276

Goal:
Preserve route method/path context when route registration fails because the app is uninitialized.

Scope:
- Let mutable-state checks carry optional operation parameters.
- Include method and path on uninitialized route registration errors.
- Keep middleware errors unchanged.
- Add focused regression coverage.

Non-goals:
- Do not change router matching.
- Do not change route registration success behavior.
- Do not add public APIs.

Files:
- `core/app_helpers.go`
- `core/routing.go`
- `core/app_test.go`

Tests:
- `go test -timeout 20s ./core/...`
- `go vet ./core/...`

Docs Sync:
- None expected; this improves error context only.

Done Definition:
- Zero-value route registration errors include method/path context.
- Existing core registration tests continue to pass.

Outcome:
- `ensureMutable` now accepts optional internal error params.
- Route registration passes method/path context into mutable-state failures.
- Added zero-value app regression coverage for method/path in route registration errors.
- Verified with `go test -timeout 20s ./core/...` and `go vet ./core/...`.
