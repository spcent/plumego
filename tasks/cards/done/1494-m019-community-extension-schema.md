# Card 1590

Milestone: M-019
Recipe: specs/change-recipes/update-docs.yaml
Priority: P3
State: done
Primary Module: specs
Owned Files:
- `specs/community-extension.schema.yaml`
- `internal/checks/community-extension/main.go`

Goal:
- Define the machine-readable contract that community-authored plumego
  extensions must satisfy, and add a check tool that validates any package's
  community-extension.yaml against the schema.

Scope:
- Create specs/community-extension.schema.yaml defining required fields:
  - name: string (package short name)
  - module_path: string (full Go module path)
  - status: enum [experimental, beta, ga]
  - handler_shape: const "func(http.ResponseWriter, *http.Request)"
  - test_commands: []string (at least one entry required)
  - forbidden_imports: []string (must include all stable root package paths)
  - no_init_side_effects: bool (must be true)
  - no_globals: bool (must be true)
  - owner: string (team or contact)
- Create internal/checks/community-extension/main.go:
  - Reads community-extension.yaml from a given directory.
  - Validates against specs/community-extension.schema.yaml.
  - Exits 1 with descriptive error if any required field is missing or invalid.
  - Exits 0 if schema is satisfied.
- Write internal/checks/community-extension/main_test.go covering:
  - Valid schema passes.
  - Missing name field exits 1.
  - Invalid status value exits 1.
  - Missing test_commands exits 1.
  - no_init_side_effects = false exits 1.

Non-goals:
- Do not build a hosted registry in this card (that is M-019 scope but a later card).
- Do not enforce forbidden_imports via static analysis here (that is card 1591).
- Do not add community-extension.yaml to any existing x/* package yet.

Files:
- `specs/community-extension.schema.yaml`
- `internal/checks/community-extension/main.go`
- `internal/checks/community-extension/main_test.go`

Tests:
- `go test -timeout 20s ./internal/checks/community-extension/...`
- `go vet ./internal/checks/community-extension/...`
- `go run ./internal/checks/module-manifests`

Docs Sync:
- none at this card; authoring guide written in card 1592.

Done Definition:
- specs/community-extension.schema.yaml exists with all required fields defined.
- go run ./internal/checks/community-extension validates a valid schema exits 0.
- All five check test cases pass.
- `go run ./internal/checks/module-manifests` exits 0.

Outcome:
- Added `specs/community-extension.schema.yaml` with required fields, status
  enum, handler-shape constant, required true booleans, list minimums, and
  stable-root forbidden import requirements.
- Added `internal/checks/community-extension`, a stdlib-only check tool that
  validates a target directory's `community-extension.yaml` against the schema.
- Added tests for valid manifests, missing `name`, invalid `status`, missing
  `test_commands`, and `no_init_side_effects: false`.
- Validated with `go test -timeout 20s ./internal/checks/community-extension/...`,
  `go vet ./internal/checks/community-extension/...`,
  `go run ./internal/checks/community-extension <valid-temp-dir>`,
  `go run ./internal/checks/module-manifests`, touched-file `gofmt -l`, and
  `git diff --check`.
