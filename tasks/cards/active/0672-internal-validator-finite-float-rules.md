# Card 0672

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P1
State: active
Primary Module: internal/validator
Owned Files: internal/validator/rules.go, internal/validator/validator_test.go
Depends On:

Goal:
Reject non-finite float values in internal validator numeric rules.

Scope:
- Make `Numeric`, `Float`, `MinFloat`, `MaxFloat`, and `RangeFloat` reject `NaN` and infinities.
- Apply the same finite check to string-parsed float values.
- Add focused tests for `NaN`, `Inf`, and `"NaN"` false positives.

Non-goals:
- Do not change integer validation rules.
- Do not redesign validation error codes or messages.

Files:
- internal/validator/rules.go
- internal/validator/validator_test.go

Tests:
- go test -timeout 20s ./internal/validator
- go test -timeout 20s ./internal/...
- go vet ./internal/...

Docs Sync:
Not required; internal validation hardening only.

Done Definition:
- Non-finite float values cannot satisfy float or range validators.
- Existing finite numeric values continue to pass.

Outcome:

