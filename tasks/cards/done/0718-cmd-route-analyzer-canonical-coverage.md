# Card 0718

Milestone: cmd stable hardening
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P1
State: done
Primary Module: cmd/plumego
Owned Files: cmd/plumego/internal/routeanalyzer/analyzer.go, cmd/plumego/commands/routes.go, cmd/plumego/internal/routeanalyzer/analyzer_test.go
Depends On: 0717

Goal:
Make `plumego routes` accurately report canonical route registration shape.

Scope:
- Recognize `http.HandlerFunc(api.Hello)` and similar canonical handler wrappers.
- Report parse errors instead of silently hiding all failures.
- Validate or remove unsupported route filters that are not populated.
- Add route analyzer tests with canonical scaffold snippets.

Non-goals:
- Do not implement whole-program dataflow.
- Do not parse dynamic route strings.
- Do not change router behavior.

Files:
- `cmd/plumego/internal/routeanalyzer/analyzer.go`
- `cmd/plumego/commands/routes.go`
- `cmd/plumego/internal/routeanalyzer/analyzer_test.go`

Tests:
- `go test ./commands ./internal/routeanalyzer`
- `go build .`

Docs Sync:
None unless command flags change.

Done Definition:
- Canonical generated routes show meaningful handlers and locations.
- Analyzer failures are visible to callers.

Outcome:
- Made the route analyzer report Go parse errors instead of silently hiding
  files it cannot parse.
- Added handler extraction for canonical `http.HandlerFunc(api.Hello)` wrappers.
- Rejected unsupported `--group` route filtering until the analyzer can populate
  group data.
- Added route analyzer regression tests for canonical handlers and parse errors.
- Validation Run:
  - `go test ./commands ./internal/routeanalyzer`
  - `go build .`
