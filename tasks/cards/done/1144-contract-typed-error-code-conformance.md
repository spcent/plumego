# Card 1144

Milestone: v1
Recipe: specs/change-recipes/stable-root-boundary-review.yaml
Priority: P1
State: done
Primary Module: contract
Owned Files:
- contract/conformance_test.go
- docs/modules/contract/README.md
Depends On:
- 0752

Goal:
Prevent typed error builders from drifting when callers combine a typed error with a mismatched contract-owned error code.

Scope:
- Add a repo-level conformance check for external `contract.NewErrorBuilder()` chains.
- Reject builder chains that call `.Type(contract.Type*)` and then override `.Code(contract.Code*)` with a code that is not the canonical code for that type.
- Fix any existing mismatches found by the new check.
- Document the rule for contract-owned codes while preserving extension-owned custom code space.

Non-goals:
- Do not ban extension-local custom error code constants.
- Do not add a new runtime error builder API.
- Do not change the stable error envelope shape.

Files:
- contract/conformance_test.go
- docs/modules/contract/README.md
- any external caller required to satisfy the new conformance rule

Tests:
- go test -timeout 20s ./contract/...
- go vet ./contract/...
- go test -timeout 20s ./x/messaging/...

Docs Sync:
- Document the typed error code conformance rule in the contract module README.

Done Definition:
- External contract-owned code overrides are checked automatically.
- Existing violations are fixed or explicitly documented as allowed custom extension codes.
- Targeted tests pass.

Outcome:
- Added an external-code conformance scan for `contract.NewErrorBuilder()` chains that combine `contract.Type*` and `contract.Code*`.
- Allowed known type-compatible contract validation code refinements while rejecting cross-family contract-owned code overrides.
- Fixed service-unavailable cases in `x/messaging` and `x/devtools/pubsubdebug` to use `TypeUnavailable`.
- Synced affected tests and contract README guidance.

Validation:
- go test -timeout 20s ./contract/...
- go vet ./contract/...
- go test -timeout 20s ./x/messaging/...
- go test -timeout 20s ./x/devtools/...
