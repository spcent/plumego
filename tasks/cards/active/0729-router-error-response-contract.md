# Card 0729

Milestone: Router stable readiness
Recipe: specs/change-recipes/stable-root-boundary-review.yaml
Priority: P2
State: active
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

