# Card 4104: JSON Derived Logger Write-Error Reporting

Milestone: none
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P3
State: done
Primary Module: log
Owned Files:
- `log/json.go`
- `log/json_test.go`
Depends On:
- `tasks/cards/done/3103-log-json-resilience.md`

Goal:
Avoid repeated JSON write-error reports from derived loggers that share the same
writer.

Problem:
`jsonLogger.WithFields` shared the output writer and write mutex with the base
logger, but each child received a fresh `sync.Once` for write-error reporting.
For a broken shared writer, every derived logger could report the same output
failure once, while the text backend reports writer failures once per backend.

Scope:
- Share the JSON write-error `sync.Once` across derived loggers.
- Keep write serialization unchanged.
- Add focused tests for shared derived logger error state.

Non-goals:
- Do not change where write errors are reported.
- Do not add logging hooks or observability exporters.
- Do not change `StructuredLogger`.

Outcome:
- Changed JSON write-error suppression state to a shared `*sync.Once`.
- Shared that state through `WithFields` alongside the shared writer lock.
- Added a focused test for derived logger error-state sharing.

Validation:
- `go test -race -timeout 60s ./log/...`
- `go test -timeout 20s ./log/...`
- `go vet ./log/...`

Done Definition:
- Derived JSON loggers share write-error suppression state with their base.
- Existing JSON logging behavior remains unchanged.
- The listed validation commands pass.
