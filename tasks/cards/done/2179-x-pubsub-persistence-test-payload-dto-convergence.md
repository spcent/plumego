# Card 2179: x/pubsub Persistence Test Payload Fixture Convergence

Milestone: none
Recipe: specs/change-recipes/module-cleanup.yaml
Priority: P1
State: active
Primary Module: x/pubsub
Owned Files:
- `x/pubsub/persistence_test.go`
Depends On: none

Goal:
Make persistent pubsub tests avoid generic map payload fixtures where scalar
test data is sufficient.

Problem:
`x/pubsub/persistence_test.go` repeatedly builds message `Data` with
`map[string]any` for values that are only counted, persisted, or delivered.
These tests do not need structured payload fields, and the maps add unnecessary
serialization surface to persistence tests.

Scope:
- Replace fixed persistence-test payload maps with scalar fixtures.

Non-goals:
- Do not change pubsub persistence behavior or public APIs.
- Do not change callback, WAL, snapshot, or restore semantics.
- Do not add dependencies.

Files:
- `x/pubsub/persistence_test.go`

Tests:
- `go test -race -timeout 60s ./x/pubsub/...`
- `go test -timeout 20s ./x/pubsub/...`
- `go vet ./x/pubsub/...`

Docs Sync:
No docs change required; this is test fixture cleanup.

Done Definition:
- Persistence tests no longer use generic maps where scalar message payloads are
  sufficient.
- The listed validation commands pass.

Outcome:
