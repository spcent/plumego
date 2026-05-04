# Card 0725

Milestone: Router stable readiness
Recipe: specs/change-recipes/stable-root-boundary-review.yaml
Priority: P0
State: active
Primary Module: router
Owned Files: router/router.go, router/registration.go, router/router_contract_test.go, docs/modules/router/README.md
Depends On: 0724-router-request-context-reset

Goal:
Make registered route paths canonical in storage, matching metadata, snapshots,
and reverse routing.

Scope:
- Collapse repeated leading slashes in registered route and group paths.
- Preserve rejection of internal empty path segments.
- Keep stored `Routes()` paths and request `RoutePattern` aligned.
- Update router docs for the canonical registration path contract.

Non-goals:
- Redirecting request paths.
- Normalizing internal double slashes in requests.
- Changing URL escaping behavior.

Files:
- router/router.go
- router/registration.go
- router/router_contract_test.go
- docs/modules/router/README.md

Tests:
- go test -timeout 20s ./router/...
- go test -race -timeout 60s ./router/...
- go vet ./router/...

Docs Sync:
- Required.

Done Definition:
- Leading-slash variants register as a single canonical stored path.
- Internal double-slash patterns still fail registration.
- Router tests, race tests, and vet pass.

Outcome:

