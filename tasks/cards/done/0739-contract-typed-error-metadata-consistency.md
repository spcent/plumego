# Card 0739

Milestone: v1
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P1
State: done
Primary Module: contract
Owned Files:
- contract/errors.go
- contract/errors_test.go
- contract/freeze_test.go
- x/messaging/api.go
- x/messaging/api_test.go
- docs/modules/contract/README.md
Depends On:
- 0738

Goal:
Ensure typed errors cannot carry status/category metadata that contradicts the selected canonical `ErrorType`.

Scope:
- Normalize typed `APIError` values so status and category match `ErrorType.Meta()`.
- Preserve explicit extension-specific codes after `Type(...)` when status/category remain canonical.
- Migrate known caller(s) that currently override typed status/category to a canonical type.
- Update freeze and error-builder tests for the stricter rule.
- Document typed error metadata consistency.

Non-goals:
- Do not remove `ErrorBuilder.Type`.
- Do not prevent extension-specific error codes.
- Do not change untyped explicit `Status/Category/Code` errors.

Files:
- contract/errors.go
- contract/errors_test.go
- contract/freeze_test.go
- x/messaging/api.go
- docs/modules/contract/README.md

Tests:
- go test -timeout 20s ./contract/... ./x/messaging/...
- go build ./...
- go run ./internal/checks/module-manifests

Docs Sync:
- Update contract docs for typed metadata behavior.

Done Definition:
- Typed errors always emit canonical status/category for their type.
- Extension-specific codes remain possible without status/category drift.
- Targeted tests and build validation pass.

Outcome:
- Normalized typed `APIError` values so status and category always match `ErrorType.Meta()`.
- Preserved explicit extension-specific codes on typed errors.
- Migrated messaging deadline failures to `TypeGatewayTimeout` and `CodeGatewayTimeout`.
- Updated freeze/error-builder tests and messaging timeout assertions.
- Documented typed error metadata consistency.

Validation:
- go test -timeout 20s ./contract/... ./x/messaging/...
- go build ./...
- go run ./internal/checks/module-manifests
