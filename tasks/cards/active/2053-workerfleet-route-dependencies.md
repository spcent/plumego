# Card 2053

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Context Package: implementation
Priority: P1
State: active
Primary Module: reference/workerfleet
Owned Files:
- reference/workerfleet/internal/app/app.go
- reference/workerfleet/internal/app/bootstrap_test.go
- reference/workerfleet/internal/handler/routes.go
- reference/workerfleet/internal/handler/routes_test.go
Depends On: 2052

Goal:
Replace the growing workerfleet route registrar positional parameter list with a named dependency struct.

Scope:
Introduce explicit app and handler route dependency structs, update wiring and tests, and preserve all existing route behavior.

Non-goals:
- Do not change route paths, handlers, auth semantics, or response envelopes.
- Do not introduce reflection or auto-discovery.
- Do not move route registration into `internal/app`.

Files:
- reference/workerfleet/internal/app/app.go
- reference/workerfleet/internal/app/bootstrap_test.go
- reference/workerfleet/internal/handler/routes.go
- reference/workerfleet/internal/handler/routes_test.go

Acceptance Tests:
- reference/workerfleet/internal/app/bootstrap_test.go: TestNewAppWiresCoreMiddlewareAndRoutes
- reference/workerfleet/internal/handler/routes_test.go: TestRegisterRoutesWiresHealthAndReadiness

Tests:
- Existing app and handler route tests.

Docs Sync:

Validation:
- cd reference/workerfleet && go test -timeout 30s ./internal/app ./internal/handler
- cd reference/workerfleet && go vet ./internal/app ./internal/handler
- gofmt -l reference/workerfleet/internal/app reference/workerfleet/internal/handler

Done Definition:
- [ ] Acceptance Tests pass.
- [ ] All Validation commands exit 0.
- [ ] gofmt -l . produces no output.
- [ ] Docs Sync targets updated (if applicable).

Outcome:
