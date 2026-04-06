# Card 0819

Priority: P1
State: active
Primary Module: x/data
Owned Files:
- `docs/modules/x-data/README.md`
- `x/data/rw/module.yaml`
- `x/data/rw/testdata/README.md`
Depends On:

Goal:
- Clarify `x/data/rw` failover, read-after-write, and health expectations in one canonical docs path.

Scope:
- Align the x/data primer, `x/data/rw` manifest wording, and the rw testdata README on primary fallback, `ForcePrimary` usage, and health-check expectations.
- Make the operational guidance explicit about when reads can use replicas and when callers should force primary routing.
- Keep the guidance constrained to implemented rw behavior and current tests.

Non-goals:
- Do not redesign rw routing policies or health behavior in this card.
- Do not broaden this card into sharding guidance.
- Do not add external-service integration requirements to the default loop.

Files:
- `docs/modules/x-data/README.md`
- `x/data/rw/module.yaml`
- `x/data/rw/testdata/README.md`

Tests:
- `go test -timeout 20s ./x/data/rw/...`
- `go test -race -timeout 60s ./x/data/rw/...`
- `go vet ./x/data/rw/...`

Docs Sync:
- Keep rw operational guidance aligned across the x/data primer, rw manifest, and rw testdata notes.

Done Definition:
- Failover, read-after-write, and health expectations are described consistently in one canonical way.
- The docs do not promise stronger consistency or health behavior than the implementation provides.
- Targeted rw validation stays green.

Outcome:
- Tightened `docs/modules/x-data/README.md`, `x/data/rw/module.yaml`, and `x/data/rw/testdata/README.md` around one canonical rule set: ordinary reads may use healthy replicas, `ForcePrimary(ctx)` is the explicit read-after-write escape hatch, and fallback to primary only happens when `FallbackToPrimary` is explicitly enabled.
- Clarified that background replica health checks are opt-in via `HealthCheck.Enabled` and rely on periodic ping probes plus configured thresholds rather than stronger freshness guarantees.
- Validation:
  - `go test -timeout 20s ./x/data/rw/...`
  - `go test -race -timeout 60s ./x/data/rw/...`
  - `go vet ./x/data/rw/...`
