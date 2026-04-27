# Card 0435: x/pubsub Distributed Cluster HTTP Contract

Milestone: none
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P0
State: done
Primary Module: x/pubsub
Owned Files:
- `x/pubsub/distributed.go`
- `x/pubsub/distributed_test.go`
- `docs/modules/x-pubsub/README.md`
Depends On: none

Goal:
Harden the distributed pubsub cluster HTTP surface so auth defaults, JSON error
codes, and success payloads have one explicit contract.

Problem:
`DistributedPubSub` exposes cluster HTTP handlers for health, heartbeat,
publish, and sync. The cluster auth helper treats an empty `AuthToken` as an
open cluster, which is easy to miss and is not fail-closed. The same handlers
also mix typed payloads with `map[string]any` / `map[string]string` success
bodies, and invalid JSON currently uses domain codes such as
`CodeInvalidPayload` or `CodeInvalidMessage` instead of the repository's
canonical invalid JSON code. This makes the cluster API less predictable than
other contract-backed handlers.

Scope:
- Make unauthenticated cluster mode explicit in configuration instead of being
  implied by an empty token; keep any compatibility path documented and
  test-covered.
- Preserve timing-safe token comparison when auth is configured.
- Replace cluster success maps with small local DTO structs for health, publish,
  and sync responses.
- Normalize malformed JSON request bodies to `contract.CodeInvalidJSON` with
  validation category.
- Add negative-path tests for empty-token behavior, invalid token behavior, and
  malformed heartbeat/publish JSON.

Non-goals:
- Do not redesign the distributed broker protocol.
- Do not change topic routing, replication, or heartbeat scheduling semantics.
- Do not add external dependencies.
- Do not move cluster HTTP behavior into stable roots.

Files:
- `x/pubsub/distributed.go`
- `x/pubsub/distributed_test.go`
- `docs/modules/x-pubsub/README.md`

Tests:
- `go test -race -timeout 60s ./x/pubsub/...`
- `go test -timeout 20s ./x/pubsub/...`
- `go vet ./x/pubsub/...`

Docs Sync:
Update `docs/modules/x-pubsub/README.md` to state the distributed cluster auth
default and any explicit opt-in for unauthenticated local clusters.

Done Definition:
- Empty-token cluster auth behavior is explicit and covered by tests.
- Invalid cluster JSON uses `contract.CodeInvalidJSON`.
- Cluster success responses have local DTO shapes instead of one-off maps.
- The listed validation commands pass.

Outcome:
- Added explicit `AllowInsecureAuth` opt-in for empty-token cluster HTTP mode;
  empty `AuthToken` now fails closed by default.
- Replaced cluster health, publish, and sync success maps with local DTO
  response structs.
- Normalized malformed heartbeat and cluster publish JSON to
  `contract.CodeInvalidJSON` with validation category.
- Added tests for empty-token fail-closed behavior, explicit insecure opt-in,
  invalid token behavior, typed success payloads, and malformed JSON errors.
- Updated `docs/modules/x-pubsub/README.md` with the cluster auth default and
  invalid JSON/DTO response contract.

Validation:
- `go test -race -timeout 60s ./x/pubsub/...`
- `go test -timeout 20s ./x/pubsub/...`
- `go vet ./x/pubsub/...`
