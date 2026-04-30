# Card 0717

Milestone: —
Recipe: specs/change-recipes/http-endpoint-bugfix.yaml
Priority: P1
State: active
Primary Module: x/rest
Owned Files:
- `x/rest/entrypoints.go`
- `x/rest/routes_test.go`
- `docs/modules/x-rest/README.md`
Depends On: —

Goal:
- Make REST resource route registration fail visibly on invalid inputs.

Problem:
`rest.RegisterResourceRoutes` silently returns nil when router or controller is nil. That makes broken app wiring look successful and conflicts with the repo rule that route registration errors must be returned to the caller.

Scope:
- Return explicit errors for nil router and nil controller.
- Preserve existing route order and successful registration surface.
- Add tests for nil router, nil controller, empty prefix normalization, and successful default route registration.
- Sync docs if they describe registration behavior.

Non-goals:
- Do not change `ResourceController`.
- Do not redesign REST CRUD route conventions.
- Do not change router matching behavior.

Files:
- `x/rest/entrypoints.go`
- `x/rest/routes_test.go`
- `docs/modules/x-rest/README.md`

Tests:
- `go test -timeout 20s ./x/rest/...`
- `go vet ./x/rest/...`
- `go run ./internal/checks/dependency-rules`

Docs Sync:
- Required if README examples imply invalid inputs are ignored.

Done Definition:
- Invalid REST route registration inputs return actionable errors.
- Existing default route registration behavior remains unchanged for valid inputs.
- Tests cover both invalid and valid paths.

Outcome:
-
