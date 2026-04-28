# Card 0670

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P1
State: active
Primary Module: internal/config
Owned Files: internal/config/convert.go, internal/config/config_test.go
Depends On:

Goal:
Prevent internal config integer conversion from wrapping overflowing numeric values.

Scope:
- Make `toInt` return the supplied default for integer and float values that cannot fit in `int`.
- Preserve existing behavior for valid in-range numeric, boolean, and string values.
- Add focused tests for overflowing unsigned integers and non-finite floats.

Non-goals:
- Do not change struct unmarshalling overflow behavior.
- Do not change `toFloat64` or `toBool` semantics.

Files:
- internal/config/convert.go
- internal/config/config_test.go

Tests:
- go test -timeout 20s ./internal/config
- go test -timeout 20s ./internal/...
- go vet ./internal/...

Docs Sync:
Not required; internal conversion hardening only.

Done Definition:
- `toInt` does not wrap overflowing integer values.
- Tests cover overflow fallback and normal conversion preservation.

Outcome:

