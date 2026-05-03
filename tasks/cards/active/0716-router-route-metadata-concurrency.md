# Card 0716

Milestone: Router stable readiness
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P1
State: active
Primary Module: router
Owned Files: router/router.go, router/metadata.go, router/dispatch.go, router/router_contract_test.go
Depends On: 0715-router-registration-validation-and-head

Goal:
Make route metadata reads use the same synchronization discipline as route registration.

Scope:
- Split locked and public metadata lookup paths so request dispatch reads routeMeta safely.
- Keep existing metadata behavior, route names, and Routes snapshots unchanged.
- Add race-oriented coverage for serving while named routes are registered.

Non-goals:
- Changing the named-route conflict policy.
- Adding new public metadata APIs.
- Moving metadata ownership out of router.

Files:
- router/router.go
- router/metadata.go
- router/dispatch.go
- router/router_contract_test.go

Tests:
- go test -timeout 20s ./router/...
- go test -race -timeout 60s ./router/...
- go vet ./router/...

Docs Sync:
- Not required unless the lifecycle contract changes.

Done Definition:
- Metadata lookup has no unlocked map reads in request dispatch.
- Existing route pattern/name behavior remains intact.
- Router tests and vet pass.

Outcome:

