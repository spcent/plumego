# Card 0732

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P0
State: active
Primary Module: x/data/rw
Owned Files:
- x/data/rw/cluster.go
- x/data/rw/policy.go
- x/data/rw/cluster_test.go
- x/data/rw/policy_test.go
Depends On:
- 0731-x-data-sharding-boundary-log-safety

Goal:
Make read-write routing fail closed and expose routing errors consistently.

Scope:
- Make `QueryRowContext` surface routing errors through `Scan` instead of returning an empty row.
- Ensure replica preference cannot route obvious writes to replicas.
- Treat unknown or unsafe SQL classifications conservatively.
- Fix misleading fallback defaults in comments or implementation.

Non-goals:
- Do not introduce a full SQL parser.
- Do not change `ExecContext` primary routing.
- Do not redesign load balancers.

Files:
- x/data/rw/cluster.go
- x/data/rw/policy.go
- x/data/rw/cluster_test.go
- x/data/rw/policy_test.go

Tests:
- go test -timeout 20s ./x/data/rw
- go test -race -timeout 60s ./x/data/rw
- go vet ./x/data/rw

Docs Sync:
- Update docs only if public routing defaults or context hints change.

Done Definition:
- `QueryContext` and `QueryRowContext` expose routing failures consistently.
- Context preference cannot force write-like statements to replicas.
- rw normal, race, and vet checks pass.
