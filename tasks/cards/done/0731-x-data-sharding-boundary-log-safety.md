# Card 0731

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P0
State: done
Primary Module: x/data/sharding
Owned Files:
- x/data/sharding/logging.go
- x/data/sharding/cluster.go
- x/data/sharding/logging_test.go
- x/data/sharding/module.yaml
- docs/modules/x-data/README.md
Depends On:
- 0730-x-data-sharding-placeholder-resolution

Goal:
Make sharding comply with its module boundary and stop logging raw SQL, query args, or shard keys by default.

Scope:
- Remove direct dependency on `contract` from sharding logging.
- Remove or relocate app-bootstrap `store/db` ownership from sharding.
- Log only safe query metadata such as operation, shard index, latency, policy, and redaction markers.
- Add focused tests for redacted logging behavior.

Non-goals:
- Do not add a generic tracing/export backend.
- Do not redesign router constructors.
- Do not change shard resolution logic.

Files:
- x/data/sharding/logging.go
- x/data/sharding/cluster.go
- x/data/sharding/logging_test.go
- x/data/sharding/module.yaml
- docs/modules/x-data/README.md

Tests:
- go test -timeout 20s ./x/data/sharding
- go vet ./x/data/sharding
- go run ./internal/checks/dependency-rules

Docs Sync:
- Update `docs/modules/x-data/README.md` if public logging or cluster ownership semantics change.

Done Definition:
- `x/data/sharding` no longer imports forbidden `contract` or out-of-manifest `store/db`.
- Logging helpers do not emit raw SQL text or shard-key values by default.
- Sharding tests, vet, and dependency rules pass.

Outcome:
- Removed sharding logging's direct `contract` dependency and stopped deriving request metadata from transport context.
- Changed sharding logging to emit operation, shard index, latency, policy, and redaction markers instead of raw SQL, rewritten SQL, or shard-key values.
- Changed `ClusterConfig` to accept existing `*sql.DB` handles so sharding no longer imports or opens `store/db` configuration.
- Updated x/data docs to state that sharding logs and traces redact raw statements, arguments, and shard keys.

Validation:
- GOCACHE=/private/tmp/plumego-go-build go test -timeout 20s ./x/data/sharding
- GOCACHE=/private/tmp/plumego-go-build go vet ./x/data/sharding
- GOCACHE=/private/tmp/plumego-go-build go run ./internal/checks/dependency-rules
