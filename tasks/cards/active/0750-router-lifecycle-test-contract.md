# Card 0750

Milestone: Router stable readiness
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P2
State: active
Primary Module: router
Owned Files: router/router_contract_test.go, docs/modules/router/README.md, tasks/cards/active/README.md
Depends On: 0749-router-concurrent-test-wait

Goal:
Make router concurrency tests align with the documented build-before-serve
lifecycle contract.

Scope:
- Reframe route metadata concurrency coverage so `AddRoute` calls happen in the
  registration phase only.
- Freeze before concurrent serving assertions.
- Preserve coverage for concurrent metadata registration safety and concurrent
  serving metadata visibility.
- Sync lifecycle docs only if wording needs clarification.

Non-goals:
- Supporting runtime route registration while traffic is being served.
- Changing production router locking.
- Adding new public APIs.

Files:
- router/router_contract_test.go
- docs/modules/router/README.md
- tasks/cards/active/README.md

Tests:
- go test -timeout 20s ./router/...
- go test -race -timeout 60s ./router/...
- go vet ./router/...

Docs Sync:
- Only if lifecycle wording changes.

Done Definition:
- Tests no longer imply dynamic registration during serving is supported.
- Concurrent serving is still covered after the route table is built.
- Router targeted tests, race tests, and vet pass.

Outcome:
