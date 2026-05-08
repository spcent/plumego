# Card 0909

Milestone: Router stable readiness
Recipe: specs/change-recipes/stable-root-boundary-review.yaml
Priority: P0
State: done
Primary Module: router
Owned Files: router/router.go, router/router_contract_test.go, docs/modules/router/README.md, router/module.yaml
Depends On: 0732-router-http-method-token-validation

Goal:
Make `Freeze` consistently lock runtime router policy in addition to route
registration.

Scope:
- Prevent `SetMethodNotAllowed` from mutating policy after `Freeze`.
- Keep `WithMethodNotAllowed` construction behavior unchanged.
- Add tests for policy mutation before and after freeze.
- Update lifecycle docs and manifest risk language.

Non-goals:
- Adding a new error-returning setter API.
- Changing core-owned router preparation behavior.
- Changing default 405 behavior.

Files:
- router/router.go
- router/router_contract_test.go
- docs/modules/router/README.md
- router/module.yaml

Tests:
- go test -timeout 20s ./router/...
- go test -race -timeout 60s ./router/...
- go vet ./router/...

Docs Sync:
- Required.

Done Definition:
- Frozen routers ignore later method-not-allowed policy toggles.
- Router docs state that Freeze locks route table and runtime policy.
- Router tests, race tests, and vet pass.

Outcome:
- Made `SetMethodNotAllowed` ignore runtime policy changes after `Freeze`.
- Added tests for disabling and enabling method-not-allowed policy after
  freeze.
- Updated router lifecycle docs and manifest risk language to cover frozen
  runtime policy.

Validation:
- `go test -timeout 20s ./router/...`
- `go test -race -timeout 60s ./router/...`
- `go vet ./router/...`
- Extra: `go run ./internal/checks/module-manifests`
