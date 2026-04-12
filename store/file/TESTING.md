# store/file Testing Guide

This package is the stable, transport-agnostic file storage contract layer.

Current package contents:

- `file.go` — `Storage` interface
- `types.go` — shared file metadata/value types
- `errors.go` — stable file error types
- `coverage_test.go` — zero-value and error-shape coverage for the stable surface

## What belongs here

Tests under `store/file` should cover only the stable contract layer:

- error shape and `errors.Is` behavior
- zero-value safety for exported types
- stable value-type fields and tags
- small helper semantics that remain transport-agnostic

## What does not belong here

Do not add tests for behaviors owned by extension modules:

- tenant-aware storage backends, metadata persistence, and image processing
  Use `x/data/file`.
- HTTP upload/download handlers, multipart parsing, and temporary URL endpoints
  Use `x/fileapi`.
- provider-specific S3/local filesystem implementations or backend config
  Use `x/data/file`.
- database schema, migrations, or metadata-manager behavior
  Use the owning extension package.

## Current validation

Run the stable package checks with:

```bash
go test -timeout 20s ./store/file
go test -race -timeout 60s ./store/file
go vet ./store/file
```

## Expected test style

- keep tests next to the stable package
- keep assertions focused on exported contract behavior
- avoid introducing backend fixtures or external services here
- prefer short table or direct value tests over integration-style setups

## Related modules

- Stable boundary overview: `docs/modules/store/README.md`
- Tenant-aware file implementations: `x/data/file`
- File HTTP transport: `x/fileapi`
- Package overview: `store/file/README.md`
