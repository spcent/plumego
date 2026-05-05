# Card 0758

Milestone: v1
Recipe: specs/change-recipes/stable-root-boundary-review.yaml
Priority: P1
State: active
Primary Module: contract
Owned Files:
- contract/conformance_test.go
- docs/modules/contract/README.md
Depends On:
- 0757

Goal:
Reduce blind spots in repo-level contract conformance checks.

Scope:
- Derive conformance scan roots from the repository control plane instead of a hard-coded subset.
- Tighten `ValidateStruct` allowlisting from file-level counts to named function callsite counts.
- Add a conformance check that rejects external non-test `contract.WriteResponse` calls with non-2xx statuses.
- Document the conformance scan surfaces.

Non-goals:
- Do not ban internal contract tests from freezing non-2xx behavior.
- Do not change `WriteResponse` runtime behavior.
- Do not migrate external callers unless the new check finds violations.

Files:
- contract/conformance_test.go
- docs/modules/contract/README.md

Tests:
- go test -timeout 20s ./contract/...
- go vet ./contract/...

Docs Sync:
- Update contract conformance guard documentation.

Done Definition:
- Conformance scans cover stable, extension, reference, and tooling roots declared by repo specs.
- External `ValidateStruct` usage is tied to named functions.
- External non-2xx `WriteResponse` usage fails tests.
