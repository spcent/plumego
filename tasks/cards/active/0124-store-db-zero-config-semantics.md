# Card 0124

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P1
State: active
Primary Module: store
Owned Files:
- store/db/sql.go
- store/db/sql_test.go
Depends On:
- 0123

Goal:
Make SQL pool configuration application honor database/sql zero-value semantics.

Scope:
- Apply zero values for max idle connections and connection lifetimes where database/sql treats zero as meaningful.
- Preserve max idle clamping when max open is positive.
- Add tests for resetting non-zero settings to zero.

Non-goals:
- Do not add driver-specific behavior.
- Do not make `Open` ping by default.
- Do not add retry or health behavior.

Files:
- store/db/sql.go
- store/db/sql_test.go

Tests:
- go test -timeout 20s ./store/db
- go vet ./store/db

Docs Sync:
- Public comments only if semantics are clarified.

Done Definition:
- `ApplyConfig` can set idle/lifetime knobs back to zero.
- Existing default config behavior remains unchanged.

Outcome:

Validation:
