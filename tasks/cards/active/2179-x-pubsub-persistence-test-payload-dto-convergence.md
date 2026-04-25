# Card 2179: x/pubsub Persistence Test Payload DTO Convergence

Milestone: none
Recipe: specs/change-recipes/module-cleanup.yaml
Priority: P1
State: active
Primary Module: x/pubsub
Owned Files:
- `x/pubsub/persistence_test.go`
Depends On: none

Goal:
Make persistent pubsub tests use typed message payload fixtures for fixed test
fields.

Problem:
`x/pubsub/persistence_test.go` repeatedly builds message `Data` with
`map[string]any` for fixed fields such as `index`, `text`, `n`, `goroutine`,
`seq`, and `i`. These are test fixtures, not dynamic payload requirements, and
the map literals are inconsistent with the typed ordering/replay fixture style.

Scope:
- Add a local persistence message payload struct.
- Replace fixed persistence-test payload maps with typed fixtures.

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
- Persistence tests no longer use generic maps for fixed message payload fields.
- The listed validation commands pass.

Outcome:
