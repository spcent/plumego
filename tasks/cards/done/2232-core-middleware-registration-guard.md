# Card 2232

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P1
State: done
Primary Module: core
Owned Files:
- core/middleware.go
- core/app_test.go
Depends On: 2231

Goal:
Make global middleware registration fail at registration time for nil middleware instead of panicking later during handler preparation.

Scope:
- Validate every middleware passed to `App.Use`.
- Return a wrapped core error that identifies the registration operation.
- Add regression coverage proving `Prepare` remains non-panicking after nil middleware is rejected.

Non-goals:
- Do not change `middleware.Chain` semantics.
- Do not add middleware ordering features.
- Do not change route registration APIs.

Files:
- `core/middleware.go`
- `core/app_test.go`

Tests:
- `go test -timeout 20s ./core/...`
- `go vet ./core/...`

Docs Sync:
- None expected; this hardens existing registration semantics.

Done Definition:
- `App.Use(nil)` returns an error and does not mutate the middleware chain.
- Valid middleware registration behavior is unchanged.
- Core module tests pass.

Outcome:
- Added preflight nil middleware validation in `App.Use` before mutating the chain.
- Covered rejected mixed registrations and confirmed valid middleware still prepares and runs.
- Validation run: `go test -timeout 20s ./core/...`; `go vet ./core/...`.
