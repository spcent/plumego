# Card 1376

Milestone: v1
Recipe: specs/change-recipes/stable-root-boundary-review.yaml
Priority: P1
State: active
Primary Module: contract
Owned Files:
- contract/errors.go
- contract/errors_test.go
- docs/modules/contract/README.md
Depends On:
- 0759

Goal:
Clarify and harden `APIError.Details` cloning for stable error payloads.

Scope:
- Deep-clone JSON-like detail values stored through `Detail`, `Details`, `Build`, and `WriteError` normalization.
- Preserve unsupported value passthrough behavior for existing callers.
- Add tests for nested map/slice mutation after build/write.
- Document the stable clone semantics.

Non-goals:
- Do not reject non-JSON detail values in this card.
- Do not change the error envelope fields.
- Do not add non-stdlib dependencies.

Files:
- contract/errors.go
- contract/errors_test.go
- docs/modules/contract/README.md

Tests:
- go test -timeout 20s ./contract/...
- go vet ./contract/...

Docs Sync:
- Update details clone semantics.

Done Definition:
- Nested JSON-like details are not mutated by caller changes after build/write.
- Unsupported values retain compatibility passthrough behavior.
- Target checks pass.
