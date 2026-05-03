# Card 0719

Milestone: cmd stable hardening
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P1
State: active
Primary Module: cmd/plumego
Owned Files: cmd/plumego/internal/watcher/watcher.go, cmd/plumego/internal/watcher/watcher_test.go, cmd/plumego/internal/checker/checker_test.go, cmd/plumego/internal/configmgr/configmgr_test.go
Depends On: 0718

Goal:
Harden low-coverage utility packages that directly affect stable CLI behavior.

Scope:
- Add focused tests for watcher include/exclude/debounce behavior.
- Add checker tests for canonical project layout.
- Add config manager tests for env parsing and redaction.
- Fix any small defects exposed by those tests.

Non-goals:
- Do not replace the watcher with fsnotify in this card.
- Do not broaden CLI command behavior.
- Do not add new dependencies.

Files:
- `cmd/plumego/internal/watcher/watcher.go`
- `cmd/plumego/internal/watcher/watcher_test.go`
- `cmd/plumego/internal/checker/checker_test.go`
- `cmd/plumego/internal/configmgr/configmgr_test.go`

Tests:
- `go test ./internal/watcher ./internal/checker ./internal/configmgr`
- `go test ./...`

Docs Sync:
None.

Done Definition:
- Previously 0%-coverage packages have focused regression coverage.
- Any discovered utility defects are fixed.

Outcome:

