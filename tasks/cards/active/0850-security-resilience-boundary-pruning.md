# Card 0850

Priority: P1
State: active
Primary Module: security
Owned Files:
- `security/resilience/circuitbreaker`
- `security/module.yaml`
- `docs/modules/security/README.md`
- `x/gateway`
- `docs/modules/x-gateway/README.md`

Goal:
- Remove non-auth resilience ownership from stable `security`.
- Move circuit-breaker runtime and HTTP middleware to an extension-owned gateway/resilience package with one canonical entrypoint.

Problem:
- `security/resilience/circuitbreaker` is a real stable subpackage, but `security/module.yaml` and the module docs only describe auth, headers, input safety, abuse, password, and JWT helpers.
- The package mixes a stateful resilience primitive with transport error writing and HTTP middleware, which is not part of the stable `security` boundary defined in `AGENTS.md` and the module manifest.
- Because the package lives under stable `security`, the repo currently exposes a hidden public surface that is outside the declared stable auth primitive contract.

Scope:
- Move the circuit breaker package out of stable `security` into an owning `x/gateway` resilience package.
- Keep the core breaker primitive and its HTTP middleware together in the owning extension package.
- Remove the stable `security/resilience/circuitbreaker` package entirely in the same change; do not leave forwarding wrappers.
- Update docs and manifests so the public boundary matches the code.

Non-goals:
- Do not redesign circuit-breaker semantics.
- Do not add new resilience features or config knobs.
- Do not change unrelated security packages.

Files:
- `security/resilience/circuitbreaker`
- `security/module.yaml`
- `docs/modules/security/README.md`
- `x/gateway`
- `docs/modules/x-gateway/README.md`

Tests:
- `go test -timeout 20s ./security/... ./x/gateway/...`
- `go test -race -timeout 60s ./security/... ./x/gateway/...`
- `go vet ./security/... ./x/gateway/...`

Docs Sync:
- Keep security docs aligned on the rule that stable `security` owns auth/security primitives, not resilience middleware or edge transport policies.
- Document the new owning `x/gateway` package as the canonical home for circuit-breaker middleware/runtime.

Done Definition:
- Stable `security` no longer exports resilience/circuit-breaker packages.
- The circuit-breaker runtime and middleware live under `x/gateway` with no compatibility wrappers left behind.
- Security and gateway docs/manifests describe the same boundary the code implements.
