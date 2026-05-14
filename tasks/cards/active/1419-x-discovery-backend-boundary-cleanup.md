# Card 1419

Milestone: v1-package-cleanup
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P2
State: active
Primary Module: x/discovery
Owned Files:
- x/discovery/discovery.go
- x/discovery/static.go
- x/discovery/backend_errors.go
- x/discovery/discovery_test.go
- x/discovery/static_test.go
Depends On:
- 1418

Goal:
- Clarify `x/discovery` core/static boundary before any backend-specific maturity work.

Scope:
- Move shared unsupported/not-supported error helpers into `backend_errors.go`.
- Keep core contracts and static backend behavior separate from Consul, Kubernetes, and etcd adapters.
- Preserve current unsupported operation behavior.

Non-goals:
- Do not change Consul, Kubernetes, or etcd behavior in this card.
- Do not promote `x/discovery`.
- Do not introduce backend dependencies.

Files:
- x/discovery/discovery.go
- x/discovery/static.go
- x/discovery/backend_errors.go
- x/discovery/discovery_test.go
- x/discovery/static_test.go

Tests:
- go test -timeout 20s ./x/discovery
- go vet ./x/discovery
- go run ./internal/checks/dependency-rules

Docs Sync:
- Update `docs/modules/x-discovery/README.md` only if core/static guidance changes.

Done Definition:
- Core/static discovery boundary is clearer.
- Unsupported operation behavior remains tested.
- Discovery tests and dependency checks pass.

Outcome:

