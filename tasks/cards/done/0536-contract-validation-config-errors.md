# Card 0536

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P1
State: done
Primary Module: contract
Owned Files:
- contract/validation.go
- contract/active_cards_regression_test.go
Depends On: 2245

Goal:
Keep malformed validation rule configuration separate from user input validation failures.

Scope:
- Return programmer errors for malformed `min=` and `max=` validator configuration.
- Keep unknown-rule behavior as a programmer error.
- Add focused coverage proving malformed rule config is not surfaced as `ValidationErrors`.

Non-goals:
- Do not add new validation rules.
- Do not change valid `min` or `max` behavior.
- Do not add business-domain validation.

Files:
- `contract/validation.go`
- `contract/active_cards_regression_test.go`

Tests:
- `go test -timeout 20s ./contract/...`
- `go vet ./contract/...`

Docs Sync:
- None expected; this aligns implementation with existing validation comments.

Done Definition:
- Invalid validator configuration is distinguishable from user data validation.
- Existing validation tests continue to pass.

Outcome:
- Malformed `min=` and `max=` rules now return programmer errors instead of `ValidationErrors`.
- Updated regression coverage to assert malformed rule config is not exposed as field-level user input validation.
