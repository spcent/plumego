# Card 0755

Milestone: v1
Recipe: specs/change-recipes/stable-root-boundary-review.yaml
Priority: P2
State: active
Primary Module: contract
Owned Files:
- contract/conformance_test.go
- docs/modules/contract/README.md
Depends On:
- 0754

Goal:
Control the remaining external production use of `contract.ValidateStruct` before stable.

Scope:
- Add a conformance allowlist for external non-test `contract.ValidateStruct` usage.
- Preserve existing known callers in `x/messaging`, `x/ops`, and `reference/workerfleet` as compatibility-approved users.
- Document that new production usage requires an explicit contract review or module-local validation rationale.

Non-goals:
- Do not migrate all existing callers in this card.
- Do not remove or deprecate `ValidateStruct`.
- Do not add a full validation framework.

Files:
- contract/conformance_test.go
- docs/modules/contract/README.md

Tests:
- go test -timeout 20s ./contract/...
- go vet ./contract/...

Docs Sync:
- Update validation stable-surface governance.

Done Definition:
- Existing external usage is inventory-backed.
- New external production usage fails the conformance check unless added deliberately.
- Targeted checks pass.
