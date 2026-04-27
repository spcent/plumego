# Card 0423: Internal Validator Code Taxonomy
Milestone: none
Recipe: specs/change-recipes/fix-bug.yaml
Priority: medium
State: done
Primary Module: internal/validator
Owned Files:
- internal/validator/validator.go
- internal/validator/rules.go
- internal/validator/validator_test.go
- internal/validator/validator_detailed_test.go
Depends On: none

Goal:
Converge internal validator error codes onto one named taxonomy instead of
hundreds of ad hoc lowercase string literals. The package currently emits codes
such as `"required"`, `"email"`, `"min"`, and `"data"` inline throughout rule
implementations, making it easy for new rules to drift and hard to audit
validation behavior.

Scope:
- Introduce package-local validation code constants.
- Replace inline built-in rule code strings with those constants.
- Add tests that assert representative rules return the expected constants.
- Keep JSON field names and existing lowercase wire values unless a broader
  caller migration card explicitly changes the wire contract.

Non-goals:
- Do not rewrite validator rule semantics.
- Do not change `contract` validation behavior.
- Do not expand this card into HTTP binding or config-schema validation.

Files:
- `internal/validator/validator.go`
- `internal/validator/rules.go`
- `internal/validator/validator_test.go`
- `internal/validator/validator_detailed_test.go`

Tests:
- go test -race -timeout 60s ./internal/validator
- go test -timeout 20s ./internal/validator
- go vet ./internal/validator

Docs Sync:
No docs update required for package-local constants.

Done Definition:
- Built-in validator rule codes are defined through named constants.
- Representative validation tests assert code values through constants, not raw
  string literals.
- No public API or response envelope shape changes.
- The listed validation commands pass.

Outcome:
Added package-local validation code constants and mechanically replaced built-in
validator rule code literals in `validator.go` and `rules.go`. Added
representative tests for required, email, min, UUID, secure URL, and data-level
validation errors to assert codes through constants while preserving existing
wire values.

Validation:
- `go test -race -timeout 60s ./internal/validator`
- `go test -timeout 20s ./internal/validator`
- `go vet ./internal/validator`
- `rg -n "Code: \\\"" internal/validator/validator.go internal/validator/rules.go`
  returned no matches.
