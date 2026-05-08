# Card 0716

Milestone: —
Recipe: specs/change-recipes/http-endpoint-bugfix.yaml
Priority: P1
State: done
Primary Module: x/gateway
Owned Files:
- `x/gateway/entrypoints.go`
- `x/gateway/entrypoints_test.go`
- `docs/modules/x-gateway/README.md`
Depends On: —

Goal:
- Make gateway route registration helpers fail visibly on invalid inputs.

Problem:
`gateway.RegisterRoute` returns nil when router, handler, or path is invalid, and `RegisterProxy` returns `(nil, nil)` when router or path is invalid. This hides app wiring mistakes and differs from `core` and `router` registration behavior.

Scope:
- Return explicit errors for nil router, nil handler, and empty path.
- Preserve successful `ANY` route registration behavior.
- Update tests for invalid inputs and successful registration.
- Document the helper behavior if docs currently imply silent no-op registration.

Non-goals:
- Do not change `router.Router.AddRoute`.
- Do not change gateway proxy configuration validation.
- Do not introduce new route helper aliases.

Files:
- `x/gateway/entrypoints.go`
- `x/gateway/entrypoints_test.go`
- `docs/modules/x-gateway/README.md`

Tests:
- `go test -timeout 20s ./x/gateway/...`
- `go vet ./x/gateway/...`
- `go run ./internal/checks/dependency-rules`

Docs Sync:
- Required if README mentions `RegisterRoute` or `RegisterProxy`.

Done Definition:
- Invalid gateway registration inputs return actionable errors.
- Existing valid registration tests still pass.
- No stable root imports `x/gateway`.

Outcome:
- `RegisterRoute` now returns explicit errors for nil router, blank path, and nil handler instead of silently succeeding.
- `RegisterProxy` now returns explicit errors for nil router and blank path before constructing a proxy.
- Updated gateway entrypoint tests and module docs to reflect invalid registration errors.
- Validation:
  - `go test -timeout 20s ./x/gateway/...`
  - `go vet ./x/gateway/...`
  - `go run ./internal/checks/dependency-rules`
