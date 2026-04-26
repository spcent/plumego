# Card 3102: Log Field Semantics Unification

Milestone: none
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P1
State: active
Primary Module: log
Owned Files:
- `log/logger.go`
- `log/fields.go`
- `log/json.go`
- `log/*_test.go`
- `docs/modules/log/README.md`
Depends On:
- `tasks/cards/done/3101-log-callsite-depth.md`

Goal:
Make structured field handling consistent and unambiguous across text and JSON
loggers.

Problem:
- Logging methods accept `fields ...Fields`, but only the first element is used;
  later field maps are silently discarded.
- Text logs render fields with `%v`, so spaces, newlines, equals signs, and odd
  keys can produce ambiguous or multi-line structured suffixes.
- The merge/override rule is only implicit in tests, not documented as part of
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

Tests:
- `go test -race -timeout 60s ./log/...`
- `go test -timeout 20s ./log/...`
- `go vet ./log/...`

Done Definition:
- No variadic field map is silently dropped.
- Text field suffixes stay one-line and deterministic for special characters.
- JSON and text loggers share the same merge precedence.
- The listed validation commands pass.
