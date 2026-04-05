# Card 0751

Priority: P1
State: active
Primary Module: core
Owned Files:
- `core/options.go`
- `core/app.go`
- `core/app_helpers.go`
- `core/options_test.go`
- `docs/modules/core/README.md`
Depends On:

Goal:
- Make router-affecting `core` options declarative and order-independent so
  `core.New(...)` has one predictable configuration path with no option-order
  surprises.

Problem:
- `WithMethodNotAllowed` currently calls `ensureRouter()` and mutates whatever
  router instance happens to exist at option-application time.
- That makes option order observable: `New(WithMethodNotAllowed(true), WithRouter(custom))`
  loses the setting, while the reverse order keeps it.
- Most `Option` values behave like config setters; this one behaves like an
  eager side-effect. That is inconsistent with the rest of the package and easy
  to miss in review.

Scope:
- Store router-level option intent on `App` instead of mutating the router
  eagerly during `Option` application.
- Ensure both the default router and a caller-provided router receive the same
  configured router options.
- Add regression coverage for both option orders.
- Prefer one explicit configuration model even if that means removing
  side-effectful behavior rather than preserving it.

Non-goals:
- Do not redesign `router.Router` or add new router features.
- Do not change unrelated option behavior in the same pass.
- Do not move route registration out of `core`.

Files:
- `core/options.go`
- `core/app.go`
- `core/app_helpers.go`
- `core/options_test.go`
- `docs/modules/core/README.md`

Tests:
- Add a regression test that `WithMethodNotAllowed(true)` works the same before
  and after `WithRouter(custom)`.
- `go test -race -timeout 60s ./core/...`
- `go vet ./core/...`

Docs Sync:
- Update the `core` module primer to state that router-affecting options are
  applied declaratively during app construction, not via eager side effects.

Done Definition:
- `WithMethodNotAllowed` no longer depends on option order.
- A custom router passed via `WithRouter` receives the same router-level config
  as the default router path.
- No router-affecting `core.Option` relies on eager hidden mutation.
- Regression tests cover both option orders and pass.

Outcome:
