# Card 0734

Milestone: cmd stable hardening
Recipe: specs/change-recipes/refactor-small.yaml
Priority: P1
State: done
Primary Module: cmd/plumego/commands
Owned Files: cmd/plumego/internal/buildtarget/buildtarget.go, cmd/plumego/internal/buildtarget/buildtarget_test.go, cmd/plumego/commands/build.go, cmd/plumego/internal/devserver/builder.go, cmd/plumego/commands/cli_e2e_test.go
Depends On: 0733

Goal:
Use one reliable build target detector for `plumego build` and `plumego dev`.

Scope:
- Add a small internal helper that parses Go files or package metadata to detect root and `cmd/app` main packages.
- Replace duplicate string-based detection in build command and dev builder.
- Add tests for comments/strings that mention `package main` but are not main packages.

Non-goals:
- Do not change the default target priority.
- Do not invoke `go list` for this helper.
- Do not alter custom build command behavior.

Files:
- `cmd/plumego/internal/buildtarget/buildtarget.go`
- `cmd/plumego/internal/buildtarget/buildtarget_test.go`
- `cmd/plumego/commands/build.go`
- `cmd/plumego/internal/devserver/builder.go`
- `cmd/plumego/commands/cli_e2e_test.go`

Tests:
- `go test ./internal/buildtarget ./commands ./internal/devserver`
- `go build .`

Docs Sync:
- None.

Done Definition:
- Build and dev use the same helper.
- Comment/string false positives are covered.
- Existing canonical `cmd/app` behavior remains covered.

Outcome:
- Added `internal/buildtarget` with parser-based main package detection.
- Replaced duplicate build target detection in `build` and dev builder.
- Added tests covering root priority, canonical `cmd/app`, and comment/string false positives.

Validation:
- `go test ./internal/buildtarget ./commands ./internal/devserver`
- `go build .`
