# Card 0657

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P1
State: done
Primary Module: internal/validator
Owned Files: internal/validator/validator.go, internal/validator/routeparam_validators.go, internal/validator/routeparam_validators_test.go, internal/validator/validator_detailed_test.go
Depends On:

Goal:
Make validator behavior more defensive and consistent for malformed registrations and route parameter length checks.

Scope:
- Ignore nil rules registered in `RuleRegistry` so malformed custom registry state cannot panic during validation.
- Make `RouteParamCompositeValidator` return an error for nil child validators instead of panicking.
- Count route parameter length in runes, not bytes, to match user-visible string length.
- Add focused tests for each behavior.

Non-goals:
- Do not change validation tag syntax.
- Do not remove panic-compatible helpers.
- Do not refactor the validator rule catalog.

Files:
- internal/validator/validator.go
- internal/validator/routeparam_validators.go
- internal/validator/routeparam_validators_test.go
- internal/validator/validator_detailed_test.go

Tests:
- go test -timeout 20s ./internal/validator
- go test -timeout 20s ./internal/...
- go vet ./internal/...

Docs Sync:
Not required; this is internal defensive behavior.

Done Definition:
- Nil registered rules and nil route parameter validators do not panic.
- Route parameter length validation handles multibyte characters as one character each.
- Focused and internal package tests pass.

Outcome:
Completed. Nil registered rules are ignored, nil route parameter composite children return errors, and route parameter length validation now counts runes.

Validation:
- go test -timeout 20s ./internal/validator
- go test -timeout 20s ./internal/...
- go vet ./internal/...
