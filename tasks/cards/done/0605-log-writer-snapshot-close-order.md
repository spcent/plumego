# Card 0605: Log Writer Snapshot and Close Ordering

Milestone: none
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P1
State: done
Primary Module: log
Owned Files:
- `log/glog.go`
Depends On:
- `tasks/cards/done/0603-log-file-backend-lifecycle.md`

Goal:
Make the internal text backend snapshot file writers and write to them while
holding the same mutex used by `Close`.

Problem:
The prior lifecycle pass made `Close` take `writeMu`, but `logInternal` still
captured the effective writer before taking `writeMu`. If `Close` ran after the
writer snapshot and before the write lock, logging could still attempt to write
to a stale closed file handle.

Scope:
- Move writer and backtrace decisions under `writeMu`.
- Preserve existing level filtering, header formatting, and rotation behavior.

Non-goals:
- Do not expose file logging or rotation as stable public API.
- Do not change text log formatting.
- Do not add dependencies.

Outcome:
- Moved effective-writer and backtrace snapshotting inside the `writeMu`
  critical section.
- Kept rotation accounting and fatal handling behavior unchanged.

Validation:
- `go test -race -timeout 60s ./log/...`
- `go test -timeout 20s ./log/...`
- `go vet ./log/...`

Done Definition:
- `logInternal` cannot write through a writer snapshot taken before `Close`.
- Race and normal log package tests pass.
