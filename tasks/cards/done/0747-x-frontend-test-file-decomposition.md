# Card 0747: x/frontend Test File Decomposition

Milestone: none
Recipe: specs/change-recipes/module-cleanup.yaml
Priority: P2
State: done
Primary Module: x/frontend
Owned Files:
- `x/frontend/frontend_test.go`
- `x/frontend/mount_test.go`
- `x/frontend/security_test.go`
- `x/frontend/compression_test.go`
- `x/frontend/response_test.go`
Depends On: 0746

Goal:
Split the oversized `frontend_test.go` into behavior-focused test files without
changing production behavior.

Scope:
- Move mount/registration tests into `mount_test.go`.
- Move path/symlink/security tests into `security_test.go`.
- Move compression/encoding tests into `compression_test.go`.
- Move response/cache/error/header tests into `response_test.go`.
- Keep shared test helpers in `frontend_test.go`.

Non-goals:
- Do not change production code.
- Do not add new coverage except where needed to keep tests compiling.
- Do not rename public symbols.

Files:
- `x/frontend/frontend_test.go`
- `x/frontend/mount_test.go`
- `x/frontend/security_test.go`
- `x/frontend/compression_test.go`
- `x/frontend/response_test.go`

Tests:
- `go test -race -timeout 60s ./x/frontend/...`
- `go test -timeout 20s ./x/frontend/...`
- `go vet ./x/frontend/...`

Docs Sync:
No docs change required.

Done Definition:
- Test files are organized by behavior area.
- No test behavior is lost.
- The listed validation commands pass.

Outcome:
- Split the former oversized `frontend_test.go` into shared helpers plus
  behavior-focused `mount_test.go`, `security_test.go`, `compression_test.go`,
  and `response_test.go`.
- Kept production code unchanged and moved existing tests/types mechanically by
  behavior area.
- Validation passed:
  - `go test -race -timeout 60s ./x/frontend/...`
  - `go test -timeout 20s ./x/frontend/...`
  - `go vet ./x/frontend/...`
