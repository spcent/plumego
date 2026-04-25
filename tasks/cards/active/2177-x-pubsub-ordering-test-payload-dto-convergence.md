# Card 2177: x/pubsub Ordering Test Payload DTO Convergence

Milestone: none
Recipe: specs/change-recipes/module-cleanup.yaml
Priority: P1
State: active
Primary Module: x/pubsub
Owned Files:
- `x/pubsub/ordering_test.go`
Depends On: none

Goal:
Make ordered pubsub tests use a typed message payload fixture for fixed
`seq/key/worker` test data.

Problem:
`x/pubsub/ordering_test.go` repeatedly creates message data as
`map[string]any` and then type-asserts the same fixed fields. This is noisy and
less type-safe than a local test DTO for ordering payloads.

Scope:
- Add a local ordering message payload struct.
- Replace fixed `seq`, `key`, and `worker` payload maps in ordering tests.
- Preserve ordering semantics and all existing assertions.

Non-goals:
- Do not change pubsub message APIs or runtime behavior.
- Do not change tests that intentionally use scalar message data.
- Do not add dependencies.

Files:
- `x/pubsub/ordering_test.go`

Tests:
- `go test -race -timeout 60s ./x/pubsub/...`
- `go test -timeout 20s ./x/pubsub/...`
- `go vet ./x/pubsub/...`

Docs Sync:
No docs change required; this is test fixture cleanup.

Done Definition:
- Ordering tests no longer use generic maps for fixed ordering payload fields.
- The listed validation commands pass.

Outcome:
