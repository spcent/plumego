# Card 0862

Milestone: Router stable readiness
Recipe: specs/change-recipes/stable-root-boundary-review.yaml
Priority: P2
State: done
Primary Module: router
Owned Files: router/router_contract_test.go, docs/modules/router/README.md, router/module.yaml
Depends On: 0728-router-static-url-path-hardening

Goal:
Make the router 404 and 405 response contract explicit and regression-tested.

Scope:
- Keep existing behavior unless tests reveal drift.
- Document plain stdlib 404 and structured contract 405 behavior.
- Add or tighten tests that lock both response shapes and headers.

Non-goals:
- Changing response envelopes.
- Adding router-owned JSON response helpers.
- Changing middleware or core error handling.

Files:
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
- Docs and tests explicitly describe 404 versus 405 response behavior.
- Router tests, race tests, and vet pass.

Outcome:
- Locked existing router miss behavior as standard-library plain-text 404.
- Locked enabled method-mismatch behavior as structured `contract.WriteError`
  405 with the sorted `Allow` header.
- Updated router docs and manifest review guidance for the 404/405 response
  contract.

Validation:
- `go test -timeout 20s ./router/...`
- `go test -race -timeout 60s ./router/...`
- `go vet ./router/...`
- Extra: `go run ./internal/checks/module-manifests`
