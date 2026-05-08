# Card 0745

Milestone: v1
Recipe: specs/change-recipes/stable-root-boundary-review.yaml
Priority: P1
State: done
Primary Module: contract
Owned Files:
- contract/validation.go
- contract/active_cards_regression_test.go
- docs/modules/contract/README.md
Depends On:
- 0744

Goal:
Make `ValidateStruct` stable compatibility semantics explicit and executable for disputed edge cases.

Scope:
- Add focused tests for nil/non-struct inputs, required zero values, and depth overflow behavior.
- Clarify that `ValidateStruct` is a compatibility validator with narrow tag rules.
- Keep rule support unchanged unless a test reveals behavior drift.

Non-goals:
- Do not add new validation rules.
- Do not add localization, cross-field validation, or business policy checks.
- Do not move validation ownership in this card.

Files:
- contract/validation.go
- contract/active_cards_regression_test.go
- docs/modules/contract/README.md

Tests:
- go test -timeout 20s ./contract/...
- go vet ./contract/...

Docs Sync:
- Update contract docs for exact validation edge semantics.

Done Definition:
- The disputed edge cases are locked by tests.
- Docs explain compatibility behavior and ownership limits.
- Contract tests and vet pass.

Outcome:
- Added compatibility edge coverage for nil inputs, non-struct values, nil non-struct pointers, and `required` zero-value semantics.
- Tightened depth-limit coverage to assert the out-of-range field behavior.
- Documented `ValidateStruct` no-op inputs, zero-value required behavior, and depth-limit behavior.

Validation:
- go test -timeout 20s ./contract/...
- go vet ./contract/...
