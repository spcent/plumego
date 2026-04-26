# Card 2266

Milestone:
Recipe: specs/change-recipes/tenant-policy-change.yaml
Priority: P2
State: active
Primary Module: x/tenant
Owned Files:
- x/tenant/example_test.go
- docs/modules/x-tenant/README.md
Depends On: 2265

Goal:
Add a scenario-oriented `x/tenant` example for multi-tenant API adoption.

Scope:
- Add an offline example that wires tenant resolution, policy, quota, and rate limit in the canonical middleware chain order.
- Demonstrate one allow path and one fail-closed deny path.
- Update the tenant primer to identify this as the SaaS API starting point.

Non-goals:
- Do not add tenant CRUD or onboarding.
- Do not change policy, quota, or rate-limit behavior.
- Do not move tenant middleware into stable `middleware`.

Files:
- `x/tenant/example_test.go`
- `docs/modules/x-tenant/README.md`

Tests:
- `go test -timeout 20s ./x/tenant/...`
- `go vet ./x/tenant/...`

Docs Sync:
- Required in the `x/tenant` primer.

Done Definition:
- The example is runnable through `go test`.
- It shows explicit tenant chain wiring and at least one fail-closed response path.

Outcome:
