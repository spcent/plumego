# Card 0742

Milestone: Router stable readiness
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P2
State: active
Primary Module: router
Owned Files: router/dispatch.go, router/router_contract_test.go, docs/modules/router/README.md, tasks/cards/active/README.md
Depends On: 0741-router-allow-head-contract

Goal:
Make router-served `HEAD` responses bodyless even when matched by
`router.MethodAny`.

Scope:
- Define a single helper for deciding when a `HEAD` response writer should
  suppress body writes.
- Apply the same rule to warm-cache and cold-match paths.
- Add regression coverage for `HEAD` handled by `MethodAny`, including cache.
- Sync docs for the final `HEAD` behavior.

Non-goals:
- Changing request method visible to handlers.
- Changing explicit non-HEAD response behavior.
- Adding automatic HEAD route registration.

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
- `HEAD` matched through `MethodAny` suppresses response body writes.
- Cached and uncached dispatch behave the same.
- Router targeted tests, race tests, and vet pass.

Outcome:
