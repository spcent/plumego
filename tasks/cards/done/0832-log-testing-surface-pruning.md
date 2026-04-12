# Card 0832

Priority: P1
State: done
Primary Module: log
Owned Files:
- `log/testing.go`
- `log/logger.go`
- `log/module.yaml`
- `docs/modules/log/README.md`
- `x/observability/testlog`

Goal:
- Remove repo-wide testkit ownership from stable `log` so the stable module only owns runtime logging contracts and base implementations.
- Converge test logging to one non-stable home instead of keeping assertion-oriented helpers inside the stable runtime package.

Problem:
- `log/testing.go` exports `TestLogger`, `LogEntry`, `NewTestLogger`, and assertion-style helpers such as `Entries`, `CountByLevel`, `HasEntry`, and `LastEntry`.
- `log/module.yaml` and `docs/modules/log/README.md` describe stable `log` as a small runtime logging package, but the code still ships a repo-wide in-memory test logger surface.
- `TestLogger` also reaches into `contract.RequestIDFromContext`, which means the stable runtime module is still mixing testkit behavior with request-correlation assertions.
- Leaving this surface in `log` keeps the stable package larger and less canonical than its current documented boundary.

Scope:
- Remove exported testing-only logger helpers from stable `log`.
- Move the in-memory capture logger and assertion-oriented helpers to a clearly owned non-stable package under `x/observability`.
- Update all downstream tests that use stable `log` test helpers in the same change.
- Keep the runtime `StructuredLogger` contract and production-safe base implementations in stable `log`.
- Sync the log manifest and primer to the reduced stable boundary.

Non-goals:
- Do not redesign runtime logger interfaces or log field semantics.
- Do not add any third-party testing or logging dependency.
- Do not keep compatibility aliases or wrapper constructors in stable `log`.
- Do not move production logger implementations out of the stable root.

Files:
- `log/testing.go`
- `log/logger.go`
- `log/module.yaml`
- `docs/modules/log/README.md`
- `x/observability/testlog`

Tests:
- `go test -timeout 20s ./log/... ./x/observability/... ./core ./middleware/...`
- `go test -race -timeout 60s ./log/... ./x/observability/... ./core ./middleware/...`
- `go vet ./log/... ./x/observability/... ./core ./middleware/...`

Docs Sync:
- Keep the log manifest and primer aligned on the rule that stable `log` owns runtime logging contracts and base implementations only, while reusable test logging helpers live outside the stable root.

Done Definition:
- Stable `log` no longer exports `TestLogger`, `LogEntry`, or assertion-oriented testing helpers.
- A single non-stable package owns the repo-wide in-memory test logger surface.
- Downstream tests compile against the relocated helper surface with no residual references to deleted stable log test APIs.
- Module docs and manifest describe the same reduced stable `log` boundary that the code implements.

Outcome:
- Completed.
- Moved stable logging testkit types and helpers out of `log` into `x/observability/testlog`.
- Stable `log` no longer owns repo-wide test helper surfaces.
