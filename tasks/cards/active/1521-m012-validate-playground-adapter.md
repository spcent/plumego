# Card 1521

Milestone: M-012
Recipe: specs/change-recipes/add-package.yaml
Priority: P1
State: active
Primary Module: x/validate
Owned Files:
- `x/validate/playground/adapter.go`
- `x/validate/playground/adapter_test.go`
- `x/validate/playground/go.mod`

Goal:
- Create x/validate/playground/ as a separately versioned sub-package that
  wraps go-playground/validator v10 and satisfies the x/validate.Validator
  interface, keeping the external dependency out of the main module.

Scope:
- Create x/validate/playground/go.mod with module
  github.com/spcent/plumego/x/validate/playground and dependency
  github.com/go-playground/validator/v10.
- Create x/validate/playground/adapter.go defining:
  - Validator struct wrapping validate.Validate
  - NewValidator(opts ...Option) *Validator — constructor accepting tag namespace
    and custom validation function options
  - Validate(v any) error — delegates to go-playground/validator; maps
    ValidationErrors to a structured field error list
  - Option functional option type
- Create x/validate/playground/adapter_test.go covering:
  - struct tag validation passes for valid input
  - required field missing returns error with field name in details
  - email format tag fails with descriptive message
  - min/max tag violations return per-field errors
  - NewValidator with custom function option registers correctly

Non-goals:
- Do not modify x/validate/go.mod to add go-playground/validator.
- Do not add playground/ to the main module dependency graph.
- Do not implement custom validation tags; use go-playground/validator native tags.

Files:
- `x/validate/playground/adapter.go`
- `x/validate/playground/adapter_test.go`
- `x/validate/playground/go.mod`

Tests:
- `go test -race -timeout 60s ./x/validate/playground/...`
- `go vet ./x/validate/playground/...`

Docs Sync:
- none at this card; usage example added in card 1522.

Done Definition:
- x/validate/playground/go.mod is a separate module with go-playground/validator.
- NewValidator() returns a Validator that satisfies x/validate.Validator.
- All five adapter_test.go test cases pass with `go test -race`.
- x/validate/go.mod does NOT contain go-playground/validator.

Outcome:
-
