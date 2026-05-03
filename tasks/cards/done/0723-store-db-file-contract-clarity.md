# Card 0723: Store DB File Contract Clarity

Milestone:
Recipe: specs/change-recipes/stable-root-cleanup.yaml
Priority: P1
State: done
Primary Module: store
Owned Files:
- store/db/sql.go
- store/db/sql_test.go
- store/file/file.go
- store/file/types.go
- docs/modules/store/README.md
Depends On:
- 0722

Goal:
Clarify the remaining DB and file stable contracts without widening stable store ownership.

Scope:
- Document that `db.Open` initializes a `*sql.DB` handle and does not prove connectivity; callers use `PingContext`.
- Document `QueryRow` deferred scan-error behavior.
- Tighten `file.Storage` comments for `List`, `Copy`, and `PutOptions` ownership.
- Add focused tests only where existing behavior needs a regression guard.

Non-goals:
- Do not add `OpenAndPing`, retry policy, health payloads, or timeout defaults.
- Do not add file provider config, signed URLs, metadata managers, or HTTP upload handling.
- Do not change the `Storage` method set.

Files:
- store/db/sql.go
- store/db/sql_test.go
- store/file/file.go
- store/file/types.go
- docs/modules/store/README.md

Tests:
- go test -timeout 20s ./store/db ./store/file
- go test -race -timeout 60s ./store/db ./store/file
- go vet ./store/db ./store/file

Docs Sync:
- Required for DB and file contract wording.

Done Definition:
- DB and file comments describe the stable behavior directly.
- No new stable DB policy or file provider ownership is introduced.
- Focused tests and vet pass.

Outcome:
- Clarified that `store/db.Open` and `OpenWith` initialize a `*sql.DB` handle but do not prove connectivity.
- Documented that `ApplyConfig` only applies positive values and that `QueryRow` defers query or scan errors until `Scan`.
- Tightened `store/file.Storage` comments for `Put`, `Delete`, `List`, and `Copy` without adding provider policy to the stable layer.
- Synced the store module primer on DB and file behavior boundaries.

Validation:
- go test -timeout 20s ./store/db ./store/file
- go test -race -timeout 60s ./store/db ./store/file
- go vet ./store/db ./store/file
