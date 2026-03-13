# CLAUDE.md — plumego

> Go 1.24+ | Toolchain: go1.24.4 | Main module: standard-library only

See `AGENTS.md` for the full operational guide (boundaries, rules, quality gates, workflow).

Repository planning and future restructuring should also follow:

- `docs/CANONICAL_STYLE_GUIDE.md`
- `docs/architecture/AGENT_FIRST_REPO_BLUEPRINT.md`
- `specs/repo.yaml`
- `specs/dependency-rules.yaml`
- target `<module>/module.yaml`
- `reference/standard-service`

## Claude-Specific Notes

**Style authority:** `docs/CANONICAL_STYLE_GUIDE.md`

**Required checks before final output:**
```bash
go run ./internal/checks/dependency-rules
go run ./internal/checks/module-manifests
go run ./internal/checks/reference-layout
go test -timeout 20s ./...
go vet ./...
gofmt -w .
```

**Recommended when relevant:**
```bash
go test -race ./...
```

**Canonical defaults:**
- Handler: `func(w http.ResponseWriter, r *http.Request)`
- Error path: `contract.WriteError`
- Middleware: `func(http.Handler) http.Handler`, transport-only
- Dependency flow: constructor injection, not context service-locator
- App shape: follow `reference/standard-service`

**Documentation sync targets:** `README.md`, `README_CN.md`, `AGENTS.md`, `CLAUDE.md`, `env.example`

Use the handler pattern from `docs/CANONICAL_STYLE_GUIDE.md` consistently within one flow.
