# Card 2255

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P1
State: done
Primary Module: contract
Owned Files:
- contract/validation.go
- contract/active_cards_regression_test.go
Depends On: 2254

Goal:
Make validation rules consistently unwrap pointers and interfaces.

Scope:
- Unwrap interface values before applying `required`, `email`, `min`, and `max`.
- Support nested pointer chains where the eventual value is supported.
- Preserve nil pointer/interface optional behavior outside `required`.
- Add focused regression coverage.

Non-goals:
- Do not add new validation rules.
- Do not validate map values.
- Do not change unsupported type policy beyond indirect value handling.

Files:
- `contract/validation.go`
- `contract/active_cards_regression_test.go`

Tests:
- `go test -timeout 20s ./contract/...`
- `go vet ./contract/...`

Docs Sync:
- None expected; this completes existing rule behavior.

Done Definition:
- Interface-held strings and multi-pointer strings validate the same as direct strings.
- Required nil interface/pointer values still fail.

Outcome:
- Validation now unwraps pointers and interfaces before applying scalar rules.
- Nested validation also follows interface-wrapped structs.
- Added coverage for required, email, min, and nested struct behavior through indirect values.
