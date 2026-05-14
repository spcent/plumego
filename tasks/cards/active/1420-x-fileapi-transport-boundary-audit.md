# Card 1420

Milestone: v1-package-cleanup
Recipe: specs/change-recipes/stable-root-boundary-review.yaml
Priority: P2
State: active
Primary Module: x/fileapi
Owned Files:
- x/fileapi/handler.go
- x/fileapi/context.go
- x/fileapi/handler_test.go
- docs/modules/x-fileapi/README.md
Depends On:
- 1419

Goal:
- Audit `x/fileapi` as an app-facing transport over `x/data/file` and stable `store/file` contracts.

Scope:
- Confirm handler responsibilities stay transport-only.
- Confirm file persistence behavior remains in `x/data/file` or stable `store/file` contracts.
- Add focused tests or documentation for any unclear ownership found during audit.

Non-goals:
- Do not change upload/download public APIs.
- Do not move storage implementations into `x/fileapi`.
- Do not promote `x/fileapi` maturity.

Files:
- x/fileapi/handler.go
- x/fileapi/context.go
- x/fileapi/handler_test.go
- docs/modules/x-fileapi/README.md

Tests:
- go test -timeout 20s ./x/fileapi
- go vet ./x/fileapi
- go run ./internal/checks/dependency-rules

Docs Sync:
- Update `docs/modules/x-fileapi/README.md` if boundary wording is clarified.

Done Definition:
- `x/fileapi` ownership is clearly limited to HTTP file transport.
- Any storage-boundary ambiguity is documented or covered by tests.
- File API tests and dependency checks pass.

Outcome:

