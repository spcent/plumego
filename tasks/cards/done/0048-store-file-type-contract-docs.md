# Card 0048

Milestone:
Recipe: specs/change-recipes/refine-api.yaml
Priority: P2
State: done
Primary Module: store
Owned Files:
- store/file/types.go
- store/file/coverage_test.go
Depends On:
- 0047-store-idempotency-contract-docs

Goal:
Clarify `store/file` shared type semantics without moving provider policy into the stable layer.

Scope:
- Tighten comments for metadata aliasing, unknown upload size, and query zero values.
- Keep tenant and provider-specific behavior explicitly outside stable `store/file`.
- Add focused coverage for shared type field semantics.

Non-goals:
- Do not remove or reshape public file types.
- Do not add file provider configuration, signed URL, metadata-manager, thumbnail, or image-processing behavior.
- Do not import extension packages.

Files:
- store/file/types.go
- store/file/coverage_test.go

Tests:
- go test -timeout 20s ./store/file
- go test -race -timeout 60s ./store/file
- go vet ./store/file

Docs Sync:
- Not required; package comments only.

Done Definition:
- Shared file type comments describe stable-layer ownership and zero-value behavior.
- Tests cover unknown size and query field representation.
- File package remains provider- and tenant-agnostic.

Outcome:
- Clarified metadata ownership, unknown upload size, and query zero-value semantics.
- Added tests for unknown-size `PutOptions` and populated `Query` fields.

Validation:
- go test -timeout 20s ./store/file
- go test -race -timeout 60s ./store/file
- go vet ./store/file
