# Card 0897: SQL Idempotency Identifier Validation

Milestone:
Recipe: specs/change-recipes/stable-root-cleanup.yaml
Priority: P1
State: done
Primary Module: x/data/idempotency
Owned Files:
- x/data/idempotency/store.go
- x/data/idempotency/sql.go
- x/data/idempotency/sql_test.go
- docs/modules/store/README.md
Depends On:
- 0731

Goal:
Close SQL identifier injection and misconfiguration risks in the idempotency SQL provider.

Scope:
- Validate `SQLConfig.Table` before interpolating it into queries.
- Validate supported dialect values.
- Return stable package sentinel errors for invalid provider configuration.
- Add tests for invalid table and dialect values.

Non-goals:
- Do not introduce a SQL builder dependency.
- Do not add runtime schema discovery or migrations.
- Do not change table naming policy beyond identifier validation.

Files:
- x/data/idempotency/store.go
- x/data/idempotency/sql.go
- x/data/idempotency/sql_test.go
- docs/modules/store/README.md

Tests:
- go test -timeout 20s ./x/data/idempotency ./store/idempotency
- go vet ./x/data/idempotency ./store/idempotency
- go run ./internal/checks/dependency-rules

Docs Sync:
- Required if provider configuration rules become explicit in store docs.

Done Definition:
- Unsafe table identifiers cannot reach query construction.
- Unsupported dialects fail predictably.
- Targeted tests, vet, and dependency boundary checks pass.

Outcome:
- Added provider-level `ErrInvalidConfig` for invalid durable idempotency
  configuration.
- Defaulted empty SQL dialect to Postgres and validated table identifiers and
  supported dialects before query construction.
- Allowed schema-qualified table identifiers such as
  `public.idempotency_keys` while rejecting unsafe identifier text.
- Documented SQL-backed idempotency identifier validation in the store module
  docs.

Validation:
- `go test -timeout 20s ./x/data/idempotency ./store/idempotency`
- `go vet ./x/data/idempotency ./store/idempotency`
- `go run ./internal/checks/dependency-rules`
