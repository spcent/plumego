# Agent-First Design

Plumego is designed for codebases where AI coding agents participate in
maintenance alongside human engineers. This document explains the rationale,
describes the mechanisms that make agent-safe maintenance possible, and provides
a reference for how agents should operate within the repository.

For the agent operating rules, see `AGENTS.md`.
For repeatable execution patterns, see `docs/CODEX_WORKFLOW.md`.
For quality preflight rules, see `docs/AGENT_CODE_QUALITY_RULES.md`.

---

## Why Most Frameworks Are Not Agent-Ready

When an AI coding agent is given a maintenance task in a typical Go web service,
it faces several structural problems:

- **Hidden registration.** Routes and middleware registered through `init()` or
  package-level globals are not locatable by reading the entry point. The agent
  cannot know what is live without tracing the full import graph.
- **No scope signals.** There is no machine-readable answer to "which paths are
  safe to modify for this task?" The agent either modifies too little or risks
  breaking unrelated behavior.
- **No standardized validation.** Knowing whether a change is correct requires
  running the right subset of tests. Without a per-module manifest, the agent
  either runs everything (slow, noisy) or guesses (risky).
- **Docs drift from behavior.** When documentation and implementation diverge,
  an agent trained on the docs produces code that disagrees with the running
  system.
- **Broad tasks without acceptance criteria.** Without scoped task cards, agents
  frequently widen their changes beyond the intended boundary.

Plumego addresses each of these problems with explicit, machine-readable
repository structures.

---

## Agent-First Mechanisms

### 1. Per-module `module.yaml`

Every package in the repository that carries public behavior has a `module.yaml`
declaring:

- `name`: canonical module path
- `status`: `experimental`, `beta`, or `ga`
- `owner`: responsible team or individual
- `risk`: `low`, `medium`, or `high`
- `responsibilities`: what this module owns
- `non_goals`: what this module explicitly does not own
- `test_commands`: the exact commands to validate a change to this module
- `review_checklist`: what a reviewer must check
- `agent_hints`: additional guidance for AI agents working in this module

An agent reading `module.yaml` before editing knows its scope before it touches
a single file. Schema enforcement runs via:

```bash
go run ./internal/checks/module-manifests
```

---

### 2. `specs/task-routing.yaml` — task-to-module routing

This file maps task intent to the correct starting point in the repository.
Before writing any code, an agent should identify the task type and follow the
routing:

| Intent | Route | Avoid |
|---|---|---|
| Kernel, lifecycle, routing, transport contracts | `core`, `router`, `contract`, `middleware` | `x/*` |
| Business feature, protocol adaptation, extension behavior | Primary `x/*` family | Stable roots |
| Application wiring, bootstrap, route registration | `reference/standard-service` | Extension modules |
| Architecture rules, quality gates | `specs/` | Reference apps |
| Execution plans, task sequencing | `tasks/cards/` | `specs/` |

Concrete task types (HTTP endpoint, middleware change, tenant policy, symbol
change, bugfix) have explicit routing entries that also specify which files to
read first and which paths to avoid.

---

### 3. `specs/checks.yaml` — standardized validation

Every change type has a defined set of validation commands. Agents do not need
to infer which tests to run. The module manifest specifies `test_commands`; the
spec file specifies how those compose into broader profiles.

Default validation order:

1. Run `test_commands` from the target module's `module.yaml`.
2. Run boundary and manifest checks.
3. Run repo-wide gates when the change is code-bearing, cross-module, or
   release-relevant.

Boundary checks:

```bash
go run ./internal/checks/dependency-rules
go run ./internal/checks/cross-extension-deps
go run ./internal/checks/agent-workflow
go run ./internal/checks/module-manifests
go run ./internal/checks/reference-layout
go run ./internal/checks/public-entrypoints-sync
```

Repo-wide gate:

```bash
make gates
```

---

### 4. `specs/package-hotspots.yaml` — ambiguous boundary flags

Some packages sit at junctions where the wrong choice causes significant rework.
The hotspot index flags these packages and provides explicit start-with and
avoid guidance that supplements the module manifest.

An agent working near a flagged package reads the hotspot entry before deciding
where to make changes.

---

### 5. `specs/change-recipes/*` — per-task-type guardrails

Each change recipe is a machine-readable YAML file that specifies:

- `goal`: what this recipe accomplishes
- `preflight`: what to verify before editing
- `steps`: ordered implementation steps
- `validation`: which checks must pass
- `non_goals`: what this recipe must not do

Available recipes:

| Recipe | Use when |
|---|---|
| `add-http-endpoint.yaml` | Adding a new route and handler |
| `add-middleware.yaml` | Adding transport-level middleware |
| `fix-bug.yaml` | Localizing and fixing an implementation bug |
| `http-endpoint-bugfix.yaml` | Fixing a defect in route wiring or handler transport logic |
| `add-acceptance-tests.yaml` | Writing task-card acceptance tests before implementation |
| `add-websocket-room.yaml` | Adding a bounded websocket room capability |
| `add-ai-tool.yaml` | Adding an AI tool surface under the existing AI family |
| `add-grpc-method.yaml` | Adding a gRPC method through the RPC transport family |
| `symbol-change.yaml` | Renaming, removing, or changing an exported symbol |
| `new-extension-module.yaml` | Creating a new `x/*` capability family |
| `new-stable-module.yaml` | Adding a new stable root package |
| `tenant-policy-change.yaml` | Changing tenant resolution or policy |
| `stable-root-boundary-review.yaml` | Reviewing stable-root changes without coding |
| `analysis-only.yaml` | Research and planning tasks with no file edits |
| `review-only.yaml` | Review tasks that produce findings, not patches |

---

### 6. `tasks/cards/` — scoped, verifiable work units

Task cards in `tasks/cards/active/` define bounded work units that agents can
execute without widening scope. Cards waiting on external prerequisites belong
in `tasks/cards/blocked/`, so agents do not mistake evidence waits for
executable work. Each card specifies:

- `Goal`: the single objective
- `Scope`: which files and modules are in scope
- `Non-goals`: what must not change
- `Tests`: exact validation commands
- `Done Definition`: machine-checkable completion criteria

An agent that executes within the card's scope, runs the specified tests, and
meets the done definition can push a reviewable PR without human intervention to
clarify the task.

---

### 7. Reference applications — canonical code output targets

When an agent generates or modifies application-level code, it has a canonical
reference: `reference/standard-service`. The reference app demonstrates the
correct output for:

- `main.go` shape
- Bootstrap sequence
- Middleware attachment order
- Route registration style
- Handler and response conventions
- Graceful shutdown wiring

Additional reference apps (`reference/with-rest`, `reference/with-gateway`,
etc.) demonstrate correct extension mounting patterns. An agent that deviates
from the reference app shape without explicit justification is producing
non-canonical output.

---

### 8. Docs-behavior sync checks

Automated checks verify that documentation and implementation do not drift:

```bash
go run ./internal/checks/agent-workflow
go run ./internal/checks/module-manifests
go run ./internal/checks/extension-maturity
```

These checks fail when manifest fields, dashboard entries, and spec references
fall out of sync with actual module state. Agents updating behavior must update
the corresponding docs or the checks will fail at CI time.

---

## Safe, Restricted, and Frozen Zones

### Safe to modify with standard change protocol

- `docs/*` — documentation (docs-only changes skip repo-wide gates)
- `tests/*` — test additions that do not change public API
- `reference/*` — reference application wiring
- `x/*` experimental modules — changes stay within the module boundary
- `tasks/cards/` — task cards and execution plans

### Restricted — requires preflight checklist and boundary review

- `core`, `router`, `contract`, `middleware`, `security`, `store`, `health`,
  `log`, `metrics` — stable root changes require explicit compatibility
  assessment; use `specs/change-recipes/stable-root-boundary-review.yaml`
- `specs/` — control plane changes affect all agents and checks; require
  analysis-only mode first
- Any exported symbol change — follow `specs/change-recipes/symbol-change.yaml`
  in full, enumerating all call sites before editing

### Frozen — do not modify without explicit milestone authorization

- Stable public API signatures after v1 tagging
- `docs/stable-api/snapshots/` — generated outputs, not hand-edited
- `specs/extension-beta-evidence.yaml` — promotion records, modified only in
  promotion cards

---

## Agent Standard Workflow

```
1. Identify task type → read the matching specs/task-routing.yaml entry
2. Select context package → docs/AGENT_CONTEXT_BUDGET.md if unclear
3. Identify owning module → read <module>/module.yaml
4. Check zone → safe / restricted / frozen
5. Read change recipe → specs/change-recipes/<type>.yaml
6. Run preflight checklist from AGENTS.md §5
7. Make minimal change within scope
8. Run module test_commands from module.yaml
9. Run boundary checks selected by the gate profile
10. If cross-module or release-relevant: run make gates
11. Update docs if behavior, API, config, security, lifecycle, or boundary changed
12. Verify against Done Definition in task card
```

Deviating from this sequence — especially skipping preflight or widening scope
without justification — produces changes that are harder to review and more
likely to break hidden dependencies.

---

## Why This Matters as AI Maintenance Scales

As AI agents take on larger shares of routine maintenance work, the structural
properties of a codebase become a safety property:

- A codebase with hidden globals and self-registering packages is unsafe for
  agent modification without full-repo analysis on every task.
- A codebase with machine-readable module manifests, explicit task routing, and
  standardized check commands is safe for agents to modify incrementally, with
  each task staying within a verifiable scope.

Plumego's design is not optimized for agent convenience. It is optimized for
correctness and reviewability — properties that happen to benefit both human
engineers and AI agents for the same structural reasons.
