# Card 1204

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
- Updated docs/modules/x-cache/README.md to prefer distributed.NewE for dynamic node sets.

Done Definition:
- Invalid constructor input is observable as an error through the new stable construction path.
- Existing distributed cache tests pass.

Outcome:
- Added distributed.NewE to return constructor validation errors from ring.Add.
- Updated New to avoid returning a partially initialized cache on constructor
  errors while preserving its existing signature.
- Added duplicate-node constructor regression tests.
- Documented the NewE guidance in the x/cache module primer.
- Validated with:
  - go test -timeout 20s ./x/cache/...
  - go test -race -timeout 60s ./x/cache/...
  - go vet ./x/cache/...
