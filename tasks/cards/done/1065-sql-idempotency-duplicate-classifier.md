# Card 1065

Milestone:
Recipe: specs/change-recipes/store-stable.yaml
Priority: P1
State: done
Primary Module: x/data/idempotency
Owned Files:
- x/data/idempotency/sql.go
- x/data/idempotency/sql_test.go
Depends On:

Goal:
Make SQL idempotency duplicate-key detection conservative and extensible.

Scope:
- Stop treating every constraint-like SQL error as a duplicate key.
- Preserve duplicate-key handling for common unique/duplicate messages.
- Add a provider hook if needed for driver-specific duplicate detection.
- Add regression coverage for non-duplicate constraint errors.

Non-goals:
- Do not add database driver dependencies.
- Do not change the stable store/idempotency interface.

Files:
- x/data/idempotency/sql.go
- x/data/idempotency/sql_test.go

Tests:
- go test -timeout 20s ./x/data/idempotency
- go test -timeout 20s ./store/idempotency

Docs Sync:
- None unless public provider configuration docs require an added hook.

Done Definition:
- Non-duplicate constraint failures are returned to callers.
- Duplicate-key handling remains covered.
- Targeted tests pass.

Outcome:
SQL duplicate classification is now conservative by default and no longer treats every constraint-like error as duplicate. SQLConfig accepts an optional IsDuplicateError hook for driver-specific classification. Added tests for custom classifier behavior and non-duplicate constraint messages.

Validation:
- go test -timeout 20s ./x/data/idempotency
- go test -timeout 20s ./store/idempotency
