# Card 0821

Priority: P1
State: active
Primary Module: x/fileapi
Owned Files:
- `docs/modules/x-fileapi/README.md`
- `docs/modules/store/README.md`
- `store/file/README.md`
- `x/data/file/module.yaml`
- `x/fileapi/module.yaml`
Depends On:

Goal:
- Keep the `x/fileapi` / `x/data/file` / `store/file` boundary docs aligned on transport, implementation, and stable-contract ownership.

Scope:
- Align the file transport primer, stable store primer, stable store/file README, and the two manifests on exact responsibility boundaries.
- Make the transport-vs-storage-vs-stable-contract split explicit and grep-friendly for future file-related work.
- Keep the card documentation-only unless a manifest wording gap requires a tiny example or clarification in place.

Non-goals:
- Do not change file upload or storage runtime behavior in this card.
- Do not add new file backends or HTTP handlers.
- Do not collapse the three layers into one “file module” narrative.

Files:
- `docs/modules/x-fileapi/README.md`
- `docs/modules/store/README.md`
- `store/file/README.md`
- `x/data/file/module.yaml`
- `x/fileapi/module.yaml`

Tests:
- `go test -timeout 20s ./store/file ./x/data/file ./x/fileapi`
- `go test -race -timeout 60s ./store/file ./x/data/file ./x/fileapi`
- `go vet ./store/file ./x/data/file ./x/fileapi`

Docs Sync:
- Keep file boundary guidance aligned across the transport primer, stable store primer, stable store/file README, and the `x/data/file` / `x/fileapi` manifests.

Done Definition:
- The file capability control plane is described consistently across stable and extension docs.
- No doc implies tenant-aware storage belongs in stable `store/file` or that HTTP handlers belong in `x/data/file`.
- Targeted file-layer validation stays green.

Outcome:
- Aligned `docs/modules/x-fileapi/README.md`, `docs/modules/store/README.md`, `store/file/README.md`, `x/data/file/module.yaml`, and `x/fileapi/module.yaml` on one boundary model: `x/fileapi` owns HTTP transport, `x/data/file` owns tenant-aware storage and metadata implementations, and `store/file` owns the stable transport-agnostic contracts and helpers.
- Fixed the stale `store/file/file.go` package comment so it no longer implies concrete local/S3 backends live in the stable layer.
- Validation:
  - `go test -timeout 20s ./store/file ./x/data/file ./x/fileapi`
  - `go test -race -timeout 60s ./store/file ./x/data/file ./x/fileapi`
  - `go vet ./store/file ./x/data/file ./x/fileapi`
