# CLAUDE.md - plumego

Go module: `github.com/spcent/plumego`
Go version: `go 1.24.0`, toolchain `go1.24.4`
Main module policy: standard library only unless explicitly approved.

Use `AGENTS.md` as the hard-rule entrypoint. This file is a Claude-oriented
quick reference for the current repository shape; do not treat it as a second
source of truth.

---

## Repository Layout

### Stable Roots

Long-lived public packages with narrow responsibilities:

- `core`: app kernel, lifecycle, route and middleware attachment, server startup
- `router`: route matching, params, groups, reverse routing
- `contract`: transport primitives, response/error helpers, context accessors, binding
- `middleware`: transport-layer cross-cutting behavior only
- `security`: auth, password, header, input-safety, and abuse primitives
- `store`: persistence primitives for DB, cache, file, KV, idempotency
- `health`, `log`, `metrics`: support contracts and base implementations

Stable roots must not import `x/*`.

### Extension Roots

Current top-level extension roots:

- `x/ai`
- `x/data`
- `x/fileapi`
- `x/frontend`
- `x/gateway`
- `x/messaging`
- `x/observability`
- `x/resilience`
- `x/rest`
- `x/tenant`
- `x/websocket`

Subordinate packages stay under their family root, for example
`x/data/cache`, `x/gateway/discovery`, `x/messaging/mq`,
`x/messaging/pubsub`, `x/messaging/scheduler`, `x/messaging/webhook`,
`x/observability/devtools`, and `x/observability/ops`.

Beta families are tracked in `docs/EXTENSION_MATURITY.md` and
`specs/extension-maturity.yaml`. For landing zones, start with
`specs/task-routing.yaml` and prefer the primary family entrypoint.

### Other Directories

| Path | Purpose |
|---|---|
| `reference/standard-service` | Canonical application layout |
| `reference/production-service` | Production-oriented explicit wiring reference |
| `reference/with-*` | Scenario-specific reference apps |
| `reference/workerfleet` | Separate worker-oriented submodule reference |
| `cmd/plumego/` | CLI: dev server, dashboard, scaffold |
| `internal/checks/` | Boundary, manifest, workflow, and maturity validation |
| `internal/testutil/` | Shared test helpers |
| `specs/` | Machine-readable authority: routing, taxonomy, dependency rules, quality rules |
| `specs/change-recipes/` | Reusable prompt recipes for standard task shapes |
| `specs/stop-condition-handlers.yaml` | Resolution paths for every agent stop condition |
| `specs/request-flows.yaml` | Machine-readable request flow diagrams (module ownership per step) |
| `tasks/milestones/` | Milestone specs and execution plans |
| `tasks/cards/` | Incremental task cards |
| `docs/` | All control-plane documentation |
| `website/` | Documentation generation |

---

## Operating Defaults

- Preserve `net/http` compatibility.
- Keep the main module dependency-free beyond stdlib unless explicitly approved.
- Stable roots must not import `x/*`.
- No hidden globals, `init()` registration, or context service-locator patterns.
- Never log secrets, tokens, signatures, or private keys.
- Fail closed on auth, verification, and policy errors.
- Use timing-safe comparison for secret checks.
- Context accessors use `With{Type}` + `{Type}FromContext`; key types are
  unexported zero-value structs inlined at the call site.
- `contract` owns transport primitives only.
- Deprecated symbols removed in the same PR that replaces their last caller.
- Handler shape: `func(http.ResponseWriter, *http.Request)`
- Route wiring: one method, one path, one handler per line
- JSON decode: `json.NewDecoder(r.Body).Decode(...)`
- Error write: `contract.WriteError` with structured error codes
- Success write: `contract.WriteResponse`
- DI: constructor-based, visible in route wiring
- Middleware: `func(http.Handler) http.Handler`, transport-only, `next` called exactly once
- No service injection, business DTO assembly, or domain-policy branching in middleware
- Reference app: `reference/standard-service`

---

## Task Contract Defaults

When not stated explicitly, assume:

- one primary module per change
- no stable public API changes
- no new dependencies
- focused tests required for behavior changes
- docs sync only for implemented behavior, API, config, security, lifecycle, or
  boundary changes

### Stop Conditions

Stop in analysis mode (do not edit) when:

- owning module is unclear
- task would force a stable root to import `x/*`
- task needs a stable public API change that was not requested
- task needs a new dependency that was not approved
- task is broad but lacks acceptance criteria or validation commands
- a spec, module manifest, and local pattern conflict in a behavior-changing way

When a stop condition is hit, read `specs/stop-condition-handlers.yaml` for the
deterministic resolution path — each condition maps to concrete next steps,
an escalation action, and unblock criteria.

---

## Preflight Checklist

Complete before editing any file:

```text
Owning module:
Target module.yaml read:
In-scope paths:
Out-of-scope paths:
Public API impact: none / yes
Dependency impact: none / yes
Behavior impact: none / yes
Security impact: none / yes
Docs impact: none / yes
Validation plan:
```

---

## Symbol Change Protocol

When removing, renaming, or changing an exported symbol:

1. Enumerate all callers: `rg -n --glob '*.go' 'SymbolName' .`
2. Address every site in the same change (migrate, update, or explicit discard).
3. Re-run the same search — deletion leaves zero results; rename leaves no old name.
4. Update tests in the same commit.
5. `go build ./...` and required tests must pass before handoff.

---

## Validation

### Default Order

1. Run target module checks from `<module>/module.yaml`.
2. Run boundary and manifest checks.
3. Run repo-wide gates for code-bearing, cross-module, or release-relevant changes.

### Boundary and Manifest Checks

```bash
go run ./internal/checks/dependency-rules
go run ./internal/checks/agent-workflow
go run ./internal/checks/module-manifests
go run ./internal/checks/reference-layout
go run ./internal/checks/public-entrypoints-sync
```

`public-entrypoints-sync` verifies that every entry in each module.yaml
`public_entrypoints` list exists as an exported Go symbol in that module's
source. Run it after any symbol rename or removal to prevent stale declarations.

### Repo-Wide Gates

```bash
make gates
```

`make gates` mirrors CI: boundary checks, `go vet ./...`, non-mutating
`gofmt -l .`, race tests, and normal tests. Run `gofmt -w <paths>` before the
gate if formatting is needed.

### Gate Selection (lightest sufficient profile)

| Change type | Gate |
|---|---|
| Docs only | Check accuracy, links, authority order. No Go gates unless code/config changed. |
| Single module behavior | `go test -race ./<module>/...`, `go vet ./<module>/...` |
| Stable root change | Single-module gates + dependency-rules + agent-workflow + module-manifests |
| router / middleware / security | Module gates + dependency-rules; confirm negative-path and ordering tests |
| Cross-module, public API, release | Full `make gates` + extension maturity/beta-evidence + `deprecation-inventory -strict` |

---

## Milestone Workflow

Milestones are human-scoped, single-PR execution units in
`tasks/milestones/active/M-NNN.md`.

When executing a milestone:

- read every file listed under Context before editing
- stay inside Affected Modules and Out of Scope
- follow Tasks in order; phase order is sequential
- use the branch named in the spec (`milestone/M-NNN-<slug>`)
- record blockers in the spec instead of widening scope silently
- run full acceptance criteria before pushing
- package the PR with `docs/github-workflows/milestone-pr-template.md`

Scaffold: `make new-milestone`, `make new-plan`, `make new-card`, `make new-verify`.
Use milestones for multi-step work; use daily implementation prompts for small fixes.

Additional targets:
- `make run-card C=active/NNNN-slug` — validate card, generate bundle, launch execution
- `make milestone-status M=active/M-NNN` — display phase checkpoint progress

---

## Docs Sync Targets

Update only the docs affected by behavior, public API, config, security
semantics, lifecycle, or boundary changes. Common targets:

- `README.md`
- `README_CN.md`
- `AGENTS.md`
- `CLAUDE.md`
- `env.example`

Also update module primers, `docs/EXTENSION_MATURITY.md`, `docs/ROADMAP.md`, or
`docs/stable-api/` when the change specifically affects that surface.

Document implemented behavior only.

---

## Anti-Patterns

Do not introduce:

- Stable package imports of `x/*`
- New dependencies without approval
- Context-based service lookup or package-level mutable registries
- `init()` side-effect registration
- Route auto-discovery or reflection-based wiring
- Middleware that builds business DTOs or injects services
- Ad hoc JSON response helpers or per-feature error envelopes
- New handler signatures
- New panic-only constructors for fallible behavior
- Generic `utils` packages
- Compatibility wrappers without a removal plan
- New broad root names: `ai`, `tenant`, `net`, `pubsub`, `rest`, `validator`,
  `utils`, `frontend`
