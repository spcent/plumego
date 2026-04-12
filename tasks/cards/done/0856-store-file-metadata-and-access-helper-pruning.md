# Card 0856

Priority: P1
State: done
Primary Module: store
Owned Files:
- `store/file`
- `store/module.yaml`
- `docs/modules/store/README.md`
- `x/data/file`
- `x/fileapi`

Goal:
- Narrow stable `store/file` to minimal storage contracts and transport-agnostic file metadata.
- Move uploader, image, path/id helper, and signed-URL ownership to the extension packages that actually own those concerns.

Problem:
- Stable `store/file` still carries `GetURL(...)`, `MetadataManager`, `GenerateID`, `IsPathSafe`, `MimeToExt`, and `ExtToMime`.
- Stable file models still include uploader and image-pipeline fields such as `UploadedBy`, `Width`, `Height`, and `ThumbnailPath`.
- These surfaces are now consumed by `x/data/file` and `x/fileapi`, which means stable `store/file` still owns provider, image, and API convenience that does not belong in a minimal persistence primitive layer.

Scope:
- Remove signed-URL, metadata-manager, uploader, image, and file-api helper ownership from stable `store/file`.
- Keep stable `store/file` limited to minimal file storage contracts and pure storage metadata.
- Move the richer models and helper functions into `x/data/file` and update `x/fileapi` in the same change.
- Sync store and extension docs/manifests to the reduced stable boundary.

Non-goals:
- Do not redesign local or S3 backend behavior.
- Do not redesign `x/fileapi` HTTP handlers beyond what the ownership move requires.
- Do not introduce compatibility aliases in stable `store/file`.

Files:
- `store/file`
- `store/module.yaml`
- `docs/modules/store/README.md`
- `x/data/file`
- `x/fileapi`

Tests:
- `go test -timeout 20s ./store/file ./x/data/file ./x/fileapi`
- `go test -race -timeout 60s ./store/file ./x/data/file ./x/fileapi`
- `go vet ./store/file ./x/data/file ./x/fileapi`

Docs Sync:
- Keep store docs aligned on the rule that stable `store/file` owns only transport-agnostic storage primitives.
- Keep `x/data/file` and `x/fileapi` docs aligned on the rule that signed URLs, uploader semantics, and image pipeline metadata are extension-owned.

Done Definition:
- Stable `store/file` no longer exports uploader/image-specific metadata fields or file-api/provider convenience helpers.
- `x/data/file` and `x/fileapi` own the richer file models and helper behavior they actually use.
- Store docs/manifests describe the same reduced stable surface implemented in code.

Outcome:
- Completed.
- Reduced stable `store/file` to the minimal storage contract, shared base file types, and file operation errors.
- Removed stable ownership of signed URLs, metadata-manager interfaces, uploader/image fields, and path/id/MIME helper functions; those concerns now live with `x/data/file` and its tenant-aware richer models.
- Updated store docs so the stable layer no longer advertises provider, signed-URL, metadata-manager, or image-pipeline responsibilities.
