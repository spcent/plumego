# Card 0828

Priority: P1
State: done
Primary Module: log
Owned Files:
- `log/logger.go`
- `log/glog.go`
- `log/traceid.go`
- `log/module.yaml`
- `docs/modules/log/README.md`
Depends On:
- `0822-observability-id-contract-convergence.md`

Goal:
- Shrink stable `log` to one clear structured-logging surface and remove request-id generation ownership from the logging module.
- Converge directly to a single canonical stable logger model instead of keeping overlapping logger families and global compatibility APIs.

Problem:
- `log/module.yaml` and `docs/modules/log/README.md` describe a small stable logging package, but the code still exports multiple overlapping families: `StructuredLogger` + `NewGLogger`, package-global glog functions (`Init`, `Info`, `Error`, `V`, `CopyStandardLogTo`, etc.), JSON/no-op/test logger constructors, and lifecycle hooks.
- `log/traceid.go` still owns request-id generation (`NewRequestID`, `NewRequestIDGenerator`, `DecodeRequestID`), and stable middleware currently depends on that surface even though request correlation is already canonically owned by `contract` + `middleware/requestid`.
- This leaves `log` as both a stable runtime contract package and a legacy/global logging subsystem, with request-correlation generation mixed into logger ownership.

Scope:
- Pick one canonical stable logger surface and remove or relocate the losing public families.
- Remove request-id generation from stable `log` and move it to the canonical request-id owner.
- Eliminate package-global logger compatibility APIs from the stable root if they are not part of the canonical structured logger model.
- Update stable callers to the converged surface in the same change.
- Sync the log module manifest and primer to the reduced public boundary.

Non-goals:
- Do not add any third-party logging dependency.
- Do not redesign log field naming or structured field semantics beyond what is necessary to converge the public surface.
- Do not move logger initialization into `core` or hide it behind app lifecycle side effects.
- Do not preserve legacy logging entrypoints for compatibility.

Files:
- `log/logger.go`
- `log/glog.go`
- `log/traceid.go`
- `log/module.yaml`
- `docs/modules/log/README.md`

Tests:
- `go test -timeout 20s ./log/... ./middleware/requestid ./core ./x/data/sharding/...`
- `go test -race -timeout 60s ./log/... ./middleware/requestid ./core ./x/data/sharding/...`
- `go vet ./log/... ./middleware/requestid ./core ./x/data/sharding/...`

Docs Sync:
- Keep the log primer and manifest aligned on the rule that stable `log` owns a small structured logger contract and base implementations only, not global compatibility APIs or request-id generation.

Done Definition:
- Stable `log` exposes one canonical logger surface instead of multiple overlapping logger families.
- Request-id generation no longer lives in `log`.
- Legacy package-global logger compatibility APIs are removed or relocated out of the stable root.
- Stable callers use the converged surface with no residual references to deleted log APIs.
- Module docs and manifest describe the same reduced stable boundary the code implements.

Outcome:
- Completed.
- Converged stable `log` on one canonical logger constructor path and removed request-id generation ownership from the stable logging root.
- Request-correlation generation now lives with the request-observability owner path instead of the logger package.
