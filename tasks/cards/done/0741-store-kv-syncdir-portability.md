# Card 0741

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P2
State: done
Primary Module: store
Owned Files:
- store/kv/kv.go
- store/kv/kv_test.go

Goal:
Make KV parent-directory fsync tolerate platforms that report unsupported directory fsync as `EINVAL`.

Scope:
- Treat `syscall.EINVAL` from directory sync as unsupported and non-fatal.
- Keep other sync errors fatal.
- Add focused unit coverage through an internal error classifier.

Non-goals:
- Do not change KV persistence format.
- Do not add platform-specific build tags.
- Do not add WAL or snapshot behavior.

Files:
- store/kv/kv.go
- store/kv/kv_test.go

Tests:
- go test -timeout 20s ./store/kv
- go vet ./store/kv
- go run ./internal/checks/dependency-rules

Docs Sync:
- Not required unless behavior wording changes.

Done Definition:
- Unsupported directory fsync via `os.ErrInvalid` or `syscall.EINVAL` is ignored.
- Other errors remain fatal.
- Targeted tests, vet, and dependency checks pass.

Outcome:
- Added an internal unsupported directory-sync classifier.
- Directory sync now ignores `os.ErrInvalid` and errors wrapping `syscall.EINVAL`.
- Added tests proving `EINVAL` is non-fatal while other sync errors remain fatal.

Validation:
- `gofmt -w store/kv/kv.go store/kv/kv_test.go`
- `go test -timeout 20s ./store/kv`
- `go vet ./store/kv`
- `go run ./internal/checks/dependency-rules`
