# Card 0762

Milestone: Router stable readiness
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P2
State: active
Primary Module: router
Owned Files: router/dispatch.go, router/router_contract_test.go, docs/modules/router/README.md, tasks/cards/active/README.md
Depends On: 0761-router-cache-key-helper-evidence

Goal:
Make incoming HTTP method `ANY` behavior match the `MethodAny` fallback
sentinel contract.

Scope:
- Avoid treating `MethodAny` as a separate exact custom method during exact
  route lookup.
- Treat incoming `ANY` like any other unusual method for fallback matching.
- Add regression coverage for incoming `ANY` requests against fallback routes.
- Sync docs for the reserved sentinel behavior.

Non-goals:
- Changing the public `MethodAny` constant.
- Adding app-level `Any` helpers.
- Changing standard method dispatch precedence.

Files:
- router/dispatch.go
- router/router_contract_test.go
- docs/modules/router/README.md
- tasks/cards/active/README.md

Tests:
- go test -timeout 20s ./router/...
- go test -race -timeout 60s ./router/...
- go vet ./router/...

Docs Sync:
- Required.

Done Definition:
- `MethodAny` is used only as fallback route storage, not an exact custom
  method branch.
- Incoming `ANY` requests can still be served by fallback routes.
- Router targeted tests, race tests, and vet pass.

Outcome:
