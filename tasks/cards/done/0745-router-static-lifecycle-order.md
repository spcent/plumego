# Card 0745

Milestone: Router stable readiness
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P2
State: done
Primary Module: router
Owned Files: router/static.go, router/static_test.go, docs/modules/router/README.md, tasks/cards/active/README.md
Depends On: 0744-router-addroute-validation-order

Goal:
Make `Static` and `StaticFS` lifecycle error precedence match `AddRoute`.

Scope:
- Add a shared readiness/frozen preflight for static registration.
- Ensure frozen or uninitialized routers do not perform filesystem resolution
  before returning lifecycle errors.
- Keep nil filesystem validation for ready, mutable routers.
- Add regression tests for frozen static registration with missing directory
  and nil filesystem inputs.
- Sync lifecycle docs for static registration.

Non-goals:
- Changing static request serving behavior.
- Adding frontend asset policy, cache headers, SPA fallback, or ETags.
- Changing route pattern shape for static mounts.

Files:
- router/static.go
- router/static_test.go
- docs/modules/router/README.md
- tasks/cards/active/README.md

Tests:
- go test -timeout 20s ./router/...
- go test -race -timeout 60s ./router/...
- go vet ./router/...

Docs Sync:
- Required.

Done Definition:
- `Static`/`StaticFS` return lifecycle errors before filesystem/input work when
  the router is uninitialized or frozen.
- Ready mutable static input validation remains intact.
- Router targeted tests, race tests, and vet pass.

Outcome:
- Added static registration preflight so `Static` and `StaticFS` return
  uninitialized/frozen lifecycle errors before directory resolution or nil
  filesystem validation.
- Reused a single static route path helper for preflight and registration.
- Added regression tests for frozen and nil-router static registration paths.
- Documented static lifecycle precedence alongside `AddRoute`.

Validation:
- `go test -timeout 20s ./router/...`
- `go test -race -timeout 60s ./router/...`
- `go vet ./router/...`
