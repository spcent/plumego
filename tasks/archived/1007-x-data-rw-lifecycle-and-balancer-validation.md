# Card 1007

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P2
State: done
Primary Module: x/data/rw
Owned Files:
- x/data/rw/cluster.go
- x/data/rw/health.go
- x/data/rw/loadbalancer.go
- x/data/rw/cluster_test.go
- x/data/rw/loadbalancer_test.go
Depends On:
- 0740-x-data-file-streaming-and-local-durability

Goal:
Make rw cluster lifecycle and weighted balancing predictable under stable use.

Scope:
- Make Cluster.Close and HealthChecker.Stop idempotent.
- Validate weighted balancer inputs so non-positive weights cannot silently route traffic.
- Preserve existing New(Config) signature.
- Add tests for double Close/Stop and invalid weights.

Non-goals:
- Do not redesign read/write routing policy.
- Do not change sql.DB ownership rules.
- Do not add background supervisor abstractions.

Files:
- x/data/rw/cluster.go
- x/data/rw/health.go
- x/data/rw/loadbalancer.go
- x/data/rw/cluster_test.go
- x/data/rw/loadbalancer_test.go

Tests:
- go test -timeout 20s ./x/data/rw
- go test -race -timeout 60s ./x/data/rw
- go vet ./x/data/rw

Docs Sync:
- Not required unless behavior changes need the x/data module README.

Done Definition:
- Repeated Close/Stop calls do not panic.
- Invalid replica weights are rejected at cluster construction.
- Weighted balancer never selects zero or negative weight replicas.

Outcome:
- Cluster.Close now uses sync.Once and returns the first close result
  idempotently.
- HealthChecker.Stop now uses sync.Once and can be called repeatedly without
  panicking.
- rw.New validates ReplicaWeights length and positive values before constructing
  a weighted balancer.
- WeightedBalancer now rejects non-positive weights from direct use and skips
  invalid weights defensively.
- Updated x/data docs for weight validation and cluster ownership/close
  semantics.

Validation:
- GOCACHE=/private/tmp/plumego-go-build go test -timeout 20s ./x/data/rw
- GOCACHE=/private/tmp/plumego-go-build go test -race -timeout 60s ./x/data/rw
- GOCACHE=/private/tmp/plumego-go-build go vet ./x/data/rw
