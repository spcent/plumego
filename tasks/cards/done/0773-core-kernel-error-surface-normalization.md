# Card 0773

Priority: P2
State: done
Primary Module: core
Owned Files:
- `core/app_helpers.go`
- `core/routing.go`
- `core/http_handler.go`
- `core/lifecycle.go`
- `core/lifecycle_test.go`
Depends On:
- `0772-core-kernel-invariant-hardening.md`

Goal:
- Converge `core` onto one kernel error contract so lifecycle and mutation
  paths stop mixing wrapped kernel errors with bare `fmt.Errorf(...)`.

Problem:
- Route registration and mutability failures currently use
  `contract.WrapError(...)` with operation metadata.
- `Prepare`, `Server`, `Start`, and server-preparation failures still return
  plain `fmt.Errorf(...)`.
- This leaves `core` with two different caller-facing error styles for the same
  kernel layer, which makes logging, testing, and API expectations harder to
  reason about.

Scope:
- Choose one canonical exported error style for `core`.
- Normalize lifecycle / preparation errors to that style.
- Keep messages concrete and stable enough for first-party callers/tests.

Non-goals:
- Do not redesign HTTP response error writing.
- Do not add feature-specific error helpers.
- Do not widen the public API with new exported error families unless strictly
  required by the chosen contract.

Files:
- `core/app_helpers.go`
- `core/routing.go`
- `core/http_handler.go`
- `core/lifecycle.go`
- `core/lifecycle_test.go`

Tests:
- `go test -race -timeout 60s ./core/...`
- `go vet ./core/...`

Docs Sync:
- Update docs only if the exported caller contract becomes observably different.

Done Definition:
- Exported `core` methods use one consistent kernel error style.
- Lifecycle/preparation failures no longer diverge from mutation failures.
- Tests assert against the normalized contract instead of mixed ad hoc strings.

Outcome:
- Added one internal `wrapCoreError(...)` path and used it across mutability,
  routing, preparation, and lifecycle failures.
- Normalized `Prepare`, `Server`, `Start`, and `Shutdown` to return wrapped
  kernel errors with `core` module metadata instead of bare `fmt.Errorf(...)`.
- Updated lifecycle tests to assert the wrapped error contract directly.
