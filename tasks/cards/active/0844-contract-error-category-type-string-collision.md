# Card 0844

Priority: P1
State: active
Primary Module: contract
Owned Files:
- `contract/errors.go`
- `contract/error_codes.go`
Depends On:

Goal:
- Eliminate the string value collision between `ErrorCategory` and `ErrorType` constants so that a raw string in a log, metric label, or JSON payload unambiguously identifies which field it came from.

Problem:
- `ErrorCategory` and `ErrorType` are distinct semantic concepts: category is a broad class (client/server/auth/…), type is a specific sub-variant within that class.
- Two pairs of constants share identical string values:
  - `CategoryValidation ErrorCategory = "validation_error"` and `TypeValidation ErrorType = "validation_error"` (`errors.go:29`, `errors.go:52`)
  - `CategoryTimeout ErrorCategory = "timeout_error"` and `TypeTimeout ErrorType = "timeout_error"` (`errors.go:27`, `errors.go:72`)
- When `"validation_error"` appears in a log field, a Prometheus label, or a JSON payload, there is no way to determine from the value alone whether it originated from the `category` or `type` field.
- The JSON output of `APIError` includes both `"category"` and `"type"` fields, but their values may be identical, which provides no additional signal.
- The collision also makes it harder to write precise metric filters or alert rules that distinguish category aggregates from type-level specifics.

Scope:
- Rename the two colliding `ErrorType` constants to use a prefix or suffix that distinguishes them from their `ErrorCategory` counterparts:
  - `TypeValidation` → keep as-is; change string value to `"validation_failure"` (or similar non-colliding value)
  - `TypeTimeout` → keep as-is; change string value to `"timeout_failure"` (or similar)
- Update `errorTypeLookup` entries for both types to use the new string values.
- Update any tests that assert against the old string values.
- No change to `ErrorCategory` constants (they are the correct ground-truth labels for their dimension).

Non-goals:
- Do not rename the `TypeValidation` or `TypeTimeout` Go identifiers; only their string values change.
- Do not change any `ErrorCategory` constant values.
- Do not change error codes in `error_codes.go`.
- Do not change `APIError` JSON field names.

Files:
- `contract/errors.go`
- `contract/errors_test.go`
- `contract/error_handling_test.go`

Tests:
- `go test -timeout 20s ./contract/...`
- `go vet ./contract/...`

Docs Sync:
- None required.

Done Definition:
- No `ErrorType` constant string value matches any `ErrorCategory` constant string value.
- `errorTypeLookup` and all tests are updated to use the new string values.
- All tests pass.

Outcome:
- Pending.
