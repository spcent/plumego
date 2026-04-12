# Card 0846

Priority: P1
State: done
Primary Module: log
Owned Files:
- `log/logger.go`
- `log/glog.go`
- `log/noop.go`
- `log/module.yaml`
- `docs/modules/log/README.md`
- `core`

Goal:
- Remove hidden global backend ownership from stable `log` so the canonical path is constructor-owned logger instances only.
- Eliminate orphaned lifecycle/start-stop semantics that are not actually driven by the stable app kernel.

Problem:
- Stable `log` now presents `StructuredLogger` plus `NewLogger` as the canonical public surface, but the text backend still relies on a package-global `std`, package-level flag state, and internal default helpers such as `initDefaultFromFlags`, `flushDefault`, and `closeDefault`.
- `defaultLogger.Start` and `defaultLogger.Stop` still implement the exported `Lifecycle` interface by driving that hidden global backend, even though `core` does not own or drive logger lifecycle in the normal application path.
- This leaves stable `log` with two competing ownership models: constructor-created logger values on the surface, and a shared singleton backend under the hood.
- Most remaining references to the default singleton path are now package-local tests, which indicates the runtime API has already converged but the implementation model has not.

Scope:
- Converge text logger ownership on constructor-created backend state rather than a package-global singleton.
- Remove or internalize logger lifecycle hooks and singleton helpers that are no longer part of the canonical stable path.
- Decide whether any CLI flag bootstrap belongs in stable `log`; if not, move or delete it in the same change.
- Update `core` tests and `log` tests to match the converged constructor-owned backend model.
- Sync log docs and manifest to the final backend ownership story.

Non-goals:
- Do not remove `StructuredLogger` or the canonical `NewLogger` constructor.
- Do not remove JSON or discard format selection from `LoggerConfig`.
- Do not add app-specific bootstrap code or observability exporter wiring to stable `log`.
- Do not reintroduce compatibility wrappers around removed singleton helpers.

Files:
- `log/logger.go`
- `log/glog.go`
- `log/noop.go`
- `log/module.yaml`
- `docs/modules/log/README.md`
- `core`

Tests:
- `go test -timeout 20s ./log/... ./core`
- `go test -race -timeout 60s ./log/... ./core`
- `go vet ./log/... ./core`

Docs Sync:
- Keep the log manifest and primer aligned on the rule that stable `log` exposes constructor-owned logger instances only, with no hidden global backend or orphaned lifecycle path.

Done Definition:
- Stable `log` no longer depends on a package-global backend singleton for the canonical text logger path.
- The exported stable surface no longer includes lifecycle semantics or singleton helpers that the app kernel does not actually own.
- `log` and `core` tests compile against the converged constructor-owned backend model with no residual references to removed singleton helpers.
- Log docs and manifest describe the same backend ownership model the code implements.

Outcome:
- Completed.
- Moved the canonical text logger path to constructor-owned `gLogger` instances instead of the package-global default backend.
- Removed the exported logger `Lifecycle` surface and dropped runtime `Start` / `Stop` ownership from stable `log`.
- Removed package-global default bootstrap helpers from production code and kept the old singleton-oriented utilities only in package-local test helpers used by `log/glog_test.go`.
- Synced log docs and manifest to the rule that stable `log` exposes constructor-owned instances only, with no hidden singleton or lifecycle path in runtime code.
