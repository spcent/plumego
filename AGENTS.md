# AGENTS.md — plumego

Operational guide for AI coding agents working in `github.com/spcent/plumego`.

## 1. Goal

Plumego is an agent-first Go toolkit built on the standard-library HTTP model.
Optimize for:

- clear module ownership
- explicit control flow
- small reversible changes
- minimal search radius
- one canonical implementation path

## 2. Non-Negotiables

- Preserve `net/http` compatibility.
- Keep the main module dependency-free (stdlib only) unless explicitly approved.
- Do not blur stable-module boundaries.
- Do not introduce hidden globals, `init()` registration, or context service-locator patterns.
- Never log secrets, tokens, signatures, or private keys.
- Fail closed on auth, verification, and policy errors.
- Use timing-safe comparison for secret checks.
- Context accessor pairs follow `With{Type}` + `{Type}FromContext` order; context
  key types are unexported zero-value structs inlined at the call site — no package-level variable.
- `contract` contains transport primitives only. Tracing infrastructure, session
  lifecycle management, and metric collection do not belong in `contract`.
- Deprecated symbols must be removed in the same PR that replaces their last
  caller. Do not leave dead wrappers behind.
- One canonical success-response path per layer; one canonical error-construction
  path per layer. Do not add per-feature response helpers or per-scenario error constructors.

## 3. Read Order

Use this default path before making changes:

1. `docs/CANONICAL_STYLE_GUIDE.md`
2. `docs/architecture/AGENT_FIRST_REPO_BLUEPRINT.md`
3. `specs/repo.yaml`
4. `specs/task-routing.yaml`
5. `specs/extension-taxonomy.yaml`
6. `specs/package-hotspots.yaml`
7. `specs/dependency-rules.yaml`
8. target `<module>/module.yaml`
9. `reference/standard-service`

For staged future work and sequencing, also read:

- `docs/ROADMAP.md`
- `specs/change-recipes/*`
- `tasks/cards/*`

When guidance overlaps, follow:

1. security and boundary rules in this file
2. `docs/CANONICAL_STYLE_GUIDE.md`
3. machine-readable repo specs
4. existing local patterns in touched files

## 4. Canonical Defaults

- Handler shape: `func(http.ResponseWriter, *http.Request)`
- Route wiring: one method + path + handler per line
- JSON decode: `json.NewDecoder(r.Body).Decode(...)`
- Error write path: `contract.WriteError` with structured error codes
- DI: constructor-based and explicit in route wiring
- Middleware: `func(http.Handler) http.Handler`, transport-only responsibility
- Reference app: `reference/standard-service` is the only canonical application layout

## 5. Agent Navigation Rules

Five rules determine where to work. Read this before scanning the entrypoints list.

| Intent | Destination |
|---|---|
| Change kernel, lifecycle, route structure, transport contracts, transport middleware, auth primitives, storage primitives | stable root |
| Change product capability, business feature, protocol adaptation, extension behavior | `x/*` |
| Change application wiring, bootstrap, DI, route registration | `reference/standard-service` or `internal/scaffold` |
| Change architecture rules, boundary definitions, quality gates | `specs/` |
| Change execution plan, work items, task sequencing | `tasks/cards/` |

**Stable roots (9):** `core`, `router`, `contract`, `middleware`, `security`, `store`, `health`, `log`, `metrics`

**x/* primary families (10):** `x/tenant`, `x/fileapi`, `x/messaging`, `x/gateway`, `x/rest`, `x/websocket`, `x/frontend`, `x/observability`, `x/data`, `x/ai`

Always start at a primary family, not a subordinate (`x/mq`, `x/pubsub`, `x/ops`, `x/cache`, `x/devtools`, etc.).
See `specs/task-routing.yaml` for the full routing table and detailed entrypoints.

## 6. Module Boundaries

Stable library roots:

- `core`, `router`, `contract`, `middleware`, `security`, `store`, `health`, `log`, `metrics`

Extension roots:

- `x/*` for optional or fast-evolving capabilities

Hard rules:

- Stable roots must not depend on `x/*`.
- `core` is the app kernel, not a feature catalog.
- `router` owns matching, params, groups, and reverse routing.
- `middleware` stays transport-only; never hide business DTO assembly or service injection there.
- `contract` owns transport contracts and response/error helpers, not protocol gateway families,
  observability infrastructure, or session lifecycle management.
  Tracing subsystems (Tracer, Span, Collector, Sampler) belong in `x/observability`.
  Session management (SessionStore, SessionValidator, RefreshManager) belongs in `x/tenant`.
- `health` owns models/readiness helpers, not HTTP handler ownership.
- Tenant-aware logic belongs in `x/tenant`, not stable `middleware` or stable `store`.
- Stable `middleware` must not grow tenant resolution, tenant policy, or tenant quota behavior.
- Stable `store` must not grow tenant-aware adapters or tenant-specific storage policy.
- New library code must live under a stable root or `x/*`; avoid broad legacy roots such as `net`, `utils`, `validator`, `tenant`, `ai`, `rest`, `pubsub`.

Task entrypoint defaults:

- HTTP endpoint work: start with the style guide, `reference/standard-service/internal/app/routes.go`, `contract`, and `router`.
- Middleware work: start with `middleware/module.yaml` and `docs/modules/middleware/README.md`.
- Security work: start with `security/module.yaml` and `docs/modules/security/README.md`.
- Store work: start with `store/module.yaml` and `docs/modules/store/README.md`.
- Gateway or edge transport work: start with `x/gateway` (includes service discovery and IPC).
- Resource API standardization: start with `x/rest`.
- Messaging work: start with `x/messaging` (not `x/mq` or `x/pubsub` directly).
- Tenant work: start with `x/tenant` and `docs/architecture/X_TENANT_BLUEPRINT.md`.
- WebSocket transport work: start with `x/websocket`.
- File upload/download/storage work: start with `x/fileapi`.
- Admin or observability surfaces: start with `x/observability` (includes ops and devtools), not `health`.
- AI capability work: start with `x/ai`.
- Data topology work (sharding, rw-split, cache): start with `x/data`.

## 7. Change Rules

- Keep changes minimal and scoped to one primary module when possible.
- Preserve stable public APIs unless explicitly asked to change them.
- If a breaking change is unavoidable, add migration notes.
- Prefer standard-library solutions over new abstractions.
- Add or update tests next to changed behavior.
- Do not invent one-off handler styles, response envelopes, or helper families for a single feature.

### 7.1 Completeness Protocol for Symbol Changes

When a card removes, renames, or changes the behavior of any exported symbol,
follow these steps before writing any code:

1. **Enumerate all sites first.**
   Run `grep -rn 'SymbolName' . --include='*.go'` and collect the full list.
   Do not start editing until you have the complete picture.

2. **Address every site.**
   For each file in the list, decide: migrate, update assertion, or add an
   explicit `_ =` discard with a TODO comment. No site is left as-is.

3. **Verify zero residual references.**
   After editing, re-run the same grep. For a deletion the result must be
   empty. For a rename the old name must not appear.

4. **Include test updates in the same commit.**
   If a response format changes (e.g. envelope wrapping), update every test
   that asserts on the old format in the same atomic commit. A passing build
   and a failing test suite is not done.

5. **A card is done only when:**
   - `go build ./...` exits 0
   - `go test ./...` exits 0 (or targeted packages per the card)
   - The symbol-removal grep is empty (for deletions)
   - No caller silently discards a now-returning function without an explicit `_`

This protocol addresses the pattern seen in cards 0503/0509 where exported
key *types* were removed but deprecated *function wrappers* that delegated to
them were left behind.

## 8. Validation Order

Default order:

1. Run the target module tests from `<module>/module.yaml`.
2. Run boundary and manifest checks.
3. Run repo-wide gates before final handoff.

Required repo-wide gates:

```bash
go run ./internal/checks/dependency-rules
go run ./internal/checks/agent-workflow
go run ./internal/checks/module-manifests
go run ./internal/checks/reference-layout
go test -race -timeout 60s ./...
go test -timeout 20s ./...
go vet ./...
gofmt -w .
```

Extra checks by change type:

- Routing: static, param, group, and reverse-routing coverage
- Middleware: ordering and error-path coverage
- Security: invalid token/signature negative tests
- Tenant: quota, policy, and isolation tests
- Store: concurrent access and persistence correctness tests

`specs/check-baseline/` contains temporary migration debt baselines. Reduce them; do not expand them casually.
If a baseline file is empty, treat the file itself as migration debt and prefer removing the placeholder once the corresponding check can tolerate a missing baseline file.

## 9. Docs Sync

Update these when behavior, public API, config, security semantics, lifecycle behavior, or boundaries change:

- `README.md`
- `README_CN.md`
- `AGENTS.md`
- `CLAUDE.md`
- `docs/ROADMAP.md`
- `env.example`

## 10. Working Loop

1. Identify the target layer: stable root or `x/*`.
2. Read the canonical sources in Section 3.
3. If a matching repo-native task card exists, follow it.
4. Confirm the owning module manifest and validation commands.
5. Make the smallest coherent change.
6. Add or update focused tests.
7. Run validation in the order from Section 8.
8. Sync docs only for implemented behavior changes.

## 11. Milestone Execution Protocol

Milestones are multi-step, single-PR scopes authored by a human and executed
autonomously by Codex in `--yolo` mode. When invoked with a milestone spec:

### Entry

1. The spec file is in `tasks/milestones/active/M-NNN.md`.
2. Read every file listed under **Context** in the spec before touching code.
3. The spec's **Tasks** list is the execution plan — follow it in order.
4. Keep changes scoped to the **Affected Modules** listed in the spec.

### Execution Rules

- One commit per logical task step when the step produces a stable intermediate state.
- All commits go to the branch named in the spec: `milestone/M-NNN-<slug>`.
- Create that branch from `main` if it does not exist.
- Do not open sub-tasks outside the spec's **Tasks** list. If a blocker is discovered,
  stop, document it at the bottom of the spec under `## Blocker`, and push the branch.

### Validation Before Push

Run the full gate sequence from Section 8. All commands must exit 0.
`gofmt -l .` must produce no output.
Fix any failures before pushing. Do not push a red state.

If the local pre-push hook is installed (`make setup-hooks`), it runs the full
gate suite automatically before `git push` on a `milestone/*` branch and blocks
the push on failure. This mirrors CI and catches issues before they reach GitHub.

### Commit Convention

```
feat(<module>): <short description> [M-NNN]
```

For the final commit after all tasks pass:

```
milestone(M-NNN): <spec title>

- <bullet per completed task>
- all quality gates pass
```

### Parallel Task Execution

The spec's **Tasks** section divides work into named phases:

- **Sequential phase:** execute steps in listed order; each step must complete before the next.
- **Parallel phase:** all checklist items are independent; run them concurrently.
- Phase order is always sequential (Phase 1 before Phase 2, etc.).

When no phase labels are present, treat all tasks as sequential.

### PR

After pushing the branch, open a PR targeting `main`:

- **Title:** `milestone(M-NNN): <spec title>`
- **Body:** fill `docs/github-workflows/milestone-pr-template.md` with actual content:
  - Approach section: describe decisions made, not just files changed.
  - Architecture Decisions table: confirm each spec decision was followed; flag deviations.
  - Scope boundary table: list every changed file with its module.
  - Gate output: paste verbatim terminal output inside the `<details>` block.
- Move the spec from `tasks/milestones/active/` to `tasks/milestones/done/`
  and append an `## Outcome` section with PR number and gate output summary.
- Update `tasks/milestones/ROADMAP.md`: set status to `[PR]`.

### Human Review Checkpoint

The PR is the only human gate. Codex does not wait for confirmation at any
intermediate step. The human reviewer checks:

1. All CI gates green.
2. Approach section matches intent.
3. No Architecture Decision violations in the table.
4. Diff stays within the spec's **Affected Modules** and **Out of Scope** list.
5. No stable root depends on `x/*` (enforced by `dependency-rules` check).
6. `go.mod` unchanged (no new external dependencies).

## 12. Delegation Contract

What humans decide vs. what Codex decides autonomously.

### Human Decides (in the Milestone Spec)

| Concern | Where in spec |
|---------|---------------|
| Goal — what must be true when done | `## Goal` |
| Which modules are in scope | `## Affected Modules` |
| Non-obvious architectural constraints | `## Architecture Decisions` |
| What files to read before coding | `## Context` |
| Task breakdown and phase order | `## Tasks` |
| Acceptance criteria (exact commands) | `## Acceptance Criteria` |
| Hard stops | `## Out of Scope` |
| Milestone ordering and dependencies | `tasks/milestones/ROADMAP.md` |

### Codex Decides Autonomously

| Concern | Constraint |
|---------|------------|
| Internal file structure within declared modules | Must follow `docs/CANONICAL_STYLE_GUIDE.md` |
| Naming of types, vars, funcs (not public API) | Must follow existing patterns in touched files |
| Test file layout and table structure | Must place tests next to changed files |
| Commit granularity within a phase | One commit per stable intermediate state |
| Order within a parallel phase | Any order; phases still run sequentially |
| Formatting and gofmt compliance | Enforced before push |
| Which helper functions to extract | Only if canonical style clearly calls for it |

### Codex Must NOT Decide

- Whether to touch modules outside **Affected Modules**.
- Whether to change stable root public APIs (unless spec explicitly lists it).
- Whether to add entries to `go.mod`.
- Whether to skip or soften a failing quality gate.
- Whether a deviation from an **Architecture Decision** is acceptable (surface it in the PR instead).

### Escalation Protocol

If Codex reaches a genuine decision point not covered by the spec:

1. Document it under `## Open Questions → Codex` in the spec file.
2. Record what decision was made and why.
3. If the decision might violate an Architecture Decision or Out of Scope rule,
   open a **Draft PR** instead of a regular PR and flag it in the PR body.
4. Do not silently deviate and hope the reviewer catches it.
