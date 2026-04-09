# Card 0857

Priority: P1
State: active
Primary Module: store
Owned Files:
- `store/db/sql.go`
- `store/db/sql_test.go`
- `store/module.yaml`
- `docs/modules/store/README.md`

Goal:
- Remove ops-style health and retry-policy convenience from stable `store/db`.
- Keep stable `store/db` limited to stdlib-shaped open/config/query/transaction helpers.

Problem:
- Stable `store/db` still exports `HealthCheck(...)`, a bespoke `HealthStatus` payload, `OpenWithRetry(...)`, and `WithTimeout(...)`.
- `HealthCheck(...)` duplicates health/ops ownership, while `OpenWithRetry(...)` and `WithTimeout(...)` encode caller policy rather than a minimal DB primitive.
- Repository grep currently shows these helpers have no non-test callers, and the package comment still advertises sharding, rw-split, and metrics capabilities that no longer belong to stable `store/db`.

Scope:
- Remove health, retry-policy, and generic timeout convenience helpers from stable `store/db`.
- Tighten package comments, docs, and manifest language to the reduced stable boundary.
- Keep the remaining stdlib-compatible open/config/query/transaction helpers intact.

Non-goals:
- Do not redesign `QueryRowStrict`, `ScanRows`, or transaction helpers.
- Do not reintroduce DB health/reporting wrappers elsewhere in stable `store`.
- Do not add compatibility wrappers for the removed helpers.

Files:
- `store/db/sql.go`
- `store/db/sql_test.go`
- `store/module.yaml`
- `docs/modules/store/README.md`

Tests:
- `go test -timeout 20s ./store/db`
- `go test -race -timeout 60s ./store/db`
- `go vet ./store/db`

Docs Sync:
- Keep store docs aligned on the rule that stable `store/db` is a small DB helper layer, not an ops/health or topology package.

Done Definition:
- Stable `store/db` no longer exports DB health payloads or retry/timeout policy helpers.
- The package comment, module manifest, and code all describe the same narrowed stable boundary.
- There are zero residual runtime callers of the removed helpers.

Outcome:
- Completed.
- Removed `OpenWithRetry(...)`, `WithTimeout(...)`, `HealthCheck(...)`, and `HealthStatus` from stable `store/db`, leaving only stdlib-shaped open, ping, query, scan, and transaction helpers.
- Tightened the package comment so stable `store/db` no longer advertises retry, health, or topology ownership.
- Updated store docs and manifest language to keep DB health/reporting payloads and retry policy explicitly out of the stable store boundary.
