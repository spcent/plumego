# Card 0758

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P1
State: active
Primary Module: x/data/rw
Owned Files:
- x/data/rw/loadbalancer.go
- x/data/rw/loadbalancer_test.go
- docs/modules/x-data/README.md
Depends On:
- 0757-x-data-sharding-sql-support-boundary

Goal:
Make direct WeightedBalancer use behave consistently when no weights are configured.

Scope:
- Preserve round-robin state for empty-weight WeightedBalancer instances.
- Keep invalid-weight behavior explicit.
- Add a direct API regression test.
- Update docs for direct weighted balancer construction.

Non-goals:
- Do not change rw.New replica weight validation.
- Do not alter least-connection or random balancers.
- Do not change SQL routing policy.

Files:
- x/data/rw/loadbalancer.go
- x/data/rw/loadbalancer_test.go
- docs/modules/x-data/README.md

Tests:
- go test -timeout 20s ./x/data/rw
- go test -race -timeout 60s ./x/data/rw
- go vet ./x/data/rw

Docs Sync:
- Document empty weights behave like stateful round-robin.

Done Definition:
- Direct NewWeightedBalancer(nil) rotates across healthy replicas.
- Tests cover empty weights and reset behavior.
