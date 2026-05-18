# Card 1592

Milestone: M-019
Recipe: specs/change-recipes/update-docs.yaml
Priority: P3
State: active
Primary Module: docs
Owned Files:
- `docs/EXTENSION_AUTHORING.md`

Goal:
- Write docs/EXTENSION_AUTHORING.md as the canonical guide for community
  developers creating plumego-compatible extensions, covering the schema
  contract, add command usage, a complete community-extension.yaml example,
  and x/rpc as a worked example.

Scope:
- Write docs/EXTENSION_AUTHORING.md with these sections:
  1. Overview: what a community extension is and what it is not (no auto-wiring,
     no globals, explicit constructor).
  2. Schema contract: table of all required fields in
     specs/community-extension.schema.yaml with descriptions and valid values.
  3. community-extension.yaml example: fully filled YAML using x/rpc as the
     model (module_path, status, handler_shape, test_commands, forbidden_imports,
     no_init_side_effects, no_globals, owner).
  4. Compliance checklist: 8-item checklist matching schema fields; each item
     has a pass/fail example.
  5. Publishing your extension: steps to publish on pkg.go.dev, make
     community-extension.yaml discoverable, and verify with
     `plumego add <your-module-path>`.
  6. Validation: how to run `go run ./internal/checks/community-extension` locally
     and interpret the output.
  7. Maintaining compatibility: following plumego's extension maturity model
     (experimental → beta → ga) and what API changes are breaking.

Non-goals:
- Do not promise a hosted registry in this guide (not yet built).
- Do not document internal plumego implementation details.
- Do not duplicate content from docs/EXTENSION_MATURITY.md.

Files:
- `docs/EXTENSION_AUTHORING.md`

Tests:
- `go run ./internal/checks/module-manifests`
- `gofmt -l .`

Docs Sync:
- Update docs/EXTENSION_MATURITY.md to reference docs/EXTENSION_AUTHORING.md
  in the "Publishing extensions" section.
- Update README.md to mention community extensions and link to the authoring guide.

Done Definition:
- docs/EXTENSION_AUTHORING.md exists with all seven sections.
- x/rpc community-extension.yaml example is syntactically valid YAML.
- docs/EXTENSION_MATURITY.md references the new guide.
- README.md links to docs/EXTENSION_AUTHORING.md.
- `gofmt -l .` outputs nothing.

Outcome:
-
