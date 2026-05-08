# Card 0750

Milestone:
Recipe: specs/change-recipes/store-stable.yaml
Priority: P2
State: done
Primary Module: x/cache/distributed
Owned Files:
- x/cache/distributed/distributed.go
- x/cache/distributed/distributed_test.go
Depends On:

Goal:
Prevent distributed cache Delete from reporting success when no healthy replica was touched.

Scope:
- Count healthy delete attempts.
- Return an unhealthy/no-node error when Delete skips every replica.
- Add regression coverage for all-unhealthy replicas.

Non-goals:
- Do not introduce quorum configuration.
- Do not change hash-ring behavior.

Files:
- x/cache/distributed/distributed.go
- x/cache/distributed/distributed_test.go

Tests:
- go test -timeout 20s ./x/cache/distributed

Docs Sync:
- None.

Done Definition:
- Delete returns an error when it deletes from zero healthy nodes.
- Existing delete success paths still pass.
- Targeted tests pass.

Outcome:
Distributed Delete now counts healthy delete attempts and returns ErrNodeUnhealthy when every replica is skipped. Added regression coverage for all-unhealthy replica sets.

Validation:
- go test -timeout 20s ./x/cache/distributed
