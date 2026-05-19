# Card 1228

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P2
State: done
Primary Module: x/data/sharding
Owned Files:
- x/data/sharding/logging.go
- x/data/sharding/tracing.go
- x/data/sharding/logging_test.go
- x/data/sharding/tracing_test.go
- docs/modules/x-data/README.md
Depends On:
- 0759-x-data-idempotency-kv-scope-contract

Goal:
Keep sharding observability from emitting raw SQL, DSN, or secret-bearing error text by default.

Scope:
- Sanitize error values before logging/tracing.
- Keep structured operation/table/shard attributes.
- Add tests for SQL/DSN/token-like error redaction.
- Document observability as lightweight and redacted by default.

Non-goals:
- Do not integrate a production tracing backend.
- Do not add observability dependencies.
- Do not remove existing metrics counters.

Files:
- x/data/sharding/logging.go
- x/data/sharding/tracing.go
- x/data/sharding/logging_test.go
- x/data/sharding/tracing_test.go
- docs/modules/x-data/README.md

Tests:
- go test -timeout 20s ./x/data/sharding
- go test -race -timeout 60s ./x/data/sharding
- go vet ./x/data/sharding

Docs Sync:
- Document redaction and lightweight tracing scope.

Done Definition:
- Raw SQL/DSN-like error strings are not emitted by default.
- Tests cover logging and tracing redaction.

Outcome:
- Sharding query logs now emit redacted error markers and error type instead of raw error text.
- Local trace spans now record redacted error messages and redaction markers.
- Added logging and tracing regression tests for SQL/password/token-like error text.
- Updated x/data docs for the observability redaction boundary.

Validation:
- `go test -timeout 20s ./x/data/sharding`
- `go test -race -timeout 60s ./x/data/sharding`
- `go vet ./x/data/sharding`
