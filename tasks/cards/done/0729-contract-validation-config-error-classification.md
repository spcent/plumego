# Card 0729

Milestone: v1
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P1
State: done
Primary Module: contract
Owned Files:
- contract/validation.go
- contract/bind_helpers.go
- contract/*_test.go
- docs/modules/contract/README.md
Depends On:
- 0728

Goal:
Prevent validation rule configuration errors from being reported as client payload validation failures.

Scope:
- Introduce a recognizable validation configuration error sentinel or wrapper in `contract`.
- Wrap unknown and misconfigured validation rules with that classification.
- Map validation configuration errors to server-side `APIError` defaults in `BindErrorToAPIError`.
- Add regression coverage for unknown validation rules and invalid min/max rules through `WriteBindError`.

Non-goals:
- Do not expand the validation rule DSL.
- Do not change normal `ValidationErrors` field output.
- Do not hide programmer configuration errors from tests.

Files:
- contract/validation.go
- contract/bind_helpers.go
- contract/active_cards_regression_test.go
- contract/bind_helpers_test.go
- docs/modules/contract/README.md

Tests:
- go test -timeout 20s ./contract/...
- go test -race -timeout 60s ./contract/...
- go run ./internal/checks/dependency-rules

Docs Sync:
- Document that validation rule configuration errors are server-side programmer errors, not field validation errors.

Done Definition:
- `errors.Is(err, ErrValidationConfig)` works for unknown and malformed validation rules.
- `WriteBindError` no longer maps validation configuration errors to 400 validation failures.
- Contract tests and dependency checks pass.

Outcome:
- Added `ErrValidationConfig` for unknown and malformed validation rules.
- Wrapped unknown rules and invalid `min`/`max` rule configurations with `ErrValidationConfig`.
- Mapped validation configuration errors to `500`/`INTERNAL_ERROR` in `BindErrorToAPIError`.
- Added regression coverage for direct validation errors and `WriteBindError` conversion.

Validation:
- go test -timeout 20s ./contract/...
- go test -race -timeout 60s ./contract/...
- go run ./internal/checks/dependency-rules
