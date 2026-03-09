# CLAUDE.md — plumego

> Go 1.24+ | Toolchain: go1.24.4 | Main module: standard-library only

See `AGENTS.md` for the full operational guide (boundaries, rules, quality gates, workflow).

## Claude-Specific Notes

**Style authority:** `docs/CANONICAL_STYLE_GUIDE.md`

**Required checks before final output:**
```bash
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

**Documentation sync targets:** `README.md`, `README_CN.md`, `AGENTS.md`, `CLAUDE.md`, `env.example`

Use the handler pattern from `docs/CANONICAL_STYLE_GUIDE.md` consistently within one flow.
