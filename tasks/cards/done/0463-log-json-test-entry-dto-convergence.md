# Card 0463: Log JSON Test Entry DTO Convergence

Milestone: none
Recipe: specs/change-recipes/http-endpoint-bugfix.yaml
Priority: P1
State: active
Primary Module: log
Owned Files:
- `log/json_test.go`
Depends On: none

Goal:
Consolidate JSON logger tests around a typed log-entry DTO and shared decode
helper.

Problem:
`log/json_test.go` repeatedly decodes logger output into `map[string]any` and
indexes the same stable fields by string. This duplicates decode boilerplate and
makes the expected log entry shape less obvious.

Scope:
- Add a local typed JSON log entry struct and helper for tests.
- Replace repeated fixed-field map decodes in JSON logger tests.
- Keep concurrency JSON-validity checks simple and behavior-preserving.

Non-goals:
- Do not change logger behavior or public APIs.
- Do not remove support for arbitrary caller fields in production logging.
- Do not add dependencies.

Files:
- `log/json_test.go`

Tests:
- `go test -race -timeout 60s ./log/...`
- `go test -timeout 20s ./log/...`
- `go vet ./log/...`

Docs Sync:
No docs change required; this is test-only cleanup.

Done Definition:
- JSON logger tests no longer repeat fixed-field `map[string]any` decoding.
- The listed validation commands pass.

Outcome:
