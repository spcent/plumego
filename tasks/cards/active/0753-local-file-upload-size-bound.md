# Card 0753

Milestone:
Recipe: specs/change-recipes/store-stable.yaml
Priority: P2
State: active
Primary Module: x/data/file
Owned Files:
- x/data/file/config.go
- x/data/file/local.go
- x/data/file/local_test.go
Depends On:

Goal:
Give LocalStorage the same predictable upload-size boundary expected from file providers.

Scope:
- Add a local upload size limit using existing package configuration style.
- Reject known oversized opts.Size before writing.
- Bound unknown-size readers during copy.
- Add regression tests for known and unknown oversized local uploads.

Non-goals:
- Do not change S3 upload buffering.
- Do not add streaming quota accounting beyond local Put.

Files:
- x/data/file/config.go
- x/data/file/local.go
- x/data/file/local_test.go

Tests:
- go test -timeout 20s ./x/data/file

Docs Sync:
- docs/modules/store/README.md if provider size defaults are documented.

Done Definition:
- Local Put cannot consume unbounded disk from a single upload.
- Oversized local uploads expose ErrInvalidSize.
- Targeted tests pass.

Outcome:

