# Card 0309

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P1
State: active
Primary Module: store
Owned Files:
- store/db/sql.go
- store/db/sql_test.go
Depends On:
- 0308-store-kv-capacity-durable-persist

Goal:
Make `store/db` helper error behavior safer and more inspectable without adding DB policy ownership.

Scope:
- Guard `WithTransaction` against a nil transaction function.
- Guard `ScanRows` against a nil scan function.
- Wrap underlying begin, transaction callback, commit, scan, rows, ping, open, and query errors with `%w` where a caller may need `errors.Is`.
- Add focused tests proving sentinel and underlying errors remain discoverable.

Non-goals:
- Do not add timeout policy, retry loops, health payloads, metrics, analytics, or pool-stat ownership.
- Do not change the `DB` interface.
- Do not add non-stdlib dependencies.

Files:
- store/db/sql.go
- store/db/sql_test.go

Tests:
- go test -timeout 20s ./store/db
- go test -race -timeout 60s ./store/db
- go vet ./store/db

Docs Sync:
- Not required; this preserves existing helper ownership.

Done Definition:
- Nil callbacks return stable sentinel errors instead of panicking.
- Wrapped DB helper errors remain matched by both Plumego sentinels and relevant underlying errors.
- Existing context ownership tests still pass.

Outcome:
