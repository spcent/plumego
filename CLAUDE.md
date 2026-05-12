# CLAUDE.md - plumego

Go module: `github.com/spcent/plumego`
Go version: `go 1.24.0`, toolchain `go1.24.4`
Main module policy: standard library only unless explicitly approved.

Use `AGENTS.md` as the hard-rule entrypoint. For implementation detail, read:

1. `docs/CODEX_WORKFLOW.md`
2. `docs/CANONICAL_STYLE_GUIDE.md`
3. `docs/agent-first.md`
4. `docs/architecture/AGENT_FIRST_REPO_BLUEPRINT.md`
5. `docs/architecture/core-boundary.md`
6. `docs/architecture/extension-boundary.md`
7. `specs/repo.yaml`
8. `specs/task-routing.yaml`
9. `specs/dependency-rules.yaml`
10. target `<module>/module.yaml`
11. `reference/standard-service`

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
make gates
```

`make gates` mirrors CI and uses `gofmt -l .` as a check. Format touched Go
files before running it.

Docs sync targets: `README.md`, `README_CN.md`, `AGENTS.md`, `CLAUDE.md`,
`docs/ROADMAP.md`, `env.example`.
