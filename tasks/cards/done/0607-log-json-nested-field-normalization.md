# Card 0607: JSON Nested Field Normalization

Milestone: none
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P2
State: done
Primary Module: log
Owned Files:
- `log/json.go`
- `log/json_test.go`
- `docs/modules/log/README.md`
Depends On:
- `tasks/cards/done/0602-log-json-resilience.md`

Goal:
Preserve safe nested JSON fields when a sibling nested value is unsupported.

Problem:
The JSON fallback normalized each top-level field with a full
`json.Marshal`/`json.Unmarshal` round trip. If a `map[string]any` field
contained one unsupported nested value, the whole nested object was stringified
and safe sibling fields were lost.

Scope:
- Normalize common nested structured values recursively for `Fields`,
  `map[string]any`, `map[string]string`, and `[]any`.
- Stringify unsupported leaf values instead of stringifying the whole parent.
- Keep logger-owned reserved keys unchanged.
- Add focused tests for nested maps and arrays.

Non-goals:
- Do not build a general schema system.
- Do not add reflection-heavy redaction or policy behavior.
- Do not change text logging.

Outcome:
- Added recursive normalization for common nested structured field values.
- Preserved safe nested map and slice values when sibling values are
  unsupported.
- Added focused tests for nested maps, arrays, and unsupported nested leaves.
- Updated module docs to describe nested sibling preservation.

Validation:
- `go test -race -timeout 60s ./log/...`
- `go test -timeout 20s ./log/...`
- `go vet ./log/...`

Done Definition:
- Safe nested JSON siblings survive unsupported sibling values.
- Unsupported nested leaves are stringified.
- The listed validation commands pass.
