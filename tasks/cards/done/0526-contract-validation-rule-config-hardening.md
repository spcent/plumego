# Card 0526

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P2
State: done
Primary Module: contract
Owned Files:
- contract/validation.go
- contract/active_cards_regression_test.go
Depends On: 2235

Goal:
Make malformed `min` and `max` validation rule limits fail consistently instead of producing type-dependent behavior.

Scope:
- Treat negative `min` / `max` limits as validator configuration failures.
- Keep failures as field-level `CodeInvalidFormat` validation errors, matching existing malformed numeric config behavior.
- Add regression coverage for negative limits on signed, unsigned, string, and float fields.

Non-goals:
- Do not add new validation tags.
- Do not change successful `required`, `email`, `min`, or `max` semantics.
- Do not turn misconfigured limits into programmer errors.

Files:
- `contract/validation.go`
- `contract/active_cards_regression_test.go`

Tests:
- `go test -timeout 20s ./contract/...`
- `go vet ./contract/...`

Docs Sync:
- None expected; this clarifies existing validation config behavior.

Done Definition:
- Negative `min` and `max` limits return `CodeInvalidFormat` field errors.
- Existing validation behavior tests continue to pass.

Outcome:
- Rejected negative `min` and `max` limits as invalid validator configuration.
- Added coverage across signed, unsigned, string, and float fields.
- Validation run: `go test -timeout 20s ./contract/...`; `go vet ./contract/...`.
