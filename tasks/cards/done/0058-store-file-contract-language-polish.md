# Card 0058

Milestone:
Recipe: specs/change-recipes/refine-docs.yaml
Priority: P2
State: done
Primary Module: store
Owned Files:
- store/file/types.go
- store/file/file.go
- store/file/coverage_test.go
Depends On:
- 0057-store-idempotency-record-ownership-docs

Goal:
Keep `store/file` wording and tests strictly stable-layer scoped and provider-neutral.

Scope:
- Clarify negative list limits and invalid-size ownership without adding helpers.
- Reword shared type comments to avoid implying stable tenant behavior.
- Remove tenant-specific fixture language from stable package tests.

Non-goals:
- Do not change the `Storage` interface.
- Do not add path validation, provider config, signed URL, metadata-manager, or image-processing behavior.
- Do not import extension packages.

Files:
- store/file/types.go
- store/file/file.go
- store/file/coverage_test.go

Tests:
- go test -timeout 20s ./store/file
- go test -race -timeout 60s ./store/file
- go vet ./store/file

Docs Sync:
- Not required; package comments only.

Done Definition:
- File comments remain provider- and tenant-neutral.
- Stable tests no longer use tenant-specific sample paths.
- File targeted tests and vet pass.

Outcome:
- Clarified `Storage.List` negative-limit behavior as `ErrInvalidSize`.
- Reworded shared file type comments to stay stable-layer scoped and provider-neutral.
- Replaced tenant-specific test fixture path with a neutral file path.

Validation:
- go test -timeout 20s ./store/file
- go test -race -timeout 60s ./store/file
- go vet ./store/file
