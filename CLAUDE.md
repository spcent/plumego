# CLAUDE.md

This file is the Claude Code entry point. It delegates to AGENTS.md via the
@-import below so that the same guidance applies to all agent types (Claude Code,
Codex, Cursor, etc.) without duplicating content. AGENTS.md is the authoritative
source; edit it, not this file.

---

## Codebase Snapshot (v1.1.0 · Go 1.26+ · single module)

**What it is:** `github.com/spcent/plumego` — a standard-library web toolkit for Go.
Zero runtime dependencies in stable roots. Extensions under `x/*` may use external deps.

### Module layout

| Layer | Packages | Notes |
|---|---|---|
| Stable roots (GA) | `core` `router` `contract` `middleware` `security` `store` `health` `log` `metrics` | stdlib-only; no `x/*` imports |
| Extensions (beta) | `x/frontend` `x/gateway` `x/messaging` `x/observability` `x/rest` `x/tenant` `x/websocket` | Production-ready with caveats |
| Extensions (experimental) | `x/ai` `x/data` `x/fileapi` `x/openapi` `x/resilience` `x/rpc` `x/validate` | APIs may change |
| Reference apps | `reference/standard-service` … `reference/with-websocket` (16 total) | Each has its own `go.mod` |
| Use-case apps | `use-cases/workerfleet` | MongoDB; own `go.mod` |
| CLI tool | `cmd/plumego` | Agent-assist, codegen, validation, scaffold |
| Validation | `internal/checks/` (14 programs) | Run via `make gates` |
| Docs website | `website/` | Astro; regenerate with `make website-sync` |

### Key commands

```bash
make validate-diff          # minimal gate profile for current diff (preferred)
make gates                  # full CI: vet, race tests, boundary checks, coverage ≥70%
make fmt                    # gofmt -w all Go files (run before gates)

go run ./internal/checks/dependency-rules
go run ./internal/checks/cross-extension-deps
go run ./internal/checks/module-manifests

make new-milestone N=NNN TITLE="..."   # scaffold milestone
make run-card C=active/NNNN-slug       # execute task card
make bundle TASK=<recipe> MODULE=<path> # generate task bundle
make website-sync                       # regenerate website/src/generated/
```

### Start here

- **Canonical shape:** `reference/standard-service` — explicit middleware, routing, stdlib only
- **Style rules:** `docs/reference/canonical-style-guide.md`
- **Boundary rules:** each module's `<module>/module.yaml` + `specs/dependency-rules.yaml`
- **Task routing:** `specs/task-routing.yaml` maps task types to context packages and start files
- **Active work:** `tasks/milestones/active/` and `tasks/cards/active/`

---

@AGENTS.md
