# Card 0759

Milestone: v1
Recipe: specs/change-recipes/stable-root-boundary-review.yaml
Priority: P1
State: done
Primary Module: contract
Owned Files:
- specs/contract-error-codes.json
- contract/conformance_test.go
- docs/modules/contract/README.md
Depends On:
- 0758

Goal:
Make extension-owned custom error codes machine-checkable against their selected contract error type.

Scope:
- Add a lightweight JSON registry for extension-owned custom error codes used with `contract.NewErrorBuilder().Type(...).Code(...)`.
- Extend conformance tests to require registered extension-owned codes and verify their declared `contract.Type*` family.
- Inventory existing extension/reference custom codes.
- Document the registry update rule.

Non-goals:
- Do not move extension-owned codes into `contract`.
- Do not add a runtime registry.
- Do not change error response shape.

Files:
- specs/contract-error-codes.json
- contract/conformance_test.go
- docs/modules/contract/README.md

Tests:
- go test -timeout 20s ./contract/...
- go vet ./contract/...

Docs Sync:
- Document registry ownership and review expectations.

Done Definition:
- Existing extension-owned custom codes used through typed builders are registered.
- New unregistered typed custom codes fail conformance tests.
- Target checks pass.

Outcome:
- Added `specs/contract-error-codes.json` as a lightweight typed custom error code registry.
- Extended contract conformance to require same-file string consts and string literals used after typed builders to be registered.
- Registered existing typed extension/reference custom codes.
- Left dynamic variable code parameters out of the registry rule because they are not stable code symbols.
- Documented the registry update requirement.

Validation:
- go test -timeout 20s ./contract/...
- go vet ./contract/...
