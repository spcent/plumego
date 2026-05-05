# Card 0768

Milestone: v1
Recipe: specs/change-recipes/stable-root-boundary-review.yaml
Priority: P1
State: active
Primary Module: contract
Owned Files:
- contract/conformance_test.go
- specs/contract-error-codes.json
- docs/modules/contract/README.md
Depends On:
- 0767

Goal:
Strengthen extension-owned error code conformance beyond same-file string literals.

Scope:
- Resolve package-level string constants used as typed custom codes, not only constants in the same file.
- Require dynamic typed custom code arguments to be explicitly allowlisted by stable callsite or helper.
- Update the registry or allowlist for current accepted dynamic sources.

Non-goals:
- Do not create a runtime error code registry.
- Do not move extension-owned code constants into `contract`.
- Do not validate business semantics beyond selected `contract.Type*` family.

Files:
- contract/conformance_test.go
- specs/contract-error-codes.json
- docs/modules/contract/README.md

Tests:
- go test -timeout 20s ./contract/...
- go vet ./contract/...

Docs Sync:
- Document package-level constants and dynamic-code allowlist expectations.

Done Definition:
- Package-level typed custom code constants are registry-checked.
- Dynamic typed custom code sources are no longer silently skipped.
- Target checks pass.

Outcome:

