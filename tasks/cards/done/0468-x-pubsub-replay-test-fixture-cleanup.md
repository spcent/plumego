# Card 0468: x/pubsub Replay Test Fixture Cleanup

Milestone: none
Recipe: specs/change-recipes/module-cleanup.yaml
Priority: P2
State: active
Primary Module: x/pubsub
Owned Files:
- `x/pubsub/replay_test.go`
Depends On: none

Goal:
Clean up replay store test fixtures by removing duplicated wait setup and using
a typed message payload for fixed replay data.

Problem:
`TestReplayStore_Basic` has duplicated readiness sleep comments/calls and uses a
generic map for a fixed `seq` test payload. The test intent is simple but the
setup is slightly noisy and inconsistent with the typed payload style now used
in ordering tests.

Scope:
- Remove the duplicate readiness wait block.
- Add a local typed replay payload fixture for the fixed `seq` message data.

Non-goals:
- Do not change replay store behavior or timing semantics beyond removing the
  accidental duplicate sleep.
- Do not change scalar message payload tests.
- Do not add dependencies.

Files:
- `x/pubsub/replay_test.go`

Tests:
- `go test -race -timeout 60s ./x/pubsub/...`
- `go test -timeout 20s ./x/pubsub/...`
- `go vet ./x/pubsub/...`

Docs Sync:
No docs change required; this is test cleanup.

Done Definition:
- The duplicated readiness wait is removed.
- The fixed replay `seq` payload no longer uses a generic map.
- The listed validation commands pass.

Outcome:
