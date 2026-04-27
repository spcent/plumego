# Card 0231

Priority: P1
State: done
Primary Module: security
Owned Files:
- `security/resilience/circuitbreaker`
- `security/module.yaml`
- `docs/modules/security/README.md`
- `x/resilience`
- `docs/modules/x-resilience/README.md`
- `specs/extension-taxonomy.yaml`
- `specs/task-routing.yaml`

Goal:
- Remove non-auth resilience ownership from stable `security`.
- Move circuit-breaker runtime and HTTP middleware to a reusable extension-owned resilience package with one canonical entrypoint.

Problem:
- `security/resilience/circuitbreaker` is a real stable subpackage, but `security/module.yaml` and the module docs only describe auth, headers, input safety, abuse, password, and JWT helpers.
- The package mixes a stateful resilience primitive with transport error writing and HTTP middleware, which is not part of the stable `security` boundary defined in `AGENTS.md` and the module manifest.
- Because the package lives under stable `security`, the repo currently exposes a hidden public surface that is outside the declared stable auth primitive contract.

Scope:
- Move the circuit breaker package out of stable `security` into an owning `x/resilience` package.
- Keep the core breaker primitive and its HTTP middleware together in the owning extension package.
- Remove the stable `security/resilience/circuitbreaker` package entirely in the same change; do not leave forwarding wrappers.
- Update docs, manifests, and extension taxonomy so the public boundary matches the code.

Non-goals:
- Do not redesign circuit-breaker semantics.
- Do not add new resilience features or config knobs.
- Do not change unrelated security packages.

Files:
- `security/resilience/circuitbreaker`
- `security/module.yaml`
- `docs/modules/security/README.md`
- `x/resilience`
- `docs/modules/x-resilience/README.md`
- `specs/extension-taxonomy.yaml`
- `specs/task-routing.yaml`

Tests:
- `go test -timeout 20s ./security/... ./x/resilience/... ./x/gateway/...`
- `go test -race -timeout 60s ./security/... ./x/resilience/... ./x/gateway/...`
- `go vet ./security/... ./x/resilience/... ./x/gateway/...`

Docs Sync:
- Keep security docs aligned on the rule that stable `security` owns auth/security primitives, not resilience middleware or edge transport policies.
- Document the new owning `x/resilience` package as the canonical home for reusable circuit-breaker middleware/runtime.

Done Definition:
- Stable `security` no longer exports resilience/circuit-breaker packages.
- The circuit-breaker runtime and middleware live under `x/resilience` with no compatibility wrappers left behind.
- Security, resilience, and extension-taxonomy docs/manifests describe the same boundary the code implements.

Outcome:
- Completed.
- Moved the reusable circuit-breaker runtime and HTTP middleware from `security/resilience/circuitbreaker` to `x/resilience/circuitbreaker` with no forwarding wrappers left behind.
- Updated gateway callers, boundary tests, manifests, and architecture docs so reusable resilience primitives now route to `x/resilience` instead of stable `security`.
- Added `x/resilience` to the extension taxonomy and task-routing guidance, and narrowed stable security docs/manifests back to auth/security ownership only.
