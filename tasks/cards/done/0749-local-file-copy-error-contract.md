# Card 0749

Milestone:
Recipe: specs/change-recipes/store-stable.yaml
Priority: P2
State: done
Primary Module: x/data/file
Owned Files:
- x/data/file/local.go
- x/data/file/local_test.go
Depends On:

Goal:
Make LocalStorage.Copy errors match the store/file error contract.

Scope:
- Wrap invalid paths, missing source, create, and copy failures with file operation context.
- Map missing source files to store/file.ErrNotFound.
- Add regression coverage for missing source and invalid paths.

Non-goals:
- Do not change overwrite semantics.
- Do not add metadata-copy behavior.

Files:
- x/data/file/local.go
- x/data/file/local_test.go

Tests:
- go test -timeout 20s ./x/data/file

Docs Sync:
- None.

Done Definition:
- Callers can use errors.Is with ErrNotFound and ErrInvalidPath for Copy.
- Copy errors include operation/path context.
- Targeted tests pass.

Outcome:
LocalStorage.Copy now wraps invalid paths, missing sources, create failures, and copy failures in *storefile.Error. Missing sources map to ErrNotFound. Added regression coverage for missing source and invalid source/destination paths.

Validation:
- go test -timeout 20s ./x/data/file
