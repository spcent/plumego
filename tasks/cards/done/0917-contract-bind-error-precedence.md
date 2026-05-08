# Card 0917

Milestone: v1
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P1
State: done
Primary Module: contract
Owned Files:
- contract/bind_helpers.go
- contract/bind_helpers_test.go
Depends On:
- 0733

Goal:
Make binding error classification fail closed so server/configuration errors cannot be downgraded to validation errors by attached field details.

Scope:
- Review `BindErrorToAPIError` precedence after `TypeOnly` removal.
- Ensure `ErrValidationConfig` and other server/configuration errors keep server status/category/code even if the error chain also exposes `ValidationErrors`.
- Add a focused regression test for mixed configuration and field errors.

Non-goals:
- Do not redesign validation field structures.
- Do not change ordinary validation failures from 400 responses.
- Do not alter `WriteBindError` response envelope.

Files:
- contract/bind_helpers.go
- contract/bind_helpers_test.go

Tests:
- go test -timeout 20s ./contract/...
- go vet ./contract/...
- go run ./internal/checks/module-manifests

Docs Sync:
- None expected unless implementation changes documented behavior.

Done Definition:
- Server/configuration bind errors retain 5xx classification when mixed with field validation details.
- Ordinary validation errors still return validation fields.
- Targeted tests, vet, and manifest checks pass.

Outcome:
- Limited field-validation override behavior to errors that are not classified as validation configuration errors.
- Kept `ErrValidationConfig` responses on 500/internal metadata even when the error chain also contains `ValidationErrors`.
- Added regression coverage for a joined configuration error plus field validation details.

Validation:
- go test -timeout 20s ./contract/...
- go vet ./contract/...
- go run ./internal/checks/module-manifests
