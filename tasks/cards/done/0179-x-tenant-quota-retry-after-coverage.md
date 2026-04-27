# Card 0179

Priority: P1
State: done
Primary Module: x/tenant
Owned Files:
- `x/tenant/core/quota_test.go`
- `x/tenant/quota/middleware_test.go`
- `x/tenant/transport/transport_test.go`
Depends On:

Goal:
- Expand quota exhaustion and `Retry-After` coverage for tenant-aware request limiting.

Scope:
- Add focused tests that lock in quota exhaustion behavior, positive `Retry-After` propagation, and remaining-budget headers on the transport path.
- Cover the fail-closed path where quota middleware rejects an exhausted tenant before the handler runs.
- Keep the assertions local to tenant quota and transport helpers.

Non-goals:
- Do not redesign quota algorithms.
- Do not broaden into rate-limit token bucket behavior.
- Do not add tenant onboarding or config-management work.

Files:
- `x/tenant/core/quota_test.go`
- `x/tenant/quota/middleware_test.go`
- `x/tenant/transport/transport_test.go`

Tests:
- `go test -timeout 20s ./x/tenant/core ./x/tenant/quota ./x/tenant/transport`
- `go test -race -timeout 60s ./x/tenant/core ./x/tenant/quota ./x/tenant/transport`

Docs Sync:
- Touch tenant docs only if the new transport assertions expose a mismatch in documented quota or `Retry-After` behavior.

Done Definition:
- Quota-exhaustion and `Retry-After` behavior are covered at both manager and middleware/transport boundaries.
- New tests stay explicit about fail-closed denial rather than relying on implicit handler behavior.
- Coverage remains local and fast.

Outcome:
- Added quota-manager tests that lock in denied-result semantics for remaining budget and positive `Retry-After`.
- Expanded quota middleware coverage so rejected requests now assert `Retry-After`, remaining-budget headers, error codes, and handler short-circuiting together.
- Added a transport helper composition test to keep quota and `Retry-After` headers aligned on the response path.

Validation Run:
- `go test -timeout 20s ./x/tenant/core ./x/tenant/quota ./x/tenant/transport`
- `go test -race -timeout 60s ./x/tenant/core ./x/tenant/quota ./x/tenant/transport`
