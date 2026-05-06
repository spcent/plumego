# Card 0766

Milestone: Router stable readiness
Recipe: specs/change-recipes/fix-test.yaml
Priority: P3
State: active
Primary Module: router
Owned Files: router/router_contract_test.go, docs/modules/router/README.md, tasks/cards/active/README.md
Depends On: 0765-router-print-snapshot-before-write

Goal:
Make cache key tests describe the actual path-based router contract.

Scope:
- Rename the host/method cache test so it no longer claims host separation.
- Assert the cache key behavior that actually exists: method plus normalized
  path.
- Document that router matching and match-cache keys are host-agnostic.

Non-goals:
- Adding host-based routing.
- Changing cache key format beyond current method/path behavior.
- Changing dispatch behavior.

Files:
- router/router_contract_test.go
- docs/modules/router/README.md
- tasks/cards/active/README.md

Tests:
- go test -timeout 20s ./router/...
- go vet ./router/...

Docs Sync:
- Required.

Done Definition:
- Tests no longer imply host-based cache separation.
- Docs describe host-agnostic matching/cache behavior.
- Router targeted tests and vet pass.

Outcome:
