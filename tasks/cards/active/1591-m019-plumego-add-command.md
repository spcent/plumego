# Card 1591

Milestone: M-019
Recipe: specs/change-recipes/add-package.yaml
Priority: P3
State: active
Primary Module: cmd/plumego
Owned Files:
- `cmd/plumego/commands/add.go`
- `cmd/plumego/commands/add_test.go`

Goal:
- Implement `plumego add <module-path>` that validates a community extension's
  community-extension.yaml against the schema, checks for forbidden imports via
  go vet, then adds the module to the project's go.mod.

Scope:
- Create cmd/plumego/commands/add.go defining AddCmd:
  - Usage: `plumego add <module-path> [--version latest]`
  - Step 1: run `go list -json <module-path>` to confirm the module exists.
  - Step 2: download community-extension.yaml from the module root via go module
    proxy and validate against specs/community-extension.schema.yaml.
  - Step 3: run internal/checks/community-extension against the downloaded
    package to verify forbidden_imports and no_init_side_effects.
  - Step 4: if all checks pass, run `go get <module-path>@<version>`.
  - Step 5: print a compliance report: fields validated, checks passed, module
    added to go.mod.
  - On any failure: exit non-zero with descriptive message; do not modify go.mod.
- Write cmd/plumego/commands/add_test.go covering:
  - Valid extension module: schema passes, go get executes, exits 0.
  - Missing community-extension.yaml: exits 1 with "schema not found" message.
  - Invalid schema (missing name): exits 1 before go get.
  - Forbidden import detected: exits 1 before go get.
  - Network failure on go list: exits 1 with retry suggestion.

Non-goals:
- Do not auto-wire routes or register globals after adding.
- Do not implement a hosted registry; use module proxy for discovery.
- Do not add --force flag to skip schema validation.

Files:
- `cmd/plumego/commands/add.go`
- `cmd/plumego/commands/add_test.go`

Tests:
- `go test -race -timeout 60s ./cmd/plumego/...`
- `go vet ./cmd/plumego/...`

Docs Sync:
- Update cmd/plumego/README.md to document the add command and compliance report format.

Done Definition:
- `plumego add` exits non-zero and does not modify go.mod on schema failure.
- `plumego add` runs go get only after all checks pass.
- All five add_test.go cases pass.
- cmd/plumego/README.md documents the add command.

Outcome:
-
