# Card 1502

Milestone: M-014
Recipe: specs/change-recipes/new-extension-module.yaml
Priority: P1
State: active
Primary Module: x/openapi
Owned Files:
- `x/openapi/`
- `docs/modules/x-openapi/README.md`
- `cmd/plumego/internal/scaffold/`

Goal:
- Complete the x/openapi CLI integration so that `plumego openapi generate`
  reads route metadata from a running or introspectable app and writes an
  OpenAPI 3.1 document to a file or stdout.
- Expand the x/openapi module primer to document the complete API surface.

Problem:
x/openapi/module.yaml notes "Serialization and CLI wiring are handled by later
OpenAPI milestone cards." The document model, RouteInfo conversion, and operation
hint merging are substantially implemented, but:
1. No `plumego openapi generate` subcommand exists in `cmd/plumego/`.
2. No serialization (JSON/YAML) path is wired.
3. The module primer is 7 lines.

Scope:
- Add a `generate` subcommand to `cmd/plumego/` that:
  - accepts `--output` (file path, defaults to stdout) and `--format` (json|yaml, default json)
  - calls x/openapi document builder with route metadata
  - writes the resulting OpenAPI 3.1 document
- Add JSON serialization in x/openapi (stdlib `encoding/json` only — no external YAML library
  in the main module; YAML output uses `encoding/json` + a thin converter, or is deferred).
- Expand `docs/modules/x-openapi/README.md` to cover: document model, hint merging,
  RouteInfo integration, CLI usage, and boundary rules.
- Add or update x/openapi tests for the serialization path.

Non-goals:
- Do not add an external OpenAPI validation library dependency.
- Do not add reflection over handler functions — generation is driven by RouteInfo only.
- Do not move generation logic into router, core, or contract.
- Do not support OpenAPI 2.0.
- YAML output may be deferred to a follow-up card if it requires a third-party dependency.

Files:
- `x/openapi/` (serialization and builder changes)
- `cmd/plumego/internal/scaffold/` or `cmd/plumego/` (new subcommand)
- `docs/modules/x-openapi/README.md`
- `x/openapi/module.yaml` (update test commands if changed)

Tests:
- `go test -race -timeout 60s ./x/openapi/...`
- `go test -timeout 20s ./cmd/plumego/...`
- `go run ./internal/checks/dependency-rules` (confirm no new forbidden imports)
- `go run ./internal/checks/module-manifests`

Docs Sync:
- Required: expand `docs/modules/x-openapi/README.md` with full API surface,
  CLI usage example (`plumego openapi generate --output openapi.json`),
  and boundary rules section.
- Update `docs/ROADMAP.md` Phase 13/14 to reflect CLI completion.

Done Definition:
- `plumego openapi generate` command exists and produces valid OpenAPI 3.1 JSON
  from a minimal test route set.
- x/openapi has serialization tests.
- Module primer covers the complete API surface and CLI usage.
- All boundary checks and x/openapi tests pass.

Notes:
- x/openapi is experimental; the CLI subcommand may be marked experimental in
  the help text until x/openapi is promoted to beta.
- Check `cmd/plumego/internal/scaffold/` for the existing subcommand registration
  pattern before adding the new subcommand.
- The RouteInfo → OpenAPI operation conversion is the primary integration point;
  confirm its current coverage in x/openapi tests before adding CLI wiring.
