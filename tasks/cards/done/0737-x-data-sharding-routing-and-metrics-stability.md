# Card 0737

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P1
State: done
Primary Module: x/data/sharding
Owned Files:
- x/data/sharding/router.go
- x/data/sharding/rule.go
- x/data/sharding/router_test.go
- x/data/sharding/rule_test.go
- docs/modules/x-data/README.md
Depends On:
- 0736-x-data-file-s3-doc-readiness

Goal:
Make sharding routing, rule access, and router metrics stable enough to fail closed and report honest counters.

Scope:
- Make QueryRowContext match QueryContext fail-closed behavior on resolution and shard validation errors.
- Ensure cross-shard fan-out does not increment single-shard counters.
- Make ShardingRuleRegistry safe for concurrent reads and writes or clearly enforce immutable behavior in code.
- Add tests for QueryRowContext default-shard refusal, cross-shard metrics, and registry race safety.

Non-goals:
- Do not implement merged cross-shard result sets.
- Do not add a third-party SQL parser.
- Do not change strategy hashing/range/list algorithms.

Files:
- x/data/sharding/router.go
- x/data/sharding/rule.go
- x/data/sharding/router_test.go
- x/data/sharding/rule_test.go
- docs/modules/x-data/README.md

Tests:
- go test -timeout 20s ./x/data/sharding
- go test -race -timeout 60s ./x/data/sharding
- go vet ./x/data/sharding

Docs Sync:
- Update x/data docs for fail-closed QueryRowContext and cross-shard metrics semantics.

Done Definition:
- QueryRowContext never falls back to an unrelated default shard after routing errors.
- Cross-shard fan-out increments cross-shard and per-shard counters without inflating single-shard totals.
- Rule registry operations are race-safe under concurrent access.

Outcome:
- QueryRowContext now returns routing and shard-validation failures from Scan
  instead of falling back to DefaultShardIndex.
- Cross-shard fan-out records per-shard executions without incrementing the
  single-shard query total.
- ShardingRuleRegistry map access is protected by an RWMutex, with a race-test
  covering concurrent register/read/snapshot paths.
- Updated x/data docs for QueryRowContext fail-closed behavior, registry
  immutability expectations, and router metrics semantics.

Validation:
- GOCACHE=/private/tmp/plumego-go-build go test -timeout 20s ./x/data/sharding
- GOCACHE=/private/tmp/plumego-go-build go test -race -timeout 60s ./x/data/sharding
- GOCACHE=/private/tmp/plumego-go-build go vet ./x/data/sharding
