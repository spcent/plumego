# Card 0828

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P2
State: done
Primary Module: store
Owned Files:
- store/file/file.go
- store/file/types.go
- docs/modules/store/README.md
Depends On: 0725

Goal:
Make the stable file-store contract explicit so users do not mistake `store/file` for a bundled backend implementation.

Scope:
- Clarify package docs that `store/file` is a contract and shared-type layer.
- Point users to extension implementations without importing or depending on `x/*`.
- Clarify caller-owned metadata semantics and implementation responsibilities.
- Sync the store module README.

Non-goals:
- Do not add local, S3, or tenant-aware file backends to stable store.
- Do not add signed URL, metadata-manager, uploader, or image-processing policy.
- Do not change exported file interfaces or structs.

Files:
- `store/file/file.go`
- `store/file/types.go`
- `docs/modules/store/README.md`

Tests:
- `go test -timeout 20s ./store/...`
- `go vet ./store/...`
- `go run ./internal/checks/dependency-rules`

Docs Sync:
- This is a docs and comments sync card.

Done Definition:
- Stable file docs clearly describe contract-only availability.
- Store docs point implementation work to owning extensions.
- Store tests, vet, and dependency boundary checks pass.

Outcome:
- Clarified `store/file` package docs as contract-only with no bundled backend.
- Documented implementation responsibilities for caller contexts and defensive metadata copying.
- Synced the store module README to point backend implementation work to owning extensions.
- Validation run: `go test -timeout 20s ./store/...`; `go vet ./store/...`; `go run ./internal/checks/dependency-rules`.
