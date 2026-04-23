# Card 2128: Data File Image Thumbnail Capability
Milestone: none
Recipe: specs/change-recipes/fix-bug.yaml
Priority: medium
State: active
Primary Module: x/data/file
Owned Files:
- x/data/file/image.go
- x/data/file/image_test.go
- x/data/file/local.go
- x/data/file/local_test.go
- docs/modules/x-data/README.md
Depends On: none

Goal:
Separate image MIME recognition from thumbnail-processing capability in
`x/data/file`. The module maps WebP MIME and extension values, and `IsImage`
currently returns true for `image/webp`, but thumbnail generation relies on the
standard library image decoders and encoders, which do not handle WebP.

Scope:
- Preserve WebP MIME and extension mapping for metadata and file validation.
- Add an explicit thumbnail-support check for the formats actually handled by
  the stdlib pipeline.
- Make `LocalStorage.Put` skip thumbnail generation for recognized but
  unsupported image formats rather than attempting a decode that must fail.
- Add tests for WebP metadata mapping and unsupported thumbnail behavior.

Non-goals:
- Do not add a WebP codec dependency.
- Do not change the main module dependency policy.
- Do not rewrite the storage API or thumbnail naming scheme.

Tests:
- go test -race -timeout 60s ./x/data/file/...
- go test -timeout 20s ./x/data/file/...
- go vet ./x/data/file/...

Docs Sync:
Update `docs/modules/x-data/README.md` if it documents thumbnail-supported image
formats or supported MIME behavior.

Done Definition:
- WebP remains a recognized MIME/extension mapping.
- Thumbnail generation is limited to formats the stdlib processor can decode and
  encode.
- Unsupported image formats do not fail an otherwise valid upload solely because
  thumbnail generation was requested.
- The x/data/file validation commands pass.

Outcome:
Added an explicit `SupportsThumbnail` check to the image processor and limited
local thumbnail generation to stdlib-supported formats. WebP remains a
recognized image MIME/extension mapping, but local storage now skips thumbnail
generation for WebP instead of attempting the unsupported decode path. Added
tests for supported JPEG thumbnail generation and unsupported WebP thumbnail
skipping.

Validation:
- `go test -race -timeout 60s ./x/data/file/...`
- `go test -timeout 20s ./x/data/file/...`
- `go vet ./x/data/file/...`
