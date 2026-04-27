# Card 4103: JSON Nested Field Normalization

Milestone: none
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P2
State: active
Primary Module: log
Owned Files:
- `log/json.go`
- `log/json_test.go`
- `docs/modules/log/README.md`
Depends On:
- `tasks/cards/done/3103-log-json-resilience.md`

Goal:
Preserve safe nested JSON fields when a sibling nested value is unsupported.

Problem:
The current JSON fallback normalizes each top-level field with a full
`json.Marshal`/`json.Unmarshal` round trip. If a `map[string]any` field contains
one unsupported nested value, the whole nested object is stringified and safe
sibling fields are lost.

Scope:
- Normalize common nested structured values recursively for `map[string]any`,
  `map[string]string`, and `[]any`.
- Stringify unsupported leaf values instead of stringifying the whole parent.
- Keep logger-owned reserved keys unchanged.
- Add focused tests for nested maps and arrays.

Non-goals:
- Do not build a general schema system.
- Do not add reflection-heavy redaction or policy behavior.
- Do not change text logging.

Tests:
- `go test -race -timeout 60s ./log/...`
- `go test -timeout 20s ./log/...`
- `go vet ./log/...`

Done Definition:
- Safe nested JSON siblings survive unsupported sibling values.
- Unsupported nested leaves are stringified.
- The listed validation commands pass.
