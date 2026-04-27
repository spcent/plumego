# Card 0537

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P1
State: done
Primary Module: contract
Owned Files:
- contract/bind_helpers.go
- contract/bind_helpers_test.go
- contract/context_test.go
Depends On: 2246

Goal:
Classify binding infrastructure failures consistently instead of reporting them as body read failures.

Scope:
- Map `ErrContextNil` and `ErrRequestNil` binding failures to `CodeInvalidRequest`.
- Preserve existing body-read failure behavior for real read errors.
- Add focused API-error mapping coverage.

Non-goals:
- Do not change `BindJSON` or `BindQuery` signatures.
- Do not change success binding behavior.
- Do not introduce new error codes.

Files:
- `contract/bind_helpers.go`
- `contract/bind_helpers_test.go`
- `contract/context_test.go`

Tests:
- `go test -timeout 20s ./contract/...`
- `go vet ./contract/...`

Docs Sync:
- None expected; this is an error-classification fix.

Done Definition:
- Nil context/request binding errors no longer emit `REQUEST_BODY_READ_FAILED`.
- Existing binding error mappings continue to pass.

Outcome:
- `BindErrorToAPIError` now maps nil context and nil request binding failures to `CodeInvalidRequest`.
- Added focused mapping tests while leaving genuine body read errors on `CodeRequestBodyReadFailed`.
