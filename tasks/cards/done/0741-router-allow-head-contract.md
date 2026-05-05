# Card 0741

Milestone: Router stable readiness
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P1
State: done
Primary Module: router
Owned Files: router/dispatch.go, router/router_contract_test.go, router/freeze_test.go, docs/modules/router/README.md, tasks/cards/active/README.md
Depends On: 0740-router-option-freeze-guard

Goal:
Make 405 `Allow` responses include implicit `HEAD` when a matching `GET`
route exists.

Scope:
- Update `allowedMethods` so a path served by `GET` advertises both `GET` and
  `HEAD` unless an explicit `HEAD` route already exists.
- Preserve sorted, deterministic `Allow` header output.
- Add root and non-root regression coverage.
- Sync router docs for the 405/HEAD stable contract.

Non-goals:
- Changing 404 response shape.
- Changing `MethodAny` fallback policy.
- Adding implicit `OPTIONS`.

Files:
- router/dispatch.go
- router/router_contract_test.go
- router/freeze_test.go
- docs/modules/router/README.md
- tasks/cards/active/README.md

Tests:
- go test -timeout 20s ./router/...
- go test -race -timeout 60s ./router/...
- go vet ./router/...

Docs Sync:
- Required.

Done Definition:
- `Allow` includes implicit `HEAD` for matching `GET` routes.
- Duplicate `HEAD` is not emitted when explicit `HEAD` is registered.
- Router targeted tests, race tests, and vet pass.

Outcome:
- `allowedMethods` now appends implicit `HEAD` when a matching `GET` route
  exists and avoids duplicate `HEAD` when an explicit `HEAD` route is
  registered.
- Updated 405 response tests for root, non-root, sorted multi-method, cached,
  and structured error paths.
- Documented that sorted `Allow` includes implicit `HEAD` for matching `GET`
  routes.

Validation:
- `go test -timeout 20s ./router/...`
- `go test -race -timeout 60s ./router/...`
- `go vet ./router/...`
