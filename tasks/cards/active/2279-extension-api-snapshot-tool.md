# Card 2279

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P1
State: active
Primary Module: internal/checks
Owned Files:
- internal/checks/extension-api-snapshot/main.go
- internal/checks/extension-api-snapshot/README.md
- specs/checks.yaml
- docs/EXTENSION_STABILITY_POLICY.md
Depends On: 2278

Goal:
Add a local exported API snapshot tool that can generate and compare package-level public API surfaces for extension beta evidence.

Scope:
- Implement a standard-library-only check under `internal/checks/extension-api-snapshot`.
- Support generating a deterministic snapshot for one module or package subtree such as `./x/rest/...`.
- Support comparing two snapshot files and reporting added, removed, or changed exported symbols.
- Document how the tool fits into the two-minor-release beta evidence workflow.

Non-goals:
- Do not fetch remote tags or call GitHub.
- Do not decide promotion status automatically.
- Do not include unstable implementation details or unexported symbols in snapshots.

Files:
- `internal/checks/extension-api-snapshot/main.go`
- `internal/checks/extension-api-snapshot/README.md`
- `specs/checks.yaml`
- `docs/EXTENSION_STABILITY_POLICY.md`

Tests:
- `go test ./internal/checks/...`
- `go run ./internal/checks/extension-api-snapshot -module ./x/rest/... -out /tmp/plumego-x-rest-api.snapshot`
- `go run ./internal/checks/extension-api-snapshot -compare /tmp/plumego-x-rest-api.snapshot /tmp/plumego-x-rest-api.snapshot`

Docs Sync:
- Required because the beta policy gains a concrete evidence tool.

Done Definition:
- A maintainer can generate a deterministic exported API snapshot locally.
- Comparing identical snapshots exits successfully.
- Removed or changed exported symbols are reported in a human-readable way.

Outcome:
