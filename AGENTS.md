# AGENTS.md - plumego

Operational guide for AI coding agents working in `github.com/spcent/plumego`.

Plumego is an agent-first Go toolkit built on the standard-library HTTP model.
Keep the stable surface small, ownership obvious, and every change easy to
review and reverse.

## 1. Authority

Use this file for hard rules and validation order. Use the rest of the control
plane for detail:

1. `docs/CODEX_WORKFLOW.md` - operating modes and task recipes
2. `docs/CANONICAL_STYLE_GUIDE.md` - code shape and canonical examples
3. `docs/architecture/AGENT_FIRST_REPO_BLUEPRINT.md` - repository layout
4. `specs/repo.yaml` - current stable and extension roots
5. `specs/task-routing.yaml` - task-to-entrypoint routing
6. `specs/extension-taxonomy.yaml` - primary and subordinate extension families
7. `specs/package-hotspots.yaml` - ambiguous package landing zones
8. `specs/dependency-rules.yaml` - import boundaries
9. target `<module>/module.yaml` - module-local scope and checks
10. `reference/standard-service` - canonical application wiring

When guidance conflicts, follow this order:

1. security and boundary rules in this file
2. `docs/CANONICAL_STYLE_GUIDE.md`
3. machine-readable specs
4. existing local patterns in touched files

## 2. Non-Negotiables

- Preserve `net/http` compatibility.
- Keep the main module dependency-free beyond the standard library unless
  explicitly approved.
- Stable roots must not import `x/*`.
- Do not introduce hidden globals, `init()` registration, or context
  service-locator patterns.
- Never log secrets, tokens, signatures, or private keys.
- Fail closed on auth, verification, and policy errors.
- Use timing-safe comparison for secret checks.
- Context accessors use `With{Type}` plus `{Type}FromContext`; key types are
  unexported zero-value structs inlined at the call site.
- `contract` owns transport primitives only: response/error helpers, request
  metadata, context accessors, and binding helpers.
- Deprecated symbols must be removed in the same PR that replaces their last
  caller. Do not leave dead wrappers behind.
- Keep one canonical success-response path and one canonical error-construction
  path per layer.

## 3. Where To Work

Stable roots are long-lived public packages:

- `core`, `router`, `contract`, `middleware`, `security`, `store`, `health`,
  `log`, `metrics`

Extension roots live under `x/*`; use `specs/task-routing.yaml` to start at the
primary family rather than a subordinate package.

Default routing:

- kernel, lifecycle, routing, transport contracts, transport middleware, auth
  primitives, storage primitives: stable root
- product capability, protocol adaptation, business feature, extension behavior:
  `x/*`
- application wiring, bootstrap, DI, route registration:
  `reference/standard-service` or `cmd/plumego/internal/scaffold`
- architecture rules, boundaries, quality gates: `specs/`
- execution plans and task sequencing: `tasks/cards/` or `tasks/milestones/`

Hard boundaries:

- `core` is the app kernel, not a feature catalog.
- `router` owns matching, params, groups, and reverse routing.
- `middleware` stays transport-only; no business DTO assembly or service
  injection.
- `health` owns health models/readiness helpers, not HTTP handler ownership.
- Tenant logic belongs in `x/tenant`, not stable `middleware` or `store`.
- Tracing, sampling, collectors, and metric export wiring belong in
  `x/observability`, not `contract`.
- Session lifecycle belongs in `x/tenant` or an owning extension, not
  `contract`.
- New library code must live under a stable root or `x/*`; avoid legacy broad
  roots such as `net`, `utils`, `validator`, `tenant`, `ai`, `rest`, or
  `pubsub`.

## 4. Canonical Defaults

- Handler shape: `func(http.ResponseWriter, *http.Request)`
- Route wiring: one method, one path, one handler per line
- JSON decode: `json.NewDecoder(r.Body).Decode(...)`
- Error write path: `contract.WriteError` with structured error codes
- Success write path: `contract.WriteResponse`
- DI: constructor-based and visible in route wiring
- Middleware: `func(http.Handler) http.Handler`, transport-only
- Reference app: `reference/standard-service`

Do not add one-off handler styles, response envelopes, error helper families, or
route registration idioms.

## 5. Working Contract

For normal implementation requests, assume:

- one primary module per change
- no stable public API changes
- no new dependencies
- focused tests for behavior changes
- docs sync only for behavior, API, config, security, lifecycle, or boundary
  changes

Use analysis mode and do not edit when ownership is unclear, the task is broad
without acceptance criteria, a new dependency is required, or a stable public API
change was not requested.

For review requests, do not patch unless explicitly asked. Findings come first,
ordered by severity, with boundary violations, regressions, hidden coupling, and
missing tests prioritized.

## 6. Change Rules

- Keep changes minimal and scoped to one primary module when possible.
- Read the target module manifest before editing module behavior.
- Preserve stable public APIs unless explicitly asked to change them.
- If a breaking change is unavoidable, include migration notes.
- Prefer standard-library solutions and existing local patterns.
- Add or update tests next to changed behavior.
- Do not refactor unrelated files opportunistically.

## 7. Symbol Changes

### 7.1 Completeness Protocol for Symbol Changes

When removing, renaming, or changing the behavior of an exported symbol:

1. Enumerate all sites first with `rg -n --glob '*.go' 'SymbolName' .`.
2. Address every site in the same change: migrate, update assertions, or add an
   explicit `_ =` discard with a TODO.
3. Re-run the same search. For deletion, the result must be empty. For rename,
   the old name must not appear.
4. Update tests in the same commit.
5. Before handoff, `go build ./...` and the required tests must pass.

No caller may silently discard a newly returning function without an explicit
discard.

## 8. Validation

Default order:

1. Run target module checks from `<module>/module.yaml`.
2. Run boundary and manifest checks.
3. Run repo-wide gates when the change is code-bearing, cross-module, or release
   relevant.

Boundary and manifest checks:

```bash
go run ./internal/checks/dependency-rules
go run ./internal/checks/agent-workflow
go run ./internal/checks/module-manifests
go run ./internal/checks/reference-layout
```

Repo-wide gates:

```bash
make gates
```

`make gates` mirrors CI: boundary checks, `go vet ./...`, a non-mutating
`gofmt -l .` check, race tests, and normal tests. Run `gofmt -w <paths>` before
the gate if formatting is needed.

Extra focus:

- routing: static, params, groups, and reverse routing
- middleware: ordering and error paths
- security: invalid token/signature negative tests
- tenant: quota, policy, and isolation
- store: concurrency and persistence correctness

Do not expand `specs/check-baseline/` casually.

## 9. Docs Sync

Update affected docs when behavior, public API, config, security semantics,
lifecycle behavior, or boundaries change:

- `README.md`
- `README_CN.md`
- `AGENTS.md`
- `CLAUDE.md`
- `docs/ROADMAP.md`
- `env.example`

Document implemented behavior only.

## 10. Milestones

Milestones are human-scoped, single-PR execution units in
`tasks/milestones/active/M-NNN.md`.

When executing a milestone:

- read every file listed under Context before editing
- stay inside Affected Modules and Out of Scope
- follow Tasks in order; phase order is always sequential
- use the branch named in the spec, usually `milestone/M-NNN-<slug>`
- record blockers in the spec instead of widening scope silently
- run the full acceptance criteria before pushing
- package the PR with `docs/github-workflows/milestone-pr-template.md`

Use `docs/MILESTONE_PIPELINE.md`, `tasks/milestones/README.md`, and the milestone
templates for detailed planning, card, verify, and PR packaging rules.
