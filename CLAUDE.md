# CLAUDE.md

Authoritative agent guide — detailed rules in `AGENTS.md`.

## Codebase (v1.1.0 · Go 1.26+ · single module)

`github.com/spcent/plumego` — stdlib-only web toolkit. Stable roots: no deps. `x/*`: external deps allowed.

### Module layout

| Layer | Packages | Notes |
|---|---|---|
| Stable roots (GA) | `core` `router` `contract` `middleware` `security` `store` `health` `log` `metrics` | stdlib-only; no `x/*` imports |
| Extensions (beta) | `x/frontend` `x/gateway` `x/messaging` `x/observability` `x/rest` `x/tenant` `x/websocket` | Production-ready with caveats |
| Extensions (experimental) | `x/ai` `x/data` `x/fileapi` `x/openapi` `x/resilience` `x/rpc` `x/validate` | APIs may change |
| Reference apps | `reference/standard-service` … `reference/with-websocket` (16) | Each has own `go.mod` |
| Use-case apps | `use-cases/cloud-vault` `use-cases/dbadmin` `use-cases/guardus` `use-cases/workerfleet` | Each has own `go.mod` |
| CLI | `cmd/plumego` | Agent-assist, codegen, validation, scaffold |
| Validation | `internal/checks/` (12 programs) | Via `make gates` |
| Docs website | `website/` | Astro; `make website-sync` |

### Key commands

```bash
make validate-diff   # minimal gate for current diff (preferred)
make gates           # full Go gates: vet, race, boundary, coverage ≥70%
make fmt             # gofmt -w (run before gates)
make website-gates   # docs website checks + build (slow)
make website-sync    # regenerate website/src/generated/
```

Scaffold: `make new-milestone`, `make run-card`, `make bundle` — see `AGENTS.md §8`.

### Start here

- Canonical shape: `reference/standard-service`
- Style: `docs/reference/canonical-style-guide.md`
- Boundaries: `<module>/module.yaml` + `specs/dependency-rules.yaml`
- Task routing: `specs/task-routing.yaml`
- Active work: `tasks/milestones/active/`, `tasks/cards/active/`

---

@AGENTS.md
