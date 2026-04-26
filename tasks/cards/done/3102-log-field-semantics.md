# Card 3102: Log Field Semantics Unification

Milestone: none
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P1
State: done
Primary Module: log
Owned Files:
- `log/logger.go`
- `log/fields.go`
- `log/json.go`
- `log/logger_semantics_test.go`
- `log/json_test.go`
- `docs/modules/log/README.md`
Depends On:
- `tasks/cards/done/3101-log-callsite-depth.md`

Goal:
Make structured field handling consistent and unambiguous across text and JSON
loggers.

Problem:
- Logging methods accept `fields ...Fields`, but only the first element was
  used; later field maps were silently discarded.
- Text logs rendered fields with `%v`, so spaces, newlines, equals signs, and
  odd keys could produce ambiguous or multi-line structured suffixes.
- The merge/override rule was only implicit in tests, not documented as part of
  the stable log contract.

Scope:
- Merge every variadic `Fields` argument in order, with later maps overriding
  earlier maps.
- Keep base logger fields lower precedence than per-entry fields.
- Render text fields deterministically with escaping for ambiguous keys or
  values.
- Update focused tests and docs.

Non-goals:
- Do not add secret redaction or policy-specific field handling in stable `log`.
- Do not change JSON reserved key ownership beyond existing behavior.
- Do not change the public method signatures.

Outcome:
- Added `mergeFieldArgs` / `mergeFieldSets` and routed text and JSON log calls
  through the same variadic merge rule.
- Changed text field formatting to quote ambiguous keys and values while keeping
  deterministic key sorting.
- Added text and JSON tests for multi-map merge precedence.
- Added text tests for space, newline, and equals escaping.
- Documented field precedence and escaping in `docs/modules/log/README.md`.

Validation:
- `go test -race -timeout 60s ./log/...`
- `go test -timeout 20s ./log/...`
- `go vet ./log/...`

Done Definition:
- No variadic field map is silently dropped.
- Text field suffixes stay one-line and deterministic for special characters.
- JSON and text loggers share the same merge precedence.
- The listed validation commands pass.
