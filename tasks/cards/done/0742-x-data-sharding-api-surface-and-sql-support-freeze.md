# Card 0742

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P2
State: done
Primary Module: x/data/sharding
Owned Files:
- x/data/sharding/module.yaml
- x/data/sharding/rewriter.go
- x/data/sharding/rewriter_test.go
- x/data/sharding/cluster.go
- docs/modules/x-data/README.md
Depends On:
- 0741-x-data-rw-lifecycle-and-balancer-validation

Goal:
Align sharding API inventory and SQL support limits with the behavior that can be safely supported.

Scope:
- Document and test fail-closed SQL rewrite support for complex SELECT shapes.
- Align module manifest public entrypoints with intentional sharding APIs.
- Clarify ClusterDB as a convenience wrapper over Router and rw clusters.
- Keep the API surface stable-compatible without adding new dependencies.

Non-goals:
- Do not add a SQL parser.
- Do not remove exported types without a separate symbol-change card.
- Do not implement result merging.

Files:
- x/data/sharding/module.yaml
- x/data/sharding/rewriter.go
- x/data/sharding/rewriter_test.go
- x/data/sharding/cluster.go
- docs/modules/x-data/README.md

Tests:
- go test -timeout 20s ./x/data/sharding
- go vet ./x/data/sharding
- go run ./internal/checks/module-manifests

Docs Sync:
- Update docs/modules/x-data/README.md and module manifest to match implemented support only.

Done Definition:
- Manifest and docs no longer omit intentional public sharding entrypoints.
- SQL rewrite limits are tested and documented.
- ClusterDB ownership and convenience-wrapper status are explicit.

Outcome:
- Added ClusterDB convenience-wrapper API symbols, strategy constructors, and
  rewrite symbols to the sharding module manifest public entrypoint inventory.
- SQL rewrite now rejects CTE and multiple-statement queries before parsing, in
  addition to nested SELECT and UNION.
- Added rewrite tests for CTE and multiple-statement refusal.
- Clarified ClusterDB as a convenience wrapper over Router and rw.Cluster, with
  New(ClusterConfig) taking ownership of configured DB handles.
- Updated x/data docs for the tested SQL rewrite support limits.

Validation:
- GOCACHE=/private/tmp/plumego-go-build go test -timeout 20s ./x/data/sharding
- GOCACHE=/private/tmp/plumego-go-build go test -race -timeout 60s ./x/data/sharding
- GOCACHE=/private/tmp/plumego-go-build go vet ./x/data/sharding
- GOCACHE=/private/tmp/plumego-go-build go run ./internal/checks/module-manifests
