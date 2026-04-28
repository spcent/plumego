# Card 0665

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P1
State: done
Primary Module: internal/config
Owned Files: internal/config/validator.go, internal/config/config_test.go
Depends On:

Goal:
Remove string-matching from legacy `ConfigSchema.Validate` required-field detection.

Scope:
- Detect `*Required` / `Required` validators by type instead of searching error text for "required".
- Preserve behavior for existing required validators and present-field validation errors.
- Add a focused test proving a custom validator error containing "required" does not fire for a missing optional field.

Non-goals:
- Do not change `ConfigSchemaManager.ValidateAll`.
- Do not redesign legacy schema APIs.

Files:
- internal/config/validator.go
- internal/config/config_test.go

Tests:
- go test -timeout 20s ./internal/config
- go test -timeout 20s ./internal/...
- go vet ./internal/...

Docs Sync:
Not required; internal validation behavior only.

Done Definition:
- Missing-field errors are emitted only for actual `Required` validators.
- Tests cover the brittle-string regression case.

Outcome:
Done. `ConfigSchema.Validate` now detects missing-field required errors by
validator type instead of matching the word "required" in the error text.
Added regression coverage for a custom optional validator whose error message
contains "required".

Validation:
- go test -timeout 20s ./internal/config
- go test -timeout 20s ./internal/...
- go vet ./internal/...
