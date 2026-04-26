# Card 0322

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P1
State: active
Primary Module: store
Owned Files:
- store/db/sql.go
- store/db/sql_test.go
Depends On:
- 0321-store-cache-mutation-ttl-preservation

Goal:
Make `store/db.QueryContext` fail closed when an injected DB returns nil rows without an error.

Scope:
- Detect `nil, nil` results from `DB.QueryContext`.
- Return an `ErrQueryFailed`-wrapped error instead of handing nil rows to callers.
- Update context-recorder tests so context propagation is still checked explicitly.
- Add focused nil-rows coverage.

Non-goals:
- Do not change the `DB` interface.
- Do not add retry, timeout, health, metrics, or observability behavior.
- Do not change transaction semantics.

Files:
- store/db/sql.go
- store/db/sql_test.go

Tests:
- go test -timeout 20s ./store/db
- go test -race -timeout 60s ./store/db
- go vet ./store/db

Docs Sync:
- Not required.

Done Definition:
- `QueryContext` never returns nil rows with nil error.
- Existing query context propagation remains tested.
- DB targeted tests and vet pass.
