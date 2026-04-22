# Card 2114: Middleware Constructor Panic Contract

Milestone: none
Recipe: specs/change-recipes/add-middleware.yaml
Priority: medium
State: done
Primary Module: middleware
Owned Files:
- middleware/accesslog/accesslog.go
- middleware/accesslog/accesslog_test.go
- middleware/recovery/recover.go
- middleware/recovery/recover_test.go
- docs/modules/middleware/README.md
Depends On: none

Goal:
- Make stateful middleware constructor failure behavior explicit and consistent.
- Avoid panic-only nil dependency handling as the only constructor path for access log and recovery middleware.

Scope:
- Audit `middleware/accesslog` and `middleware/recovery` constructors for nil logger handling.
- Add explicit error-returning constructor variants or another documented canonical pattern that preserves existing API compatibility.
- Keep middleware responsibilities transport-only and `func(http.Handler) http.Handler` compatible.
- Add tests for nil logger behavior and default constructor compatibility.

Non-goals:
- Do not introduce observability exporter catalogs, business DTO assembly, tenant policy, or service injection into stable middleware.
- Do not remove existing public constructors without following the symbol-change protocol.
- Do not alter request/response logging semantics beyond constructor validation.

Files:
- `middleware/accesslog/accesslog.go`: normalize constructor validation behavior.
- `middleware/accesslog/accesslog_test.go`: cover nil logger and compatibility behavior.
- `middleware/recovery/recover.go`: normalize constructor validation behavior.
- `middleware/recovery/recover_test.go`: cover nil logger and compatibility behavior.
- `docs/modules/middleware/README.md`: document constructor conventions if behavior changes.

Tests:
- `go test -race -timeout 60s ./middleware/accesslog ./middleware/recovery`
- `go test -timeout 20s ./middleware/accesslog ./middleware/recovery`
- `go vet ./middleware/accesslog ./middleware/recovery`

Docs Sync:
- Required if constructor behavior or public constructor guidance changes.

Done Definition:
- Access log and recovery middleware expose a clear non-panic construction path for nil dependencies.
- Existing public constructor compatibility is preserved or migrated with symbol-change protocol evidence.
- Focused tests cover both constructor families.
- The three listed validation commands pass.

Outcome:
- Added `accesslog.MiddlewareE` and `recovery.RecoveryE` as explicit non-panic constructor paths for nil logger dependencies.
- Preserved existing `accesslog.Middleware` and `recovery.Recovery` panic compatibility by delegating to the error-returning variants.
- Added focused nil-dependency and successful-construction tests for both middleware packages.
- Documented the `E` constructor variant convention in `docs/modules/middleware/README.md`.
- Validation passed:
  - `go test -race -timeout 60s ./middleware/accesslog ./middleware/recovery`
  - `go test -timeout 20s ./middleware/accesslog ./middleware/recovery`
  - `go vet ./middleware/accesslog ./middleware/recovery`
