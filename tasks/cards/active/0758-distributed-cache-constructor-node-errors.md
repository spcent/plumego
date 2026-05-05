# Card 0758

Milestone:
Recipe: specs/change-recipes/store-stability.yaml
Priority: P2
State: active
Primary Module: x/cache/distributed
Owned Files:
- x/cache/distributed/distributed.go
- x/cache/distributed/distributed_test.go
Depends On:

Goal:
Stop distributed cache construction from silently ignoring invalid node additions.

Scope:
- Add an error-returning constructor path that validates ring.Add outcomes before health checker registration.
- Preserve source compatibility where practical for existing callers.
- Add regression coverage for duplicate or invalid node input.

Non-goals:
- Do not change hash ring internals beyond constructor wiring.
- Do not redesign health checking.

Files:
- x/cache/distributed/distributed.go
- x/cache/distributed/distributed_test.go

Tests:
- go test ./x/cache/distributed

Docs Sync:
- Update docs only if constructor API guidance changes.

Done Definition:
- Invalid constructor input is observable as an error through the new stable construction path.
- Existing distributed cache tests pass.

Outcome:

