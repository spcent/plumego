# Card 0268

Priority: P1
State: done
Primary Module: x/tenant/core
Owned Files:
- `x/tenant/core/config.go`
- `x/tenant/core/middleware.go`
- `x/tenant/core/policy.go`
- `x/tenant/core/quota.go`
- `x/tenant/core/ratelimit.go`
- `x/tenant/core/route_policy.go`
Depends On:

Goal:
- Collect all sentinel errors from `x/tenant/core` into a single `errors.go` file so they are discoverable at a glance and consistent with contract's pattern.

Problem:
- Six sentinel errors are declared inline at the top of six different files, with no central location:
  - `ErrTenantNotFound` in `config.go:33`
  - `ErrInvalidTenantID` in `middleware.go:12`
  - `ErrPolicyDenied` in `policy.go:9`
  - `ErrQuotaExceeded` in `quota.go:10`
  - `ErrRateLimitExceeded` in `ratelimit.go:10`
  - `ErrRoutePolicyNotFound` in `route_policy.go:10`
- The `contract` package groups all its sentinel errors together in `auth.go` and each thematic file; `x/tenant/core` is inconsistent with that and with itself.
- A caller importing the package cannot find all errors without reading every source file.

Scope:
- Create `x/tenant/core/errors.go` and move all six `var Err*` declarations there.
- Remove the individual declarations from their current files (the package-level `var` blocks at the top of each).
- No behavior change; compilation and all tests must stay green.

Non-goals:
- Do not change error string values or sentinel identities.
- Do not rename any errors.
- Do not add new errors.

Files:
- `x/tenant/core/errors.go` (new)
- `x/tenant/core/config.go`
- `x/tenant/core/middleware.go`
- `x/tenant/core/policy.go`
- `x/tenant/core/quota.go`
- `x/tenant/core/ratelimit.go`
- `x/tenant/core/route_policy.go`

Tests:
- `go test -timeout 20s ./x/tenant/core/...`
- `go vet ./x/tenant/core/...`

Docs Sync:
- None required.

Done Definition:
- `x/tenant/core/errors.go` contains all six sentinel error declarations.
- No other file in the package declares a `var Err*` sentinel.
- All existing tests pass without modification.

Outcome:
- Pending.
