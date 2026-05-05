# Card 0749

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P2
State: active
Primary Module: x/data/rw
Owned Files:
- x/data/rw/cluster.go
- x/data/rw/loadbalancer.go
- x/data/rw/cluster_test.go
- x/data/rw/loadbalancer_test.go
- docs/modules/x-data/README.md
Depends On:
- 0748-x-data-file-resource-boundaries

Goal:
Make rw lifecycle ownership and weighted balancing API behavior clearer.

Scope:
- Add NewWeightedBalancerE for callers that want constructor-time validation.
- Make WeightedBalancer.Next reject weight length mismatch with replicas.
- Allow cluster health-check context ownership to be explicit through Config.
- Add tests for mismatch rejection and context-cancelled health checker shutdown.

Non-goals:
- Do not remove NewWeightedBalancer.
- Do not redesign routing policies.
- Do not change sql.DB ownership semantics.

Files:
- x/data/rw/cluster.go
- x/data/rw/loadbalancer.go
- x/data/rw/cluster_test.go
- x/data/rw/loadbalancer_test.go
- docs/modules/x-data/README.md

Tests:
- go test -timeout 20s ./x/data/rw
- go test -race -timeout 60s ./x/data/rw
- go vet ./x/data/rw

Docs Sync:
- Update x/data docs for context ownership and weighted balancer validation.

Done Definition:
- Weighted balancer direct use rejects replica/weight length mismatch.
- NewWeightedBalancerE validates weights at construction.
- Cluster health checks can inherit a caller-owned context.

Outcome:
