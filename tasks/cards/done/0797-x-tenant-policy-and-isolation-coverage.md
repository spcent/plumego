# Card 0797

Priority: P1
State: done
Primary Module: x/tenant
Owned Files:
- `x/tenant/core/policy_test.go`
- `x/tenant/resolve/middleware_test.go`
- `x/tenant/policy/middleware_test.go`
- `x/tenant/ratelimit/middleware_test.go`
Depends On:

Goal:
- Expand deny-path and isolation coverage across tenant resolution and policy/rate-limit middleware.

Scope:
- Add focused tests for isolation-sensitive resolution order, invalid tenant IDs, explicit policy-deny behavior, and tenant-scoped rate-limit separation.
- Keep coverage at the middleware and evaluator boundary instead of introducing a reference app.
- Preserve fail-closed expectations when tenant context or policy decisions are missing or denied.

Non-goals:
- Do not change middleware ordering semantics unless a failing test proves a real issue.
- Do not add cross-module reference wiring.
- Do not widen this card into store adapter work.

Files:
- `x/tenant/core/policy_test.go`
- `x/tenant/resolve/middleware_test.go`
- `x/tenant/policy/middleware_test.go`
- `x/tenant/ratelimit/middleware_test.go`

Tests:
- `go test -timeout 20s ./x/tenant/core ./x/tenant/resolve ./x/tenant/policy ./x/tenant/ratelimit`
- `go test -race -timeout 60s ./x/tenant/core ./x/tenant/resolve ./x/tenant/policy ./x/tenant/ratelimit`

Docs Sync:
- Touch tenant docs only if the new isolation assertions expose a mismatch in the documented resolution or deny-path behavior.

Done Definition:
- Resolution, policy, and rate-limit tests cover the key deny and isolation paths the roadmap calls out.
- Fail-closed behavior is asserted explicitly.
- No stable-root boundary drift is introduced.

Outcome:
- Added focused policy-evaluator isolation tests so one tenant's allowed-model policy cannot bleed into another tenant's decisions.
- Expanded resolve, policy, and rate-limit middleware coverage for principal precedence, invalid tenant IDs, canonical deny responses, handler short-circuiting, and per-tenant rate-limit separation.
- Kept the new coverage package-local and fail-closed without widening any tenant behavior into stable roots.

Validation Run:
- `go test -timeout 20s ./x/tenant/core ./x/tenant/resolve ./x/tenant/policy ./x/tenant/ratelimit`
- `go test -race -timeout 60s ./x/tenant/core ./x/tenant/resolve ./x/tenant/policy ./x/tenant/ratelimit`
