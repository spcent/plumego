# Card 0228

Priority: P1
State: done
Primary Module: middleware
Owned Files:
- `middleware/compression/gzip.go`
- `middleware/timeout/timeout.go`
- `middleware/module.yaml`
- `docs/modules/middleware/README.md`

Goal:
- Finish the constructor convergence work in stable `middleware` by removing the last duplicate wrapper pairs in configurable middleware packages.
- Keep `compression` and `timeout` on one explicit exported constructor path each.

Problem:
- `middleware/compression` still exports both `Gzip()` and `GzipWithConfig(...)` for one behavior.
- `middleware/timeout` still exports both `Timeout(duration)` and `TimeoutWithConfig(...)` for one behavior.
- `docs/modules/middleware/README.md` says middleware packages should keep one constructor path, but these two packages still preserve legacy convenience wrappers.
- Leaving both entrypoints keeps stable middleware inconsistent even after previous canonical-entrypoint pruning.

Scope:
- Pick one canonical exported constructor for `compression` and one for `timeout`.
- Remove the losing wrapper functions and update tests/docs in the same change.
- Keep middleware behavior unchanged apart from the constructor/API convergence.
- Sync middleware docs and manifest language to the explicit single-constructor rule.

Non-goals:
- Do not redesign compression or timeout internals.
- Do not add compatibility aliases or deprecated wrappers.
- Do not widen middleware scope beyond constructor convergence.

Files:
- `middleware/compression/gzip.go`
- `middleware/timeout/timeout.go`
- `middleware/module.yaml`
- `docs/modules/middleware/README.md`

Tests:
- `go test -timeout 20s ./middleware/compression ./middleware/timeout`
- `go test -race -timeout 60s ./middleware/compression ./middleware/timeout`
- `go vet ./middleware/compression ./middleware/timeout`

Docs Sync:
- Keep middleware docs aligned on the rule that configurable middleware packages expose one explicit constructor path only.

Done Definition:
- `middleware/compression` has one exported constructor path.
- `middleware/timeout` has one exported constructor path.
- No residual call sites reference the removed wrapper names.
- Middleware docs describe the converged constructor surface.

Outcome:
- Completed.
- `middleware/compression` now converges on `Gzip(GzipConfig)` as its only exported constructor path.
- `middleware/timeout` now converges on `Timeout(TimeoutConfig)` as its only exported constructor path.
- Removed `GzipWithConfig` and the duration-only `Timeout(...)` wrapper shape, and updated tests/docs to the explicit config constructors.
