# Card 3103: JSON Logger Encoding Resilience

Milestone: none
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P2
State: active
Primary Module: log
Owned Files:
- `log/json.go`
- `log/json_test.go`
- `docs/modules/log/README.md` only if behavior is documented
Depends On:
- `tasks/cards/done/3102-log-field-semantics.md`

Goal:
Preserve the core JSON log entry when caller-supplied field values cannot be
encoded directly by `encoding/json`.

Problem:
`jsonLogger.log` marshals the whole entry at once. If one caller field is a
function, channel, invalid number, or otherwise unsupported value, the logger
replaces the entire entry with a generic marshal-error line and drops the
original level, message, timestamp, and safe fields.

Scope:
- Add a standard-library-only normalization path for caller fields before the
  final marshal.
- Preserve `time`, `level`, and `msg` as logger-owned keys.
- Preserve safe fields and stringify unsupported field values.
- Add tests for unsupported field values and reserved key ownership.

Non-goals:
- Do not introduce a schema registry or observability export backend.
- Do not implement secret detection or redaction.
- Do not add dependencies.

Tests:
- `go test -race -timeout 60s ./log/...`
- `go test -timeout 20s ./log/...`
- `go vet ./log/...`

Done Definition:
- Unsupported field values no longer collapse the whole JSON log entry.
- The original message and safe fields remain present.
- The listed validation commands pass.
