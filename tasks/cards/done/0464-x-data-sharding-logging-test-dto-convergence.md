# Card 0464: x/data/sharding Logging Test DTO Convergence

Milestone: none
Recipe: specs/change-recipes/module-cleanup.yaml
Priority: P1
State: active
Primary Module: x/data/sharding
Owned Files:
- `x/data/sharding/logging_test.go`
Depends On: none

Goal:
Consolidate sharding logging tests around a typed JSON log-entry DTO.

Problem:
`x/data/sharding/logging_test.go` repeatedly decodes fixed JSON log entries
into `map[string]any` and manually casts numeric fields. The tests are asserting
stable logging fields, so the current shape is repetitive and less explicit
than a typed fixture.

Scope:
- Add a local test DTO and decode helper for sharding log entries.
- Replace repeated fixed-field map decodes in logging tests.
- Preserve the current request ID, query, shard, policy, rewrite, and component
  assertions.

Non-goals:
- Do not change sharding router or logging behavior.
- Do not alter sharding module imports outside the existing test file.
- Do not add dependencies.

Files:
- `x/data/sharding/logging_test.go`

Tests:
- `go test -race -timeout 60s ./x/data/sharding/...`
- `go test -timeout 20s ./x/data/sharding/...`
- `go vet ./x/data/sharding/...`

Docs Sync:
No docs change required; this is test-only cleanup.

Done Definition:
- Sharding logging tests no longer repeat fixed-field `map[string]any` decodes.
- The listed validation commands pass.

Outcome:
