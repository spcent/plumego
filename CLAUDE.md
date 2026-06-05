# CLAUDE.md

Entry point for all agent types (Claude Code, Codex, Cursor, …). It delegates to
AGENTS.md via the @-import below. **AGENTS.md is authoritative — edit it, not this file.**

---

## Codebase Snapshot (v1.1.0 · Go 1.26+ · single module)

`github.com/spcent/plumego` — a stdlib web toolkit for Go. Stable roots have zero
runtime deps; `x/*` extensions may use external deps.

| Layer | Packages | Notes |
|---|---|---|
| Stable roots (GA) | `core` `router` `contract` `middleware` `security` `store` `health` `log` `metrics` | stdlib-only; no `x/*` imports |
| Extensions (beta) | `x/frontend` `x/gateway` `x/messaging` `x/observability` `x/rest` `x/tenant` `x/websocket` | production-ready, caveats |
| Extensions (experimental) | `x/ai` `x/data` `x/fileapi` `x/openapi` `x/resilience` `x/rpc` `x/validate` | APIs may change |
| Reference apps | `reference/*` (16) | each has own `go.mod` |
| Use-case apps | `use-cases/{workerfleet,cloud-vault,dbadmin,guardus}` | each has own `go.mod` |
| CLI | `cmd/plumego` | agent-assist, codegen, validation, scaffold |
| Validation | `internal/checks/` (11 programs) | via `make gates` |
| Docs site | `website/` | Astro; `make website-sync` |

### Key commands

```bash
make validate-diff   # minimal gate for current diff (preferred)
make gates           # Go gates: vet, race tests, boundary checks, coverage ≥70%
make website-gates   # docs site: content/API checks + static build (slow)
make fmt             # gofmt -w (run before gates)
make website-sync    # regenerate website/src/generated/
make new-milestone N=NNN TITLE="..."     # scaffold milestone
make run-card C=active/NNNN-slug         # execute task card
make bundle TASK=<recipe> MODULE=<path>  # generate task bundle
```

### Start here

- **Canonical shape:** `reference/standard-service` (explicit middleware, routing, stdlib only)
- **Style:** `docs/reference/canonical-style-guide.md`
- **Boundaries:** each `<module>/module.yaml` + `specs/dependency-rules.yaml`
- **Task routing:** `specs/task-routing.yaml` (task types → context packages + start files)
- **Active work:** `tasks/milestones/active/`, `tasks/cards/active/`

---

@AGENTS.md
