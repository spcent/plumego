# Card 0842

Priority: P1
State: active
Primary Module: store
Owned Files:
- `store/file/file.go`
- `store/file/types.go`
- `store/file/image.go`
- `store/module.yaml`
- `docs/modules/store/README.md`
- `x/data/file`
- `x/fileapi`

Goal:
- Shrink stable `store/file` to tenant-agnostic file storage contracts and shared file metadata only.
- Remove provider-specific backend config and image-processing ownership from the stable store layer.

Problem:
- `store/file/types.go` still exports `StorageConfig` with local and S3-specific fields, which makes the stable store layer own backend/provider configuration rather than pure storage contracts.
- `store/file/file.go` and `store/file/image.go` export `ImageProcessor`, `NewImageProcessor`, and thumbnail-oriented behavior, which is file-processing feature logic rather than a persistence primitive.
- Stable `PutOptions` still carries thumbnail generation controls (`GenerateThumb`, `ThumbWidth`, `ThumbHeight`), which couples the storage contract to an image-processing pipeline.
- `x/data/file` local and S3 implementations depend directly on the stable image-processing helper and duplicate much of the stable file type surface, which indicates the stable boundary is still carrying implementation concerns.

Scope:
- Remove backend/provider configuration ownership from stable `store/file` and move it to `x/data/file`.
- Remove stable image-processing implementations and thumbnail-generation controls from `store/file`; move them to the owning data/file or file API layer.
- Keep stable `store/file` focused on tenant-agnostic storage interfaces, shared file metadata records, and transport-agnostic errors.
- Update `x/data/file` and `x/fileapi` to own thumbnail generation and backend-specific configuration in the same change.
- Sync store docs and manifest to the reduced `store/file` boundary.

Non-goals:
- Do not add multipart parsing or HTTP handlers to stable `store/file`.
- Do not move tenant-aware path policy or metadata querying into stable `store/file`.
- Do not preserve removed provider config or image helpers via compatibility wrappers.
- Do not redesign the tenant-aware `x/data/file` API beyond what is needed for ownership cleanup.

Files:
- `store/file/file.go`
- `store/file/types.go`
- `store/file/image.go`
- `store/module.yaml`
- `docs/modules/store/README.md`
- `x/data/file`
- `x/fileapi`

Tests:
- `go test -timeout 20s ./store/file ./x/data/file ./x/fileapi`
- `go test -race -timeout 60s ./store/file ./x/data/file ./x/fileapi`
- `go vet ./store/file ./x/data/file ./x/fileapi`

Docs Sync:
- Keep the store manifest and primer aligned on the rule that stable `store/file` owns tenant-agnostic file contracts and shared metadata only, while backend-specific config and image-processing pipelines live in `x/data/file` or `x/fileapi`.

Done Definition:
- Stable `store/file` no longer exports provider-specific backend config or image-processing implementations.
- Thumbnail-generation controls and image-processing ownership live in the owning extension layer.
- `x/data/file` and `x/fileapi` compile against the converged store/file boundary with no residual references to removed stable helpers.
- Store docs and manifest describe the same reduced `store/file` surface the code implements.

Outcome:
- Completed.
- Removed `ImageProcessor`, `NewImageProcessor`, `ImageInfo`, `StorageConfig`, and thumbnail controls from stable `store/file`.
- Moved S3 backend config and image-processing ownership into `x/data/file`, including the thumbnail helper used by local storage.
- Updated `x/data/file` and `x/fileapi` to compile against the reduced stable boundary with no residual references to the removed store/file helpers.
- Synced the store manifest and module docs to the reduced `store/file` contract boundary.
