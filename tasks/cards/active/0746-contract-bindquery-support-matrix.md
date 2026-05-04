# Card 0746

Milestone: v1
Recipe: specs/change-recipes/stable-root-boundary-review.yaml
Priority: P1
State: active
Primary Module: contract
Owned Files:
- contract/context_bind.go
- contract/context_extended_test.go
- docs/modules/contract/README.md
Depends On:
- 0745

Goal:
Freeze the `BindQuery` support matrix and cover remaining query-binding edge cases that affect stable usability.

Scope:
- Add tests for repeated scalar selection, explicit empty values, embedded struct behavior, unsupported maps, and time/TextUnmarshaler behavior.
- Document the supported and unsupported field shapes.
- Keep query binding reflection scope narrow.

Non-goals:
- Do not add map binding.
- Do not add nested object binding.
- Do not create a new validation framework.

Files:
- contract/context_bind.go
- contract/context_extended_test.go
- docs/modules/contract/README.md

Tests:
- go test -timeout 20s ./contract/...
- go vet ./contract/...

Docs Sync:
- Update contract docs with a concrete `BindQuery` matrix.

Done Definition:
- Supported and unsupported query-binding field shapes are covered by tests.
- Docs match actual behavior.
- Contract tests and vet pass.

Outcome:

