# Card 0839

Priority: P1
State: done
Primary Module: log
Owned Files:
- `log/logger.go`
- `log/glog.go`
- `log/json.go`
- `log/noop.go`
- `log/module.yaml`
- `docs/modules/log/README.md`
- `core`
- `middleware`

Goal:
- Shrink stable `log` to one canonical logger surface centered on `StructuredLogger` and `NewLogger`.
- Remove transport-contract coupling from stable `log` so logging no longer imports `contract` or auto-reads request metadata from request contexts.

Problem:
- `log/module.yaml` and `docs/modules/log/README.md` describe a compact stable surface, but the package still exports multiple concrete implementations and constructors (`Logger`/`New`, `JSONLogger`/`NewJSONLogger`, `NoOpLogger`/`NewNoOpLogger`) in addition to `StructuredLogger` and `NewLogger`.
- `log/logger.go` and `log/json.go` import `contract` to auto-extract `request_id` from context, which violates the log manifest's intended stdlib-only dependency shape and couples logging to transport metadata contracts.
- The default text logger and JSON logger currently apply different verbosity semantics and context enrichment behavior, which means the stable logging contract is no longer “one canonical implementation path”.
- `core` and middleware packages already attach structured fields explicitly in many call sites, so keeping context-driven field extraction in stable `log` is redundant and muddies ownership.

Scope:
- Decide the minimum stable logger surface and remove redundant exported concrete constructors or implementations that are not part of the canonical path.
- Remove `contract` imports from stable `log`; request correlation fields must be attached by callers instead of logger internals.
- Converge debug/verbosity behavior so the canonical logger path has one predictable rule set.
- Update `core`, `middleware`, and any remaining stable call sites that relied on logger-side context extraction.
- Sync log docs and manifest to the converged surface.

Non-goals:
- Do not add a compatibility alias layer for removed logger constructors.
- Do not move feature-specific logging schemas into stable `log`.
- Do not reintroduce request-id generation or transport observability policy into stable `log`.
- Do not add exporter or sink integrations to stable `log`.

Files:
- `log/logger.go`
- `log/glog.go`
- `log/json.go`
- `log/noop.go`
- `log/module.yaml`
- `docs/modules/log/README.md`
- `core`
- `middleware`

Tests:
- `go test -timeout 20s ./log/... ./core ./middleware/...`
- `go test -race -timeout 60s ./log/... ./core ./middleware/...`
- `go vet ./log/... ./core ./middleware/...`

Docs Sync:
- Keep the log manifest and primer aligned on the rule that stable `log` owns a minimal structured logger contract and one canonical implementation path, not transport-aware context enrichment or multiple parallel logger families.

Done Definition:
- Stable `log` no longer imports `contract` or auto-extracts request metadata from contexts.
- Redundant exported logger constructors and concrete surfaces are removed or internalized so the public stable path is clearly `StructuredLogger` + `NewLogger`.
- Core and middleware callers compile against explicit field injection with no residual reliance on logger-side context enrichment.
- Log docs and manifest describe the same reduced public surface that the code implements.

Outcome:
- Completed. Stable `log` now exposes one canonical constructor path through `NewLogger(LoggerConfig)`, no longer imports `contract`, and no longer auto-extracts request metadata from contexts.
