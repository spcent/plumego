# Card 1215

Milestone:
Recipe: specs/change-recipes/store-stability.yaml
Priority: P2
State: done
Primary Module: x/cache/distributed
Owned Files:
- x/cache/distributed/distributed.go
- x/cache/distributed/distributed_test.go
Depends On:

Goal:
Make DistributedCache.Clear fail when no healthy node is actually cleared.

Scope:
- Track attempted healthy nodes in Clear.
- Return ErrNodeUnhealthy when all nodes are skipped.
- Add focused regression coverage.

Non-goals:
- Do not change Clear's all-node broadcast model.
- Do not add partial-success reporting.

Files:
- x/cache/distributed/distributed.go
- x/cache/distributed/distributed_test.go

Tests:
- go test ./x/cache/distributed

Docs Sync:
- Not required unless public docs state Clear behavior.

Done Definition:
- Clear cannot return nil after zero attempted clears.
- Distributed cache tests pass.

Outcome:
- Added attempted-node accounting to DistributedCache.Clear.
- Clear now returns ErrNodeUnhealthy when every node is skipped as unhealthy.
- Added regression coverage for all-unhealthy Clear.
- Validated with:
  - go test -timeout 20s ./x/cache/distributed
  - go test -race -timeout 60s ./x/cache/distributed
  - go vet ./x/cache/distributed
