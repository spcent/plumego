# Card 0637

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P1
State: done
Primary Module: internal/validator
Owned Files: internal/validator/rules.go, internal/validator/validator_detailed_test.go
Depends On: tasks/cards/done/0636-internal-nethttp-error-response-body-close.md

Goal:
Make regex-backed validation rules consistent and fail closed without repeated
per-constructor regexp compilation for shared built-ins.

Scope:
- Hoist repeated built-in regex patterns to package-level compiled values.
- Make `Regex` and `CaseInsensitive` return validation errors for invalid
  patterns instead of panicking during construction.
- Add focused tests for invalid regex specs.

Non-goals:
- Do not rename validation codes.
- Do not change route parameter validator constructors.
- Do not split the large rules file in this card.

Files:
- internal/validator/rules.go
- internal/validator/validator_detailed_test.go

Tests:
- go test ./internal/validator
- go test ./internal/...

Docs Sync:
- None; no external docs describe panic behavior.

Done Definition:
- Invalid regex patterns are reported as validation errors.
- Shared built-in regexes are compiled once.
- Existing validator behavior remains covered.

Outcome:
- Hoisted shared built-in regex patterns to package-level compiled values.
- Changed invalid `Regex` and `CaseInsensitive` patterns to return validation errors instead of panicking.
- Added direct and struct-tag coverage for invalid regex patterns.
- Validation: `go test ./internal/validator`; `go test ./internal/...`.
