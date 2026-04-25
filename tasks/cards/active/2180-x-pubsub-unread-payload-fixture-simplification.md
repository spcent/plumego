# Card 2180: x/pubsub Unread Payload Fixture Simplification

Milestone: none
Recipe: specs/change-recipes/module-cleanup.yaml
Priority: P2
State: active
Primary Module: x/pubsub
Owned Files:
- `x/pubsub/pubsub_basic_test.go`
- `x/pubsub/distributed_test.go`
- `x/pubsub/consumergroup_test.go`
Depends On: none

Goal:
Simplify pubsub tests that attach structured map payloads which the test never
reads.

Problem:
Several pubsub tests create `map[string]any` payloads only to verify delivery,
distribution, or counting. The structured fields are not asserted, so they add
unnecessary dynamic payload surface and distract from the test intent.

Scope:
- Replace unread fixed map payloads with scalar fixtures.
- Leave `TestPubSub_MessageImmutability` maps intact because that test
  intentionally verifies map cloning/isolation.

Non-goals:
- Do not change pubsub runtime behavior or public APIs.
- Do not change tests that assert structured payload mutation semantics.
- Do not add dependencies.

Files:
- `x/pubsub/pubsub_basic_test.go`
- `x/pubsub/distributed_test.go`
- `x/pubsub/consumergroup_test.go`

Tests:
- `go test -race -timeout 60s ./x/pubsub/...`
- `go test -timeout 20s ./x/pubsub/...`
- `go vet ./x/pubsub/...`

Docs Sync:
No docs change required; this is test fixture cleanup.

Done Definition:
- Pubsub tests no longer use structured maps where payload content is not read.
- The listed validation commands pass.

Outcome:
