# Card 0718: Store DB QueryRow Canonicalization

Priority: P1
State: active
Primary Module: store
Owned Files:
- store/db/sql.go
- store/db/sql_test.go
- x/rest/resource_db.go

Goal:
Reduce ambiguity in `store/db` QueryRow helper behavior without adding DB policy ownership.

Scope:
- Pick one canonical single-row helper path for nil database error handling.
- Remove or de-emphasize helpers that silently return nil rows.
- Update call sites and tests in the same change if exported symbols are removed.
- Preserve caller-owned context behavior.

Non-goals:
- Do not add DB health, retry, timeout, metrics, or topology helpers.
- Do not change SQL builder behavior in `x/rest`.
- Do not change transaction behavior.

Files:
- store/db/sql.go
- store/db/sql_test.go
- x/rest/resource_db.go

Tests:
- go test -timeout 20s ./store/db ./x/rest
- go test -race -timeout 60s ./store/db ./x/rest
- go vet ./store/db ./x/rest

Docs Sync:
- Not required unless package comments mention removed helpers.

Done Definition:
- QueryRow helper behavior has one clear error-returning canonical path.
- All removed symbols have zero residual Go references.
- Focused tests and vet pass.
