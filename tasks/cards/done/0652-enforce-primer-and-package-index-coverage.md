# Card 0652

Priority: P0

Goal:
- Add mechanical checks for module primer coverage and package hotspot metadata coverage.

Scope:
- app-facing extension root primer coverage
- package hotspot index coverage and path validation
- failure messaging for missing or stale control-plane entries

Non-goals:
- Do not add broader semantic boundary checks yet.
- Do not require package-hotspot coverage for every package in the repository.
- Do not redesign the package-hotspot format in this card.

Files:
- `internal/checks/agent-workflow/main.go`
- `internal/checks/checkutil/checkutil.go`
- `internal/checks/checkutil/checkutil_test.go`
- `specs/package-hotspots.yaml`

Tests:
- `go test ./internal/checks/...`
- `go run ./internal/checks/agent-workflow`

Docs Sync:
- Keep package hotspot and module primer coverage rules aligned with the current control-plane definitions.

Done Definition:
- Registered app-facing extension roots fail mechanical checks when they lack module primer coverage.
- `specs/package-hotspots.yaml` entries fail mechanical checks when package paths or declared start files do not resolve.
- Control-plane coverage failures produce actionable output instead of relying on review-time discovery.

Outcome:
- `internal/checks/agent-workflow/main.go` now enforces module primer coverage for canonical app-facing extension entrypoints.
- `internal/checks/checkutil/checkutil.go` now validates `specs/package-hotspots.yaml` package existence and `start_with` path resolution.
- `internal/checks/checkutil/checkutil_test.go` now covers canonical entrypoint parsing, primer coverage failures, package-hotspot parsing, and stale package-hotspot path failures.

Validation Run:
- `go test ./internal/checks/...`
- `go run ./internal/checks/agent-workflow`
