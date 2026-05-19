# Card 0912

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P0
State: done
Primary Module: x/data/idempotency
Owned Files:
- x/data/idempotency/kv.go
- x/data/idempotency/sql.go
- x/data/idempotency/kv_test.go
- x/data/idempotency/sql_test.go
Depends On:
- 0732-x-data-rw-fail-closed-routing

Goal:
Make idempotency storage behavior deterministic enough for stable use.

Scope:
- Make KV `PutIfAbsent` atomic for concurrent callers.
- Validate SQL table identifiers before interpolating them into queries.
- Make zero-value SQL config match documented defaults.
- Add regression tests for concurrent KV insertion and invalid SQL table names.

Non-goals:
- Do not introduce external SQL driver dependencies.
- Do not implement cross-process KV locking.
- Do not change the public `Record` model.

Files:
- x/data/idempotency/kv.go
- x/data/idempotency/sql.go
- x/data/idempotency/kv_test.go
- x/data/idempotency/sql_test.go

Tests:
- go test -timeout 20s ./x/data/idempotency
- go test -race -timeout 60s ./x/data/idempotency
- go vet ./x/data/idempotency

Docs Sync:
- Not required unless public config behavior changes beyond matching existing defaults.

Done Definition:
- Concurrent same-key KV inserts produce exactly one successful insert in-process.
- Invalid SQL table identifiers are rejected before query construction.
- idempotency normal, race, and vet checks pass.

Outcome:
- Serialized KV `PutIfAbsent` in the x/data wrapper so concurrent same-key claims are atomic within a process.
- Defaulted zero-value SQL config to PostgreSQL placeholders.
- Added SQL table identifier validation before query construction.
- Added regression coverage for concurrent KV claims, default dialect, and invalid SQL table names.

Validation:
- GOCACHE=/private/tmp/plumego-go-build go test -timeout 20s ./x/data/idempotency
- GOCACHE=/private/tmp/plumego-go-build go test -race -timeout 60s ./x/data/idempotency
- GOCACHE=/private/tmp/plumego-go-build go vet ./x/data/idempotency
