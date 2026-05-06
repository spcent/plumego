# Card 0773

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P1
State: done
Primary Module: x/data/idempotency
Owned Files: x/data/idempotency/sql.go, x/data/idempotency/sql_test.go
Depends On:

Goal:

Make SQL idempotency duplicate-key detection conservative enough that unrelated constraint errors are not swallowed as inserted=false,nil.

Scope:

- Prefer driver/dialect-style duplicate codes when available.
- Keep custom SQLConfig.IsDuplicateError as the authoritative override.
- Restrict string fallback to duplicate-key phrases, not generic unique/constraint wording.
- Add regression coverage for unrelated unique/constraint errors.

Non-goals:

- Importing vendor-specific SQL drivers.
- Changing the stable store/idempotency contract.
- Changing expired reclaim behavior.

Files:

- x/data/idempotency/sql.go
- x/data/idempotency/sql_test.go

Tests:

- go test -race -timeout 60s ./x/data/idempotency
- go test -timeout 20s ./x/data/idempotency
- go vet ./x/data/idempotency

Docs Sync:

- Not required; this preserves existing provider configuration and tightens default behavior.

Done Definition:

- Generic constraint/unique errors are returned, not converted to duplicate occupancy.
- Recognized duplicate-key errors still produce inserted=false,nil.
- Module tests and vet pass.

Outcome:

- Default duplicate detection now recognizes PostgreSQL SQLSTATE 23505, MySQL error number 1062, and explicit duplicate-key phrases.
- Generic unique/constraint wording is no longer swallowed as duplicate occupancy.
- Updated tests to lock the conservative classifier behavior.
- Validation passed:
  - go test -race -timeout 60s ./x/data/idempotency
  - go test -timeout 20s ./x/data/idempotency
  - go vet ./x/data/idempotency
