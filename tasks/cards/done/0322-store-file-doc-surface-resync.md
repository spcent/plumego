# Card 0322: Store File Doc Surface Resync

Priority: P1
State: done
Primary Module: store

## Goal

Bring the `store/file` documentation and testing guide back in line with the actual stable file primitive surface after tenant-aware and transport-heavy behavior moved to `x/data/file` and `x/fileapi`.

## Problem

- `store/file/README.md` correctly describes `store/file` as a pure, tenant-agnostic storage contract layer.
- `store/file/TESTING.md` is now badly out of sync with that boundary:
  - it documents non-existent files such as `utils_test.go`, `image_test.go`, `local_test.go`, `s3_signer_test.go`, `handler_test.go`, and `example_test.go`
  - it documents tenant-aware metadata, handler behavior, MinIO/PostgreSQL integration, uploader filters, `tenant_id` schema, thumbnails, and image processing under the stable root
  - it references migrations and integration surfaces that no longer live in `store/file`
- The live `store/file` directory now contains only the stable contract files plus `coverage_test.go`, so the testing guide is misleading both architecturally and mechanically.

## Scope

- Rewrite or drastically reduce `store/file/TESTING.md` so it documents only the current stable `store/file` surface and tests.
- Point tenant-aware, backend-specific, metadata-manager, image-processing, and transport-heavy testing guidance to the owning extension/module paths.
- Keep `store/file/README.md` and `store/file/TESTING.md` consistent about the stable boundary.

## Non-Goals

- Do not reintroduce handler, image, S3, or metadata-manager code into `store/file`.
- Do not migrate extension test suites in this card.
- Do not change the `store/file` API itself unless a documentation mismatch reveals a real stable-surface bug.

## Files

- `store/file/README.md`
- `store/file/TESTING.md`
- `docs/modules/store/README.md`

## Tests

- `go test -timeout 20s ./store/file`
- `go test -race -timeout 60s ./store/file`

## Docs Sync

- Keep `docs/modules/store/README.md` aligned if the stable file boundary wording becomes more explicit.

## Done Definition

- `store/file` docs describe only the current stable contract surface.
- The testing guide no longer claims tenant-aware, image, handler, S3, or database integration ownership for the stable root.
- Stable store docs and package docs point readers to the correct owning extension modules for those behaviors.

## Outcome

- Replaced the stale `store/file/TESTING.md` with a short stable-surface testing guide tied to the current package contents.
- Added an explicit testing note to `store/file/README.md` so the package docs and testing guide now describe the same stable boundary.
- Removed migration-era claims that `store/file` owns handlers, image processing, backend implementations, MinIO/Postgres integration, or metadata-manager behavior.

## Validation Run

```bash
go test -timeout 20s ./store/file
go test -race -timeout 60s ./store/file
go vet ./store/file
```
