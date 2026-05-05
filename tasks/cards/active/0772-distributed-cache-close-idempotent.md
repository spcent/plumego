# Card 0772

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P3
State: active
Primary Module: x/cache
Owned Files: x/cache/distributed/node.go, x/cache/distributed/distributed.go, x/cache/distributed/distributed_test.go
Depends On:

Goal:

Make distributed cache Close/health Stop lifecycle idempotent so repeated cleanup does not panic.

Scope:

- Guard health checker Stop with sync.Once or equivalent.
- Keep existing Start behavior unchanged.
- Add regression coverage for repeated Close.

Non-goals:

- Redesigning health check scheduling.
- Adding lifecycle state inspection APIs.
- Changing node health policies.

Files:

- x/cache/distributed/node.go
- x/cache/distributed/distributed.go
- x/cache/distributed/distributed_test.go

Tests:

- go test -race -timeout 60s ./x/cache/distributed
- go test -timeout 20s ./x/cache/distributed
- go vet ./x/cache/distributed

Docs Sync:

- Not required; this hardens lifecycle cleanup.

Done Definition:

- Calling Close repeatedly is safe.
- Health checker Stop remains safe after Start.
- Package tests and vet pass.

Outcome:

