# Card 0725

Milestone:
Recipe: specs/change-recipes/symbol-change.yaml
Priority: P1
State: active
Primary Module: store
Owned Files:
- store/db/sql.go
- store/db/sql_test.go
- docs/modules/store/README.md
- docs/stable-api/snapshots/store-head.snapshot
Depends On: 0724

Goal:
Unify `store/db` row-helper error semantics so nil database inputs fail explicitly.

Scope:
- Change the exported row helper that currently returns nil on nil DB to return a typed error alongside the row.
- Update all in-repo call sites and tests.
- Preserve `QueryRow` as the convenience helper with explicit error return.
- Update docs and the store API snapshot.

Non-goals:
- Do not add retry, timeout policy, health payloads, analytics, or instrumentation.
- Do not add database driver dependencies.
- Do not change `QueryRowStrict` single-row semantics.

Files:
- `store/db/sql.go`
- `store/db/sql_test.go`
- `docs/modules/store/README.md`
- `docs/stable-api/snapshots/store-head.snapshot`

Tests:
- `go test -race -timeout 60s ./store/db`
- `go test -timeout 20s ./store/...`
- `go vet ./store/...`

Docs Sync:
- Required for row-helper error behavior.

Done Definition:
- Nil DB row-helper calls return `ErrQueryFailed` through `errors.Is`.
- All in-repo call sites compile with the new signature.
- Targeted store tests, vet, and snapshot are updated.

Outcome:
- Pending.
