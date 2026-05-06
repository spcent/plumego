# Card 0779

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P2
State: active
Primary Module: x/data/sharding
Owned Files:
- x/data/sharding/parser.go
- x/data/sharding/parser_test.go
- x/data/sharding/resolver_test.go
- docs/modules/x-data/README.md
Depends On:
- 0778-x-data-sharding-fanout-policy-contract

Goal:
Tighten the documented narrow SQL subset so unsupported clause shapes fail closed predictably.

Scope:
- Stop WHERE parsing before additional common trailing clauses such as OFFSET, FOR, LOCK, and FETCH.
- Add tests for supported single-table queries and rejected complex shapes.
- Keep the parser dependency-free and explicitly narrow.

Non-goals:
- Do not introduce a third-party SQL parser.
- Do not support JOIN, UNION, subqueries, RETURNING, or multi-statement SQL.
- Do not change sharding strategy APIs.

Files:
- x/data/sharding/parser.go
- x/data/sharding/parser_test.go
- x/data/sharding/resolver_test.go
- docs/modules/x-data/README.md

Tests:
- go test -timeout 20s ./x/data/sharding
- go test -race -timeout 60s ./x/data/sharding
- go vet ./x/data/sharding

Docs Sync:
- Update x/data docs with the stricter SQL subset boundary.

Done Definition:
- Trailing clauses no longer pollute WHERE condition extraction.
- Unsupported complex SQL continues to fail closed.
- Tests and docs cover the behavior.

Outcome:
- Extended WHERE clause termination to stop before OFFSET, FETCH, FOR, and
  LOCK IN SHARE MODE.
- Added parser and resolver regression tests for trailing clauses.
- Updated x/data docs with the narrowed SQL subset boundary.

Validation:
- `go test -timeout 20s ./x/data/sharding`
- `go test -race -timeout 60s ./x/data/sharding`
- `go vet ./x/data/sharding`
