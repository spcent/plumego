# Card 0216

Priority: P1
State: done
Primary Module: core
Owned Files:
- `core/config.go`
- `core/introspection.go`
- `core/module.yaml`
- `docs/modules/core/README.md`
- `x/devtools/devtools.go`

Goal:
- Remove debug-only runtime snapshot ownership from stable `core` so the kernel only owns app construction and lifecycle behavior, not first-party tooling payload contracts.
- Converge runtime/config snapshot exposure onto the owning debug surface instead of keeping a devtools-facing snapshot API in the stable kernel.

Problem:
- `core/module.yaml` still lists “Runtime introspection (config snapshot)” as a stable responsibility.
- `core/config.go` exports `RuntimeSnapshot` and `RuntimeTLSSnapshot`, and `core/introspection.go` exports `(*App).RuntimeSnapshot()`.
- Repository usage is effectively limited to `x/devtools` and core self-tests, which means the stable kernel is carrying a debug/tooling payload contract instead of a runtime behavior contract.
- `x/devtools` already owns the debug config endpoint and its wrapping payload, so keeping the snapshot type in `core` blurs the boundary between kernel lifecycle and debug surfacing.

Scope:
- Remove stable `core` ownership of runtime snapshot payload types and the public introspection accessor.
- Move the snapshot contract and projection logic to `x/devtools` or another owning debug-only package in the `x/observability` family.
- Update `x/devtools` hooks and tests to use the converged debug-owned snapshot surface.
- Keep `core` focused on lifecycle behavior, server preparation, and explicit route/middleware attachment.
- Sync core docs and manifest to the reduced kernel boundary.

Non-goals:
- Do not redesign `AppConfig`, `Prepare`, `Server`, or `Shutdown`.
- Do not move debug route registration into `core`.
- Do not add a new generic tooling/plugin interface to the kernel.
- Do not preserve compatibility wrappers for removed runtime snapshot APIs.

Files:
- `core/config.go`
- `core/introspection.go`
- `core/module.yaml`
- `docs/modules/core/README.md`
- `x/devtools/devtools.go`

Tests:
- `go test -timeout 20s ./core/... ./x/devtools/...`
- `go test -race -timeout 60s ./core/... ./x/devtools/...`
- `go vet ./core/... ./x/devtools/...`

Docs Sync:
- Keep the core manifest and primer aligned on the rule that stable `core` owns runtime lifecycle behavior, while debug snapshot payloads belong to `x/devtools`.

Done Definition:
- Stable `core` no longer exports runtime snapshot payload types or a public snapshot accessor for tooling.
- `x/devtools` owns the debug-facing runtime snapshot contract it serves.
- Core and devtools tests pass against the converged debug-owned snapshot surface with no residual references to removed stable introspection APIs.
- Core docs and manifest describe the same reduced kernel boundary the code implements.

Outcome:
- Completed.
- Removed debug/runtime snapshot payload ownership from stable `core`; the kernel no longer exports runtime snapshot payload types or a debug-facing snapshot accessor.
- `x/devtools` now owns the runtime snapshot contract it exposes on debug endpoints.
