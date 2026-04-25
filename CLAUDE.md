# CLAUDE.md - plumego

Go module: `github.com/spcent/plumego`
Go version: `go 1.24.0`, toolchain `go1.24.4`
Main module policy: standard library only unless explicitly approved.

Use `AGENTS.md` as the hard-rule entrypoint. For implementation detail, read:

1. `docs/CODEX_WORKFLOW.md`
2. `docs/CANONICAL_STYLE_GUIDE.md`
3. `docs/architecture/AGENT_FIRST_REPO_BLUEPRINT.md`
4. `specs/repo.yaml`
5. `specs/task-routing.yaml`
6. `specs/dependency-rules.yaml`
7. target `<module>/module.yaml`
8. `reference/standard-service`

## Claude Defaults

- Preserve `net/http` compatibility.
- Keep stable roots out of `x/*`.
- Use stdlib-shaped handlers: `func(http.ResponseWriter, *http.Request)`.
- Use `contract.WriteError` and `contract.WriteResponse`.
- Keep middleware transport-only.
- Use explicit constructor or route-wiring DI; do not use context service
  location.
- Prefer analysis mode over editing when ownership, scope, or API policy is
  unclear.

## Checks

Boundary checks:

```bash
go run ./internal/checks/dependency-rules
go run ./internal/checks/agent-workflow
go run ./internal/checks/module-manifests
go run ./internal/checks/reference-layout
```

Repo-wide gates when code changes or release confidence is needed:

```bash
go test -race -timeout 60s ./...
go test -timeout 20s ./...
go vet ./...
gofmt -w .
```

Docs sync targets: `README.md`, `README_CN.md`, `AGENTS.md`, `CLAUDE.md`,
`docs/ROADMAP.md`, `env.example`.
