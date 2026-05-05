# Card 0748

Milestone:
Recipe: specs/change-recipes/store-stable.yaml
Priority: P1
State: active
Primary Module: x/data/file
Owned Files:
- x/data/file/local.go
- x/data/file/local_test.go
Depends On:

Goal:
Make LocalStorage.GetURL enforce the same invalid-path contract as other local file operations.

Scope:
- Validate requested path before generating static URLs.
- Escape URL path segments without allowing traversal.
- Add regression tests for unsafe paths and normal nested paths.

Non-goals:
- Do not implement signed local URLs.
- Do not change stable store/file interfaces.

Files:
- x/data/file/local.go
- x/data/file/local_test.go

Tests:
- go test -timeout 20s ./x/data/file

Docs Sync:
- None unless local URL behavior is documented elsewhere.

Done Definition:
- Unsafe Local GetURL inputs return ErrInvalidPath.
- Safe paths produce correctly escaped static URLs.
- Targeted tests pass.

Outcome:

