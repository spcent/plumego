# Card 1542

Milestone: M-014
Recipe: specs/change-recipes/add-package.yaml
Priority: P1
State: done
Primary Module: cmd/plumego
Owned Files:
- `cmd/plumego/commands/spec.go`
- `cmd/plumego/commands/spec_test.go`
- `reference/with-rest/Makefile`

Goal:
- Add `plumego generate spec` as a subcommand of the existing generate command,
  wiring x/openapi.Generator to produce an OpenAPI 3.1 document from a running
  plumego app's route table.

Scope:
- Add "spec" as a recognised generate type in cmd/plumego/commands/generate.go
  dispatch or create cmd/plumego/commands/spec.go with a SpecCmd.
- Flags: --output (file path or "-" for stdout), --format (json|yaml, default json),
  --app (package path of the app to introspect, default ".").
- Implementation: import x/openapi; call app.Routes() to get []router.RouteInfo;
  load Op hints from a plumego.spec.yaml file in the project root if present;
  generate and write the document.
- Write spec_test.go as a CLI integration test: run `plumego generate spec
  --output /dev/null` against reference/with-rest and confirm exit 0.
- Add a `spec` Makefile target to reference/with-rest/Makefile:
  `plumego generate spec --output openapi.yaml`.

Non-goals:
- Do not auto-register Op hints; hints come from an explicit plumego.spec.yaml.
- Do not add x/openapi as a dependency of the main cmd/plumego module; use
  a subprocess invocation or plugin hook if needed to keep the main module
  dependency-free.
- Do not validate the generated spec for OpenAPI compliance.

Files:
- `cmd/plumego/commands/spec.go`
- `cmd/plumego/commands/spec_test.go`
- `reference/with-rest/Makefile`

Tests:
- `go test -race -timeout 60s ./cmd/plumego/...`
- `go vet ./cmd/plumego/...`
- `go build ./reference/with-rest/...`

Docs Sync:
- Update reference/with-rest/README.md with `make spec` usage.
- Update cmd/plumego/README.md to list the new generate spec subcommand.

Done Definition:
- `plumego generate spec --output /dev/null` exits 0 against reference/with-rest.
- --format yaml produces YAML output, --format json produces JSON output.
- reference/with-rest Makefile has a `spec` target.
- `go test ./cmd/plumego/...` passes including the integration test.

Outcome:
- Added `plumego generate spec` dispatch under the existing generate command
  with `--output`, `--format`, and `--app` flags.
- Kept `cmd/plumego` free of a direct `x/openapi` dependency by running a
  temporary helper module in the target app context. The helper imports
  `x/openapi`, registers the canonical app routes, reads optional
  `plumego.spec.yaml` operation hints, and writes JSON or YAML output.
- Added CLI integration coverage against `reference/with-rest`, including
  `/dev/null` output plus JSON and YAML file generation.
- Added `reference/with-rest` `make spec` support and documented the command in
  the CLI and reference README files.
- Validation passed with `cmd/plumego` race tests, `cmd/plumego` vet,
  `reference/with-rest` build, dependency-rules, module-manifests,
  reference-layout, agent-workflow, `gofmt -l .`, and `git diff --check`.
