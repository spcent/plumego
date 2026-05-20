# Agent-First Design

Plumego treats automated coding agents as regular maintainers, not as a
separate workflow bolted onto the project. The repository is arranged so an
agent can determine ownership, scope, validation, and stop conditions from
checked-in files before it edits code. That does not replace human review. It
narrows the problem: humans define the direction and review the result, while
agents follow explicit boundaries and produce evidence that can be inspected.

The model has four control planes.

The first is `docs/`. Documentation describes observable behavior, architecture
boundaries, release policy, migration paths, and module primers. Agent-facing
docs such as `docs/CODEX_WORKFLOW.md`, `docs/AGENT_CONTEXT_BUDGET.md`,
`docs/AGENT_CODE_QUALITY_RULES.md`, and `docs/CANONICAL_STYLE_GUIDE.md` keep
implementation style consistent while keeping default context small. Product
docs stay close to the code they explain, so doc drift is visible in review
instead of hidden in a separate planning system.

The second is `specs/`. These machine-readable files make repository policy
queryable. `specs/task-routing.yaml` maps request shapes to the owning module,
context package, and bounded first reads. `specs/dependency-rules.yaml` prevents
stable roots from importing extension packages or gaining unapproved
dependencies. `specs/checks.yaml` and `specs/gate-profiles.yaml` describe
validation expectations. `specs/stop-condition-handlers.yaml` turns uncertainty
into deterministic next steps instead of letting an agent guess. The goal is not
to make every rule executable, but to keep important rules structured enough for
tools and humans to compare against behavior.

The third is `tasks/`. Milestones and task cards are the execution surface.
A milestone under `tasks/milestones/active/` defines one reviewable unit:
goal, context, affected modules, ordered tasks, acceptance criteria, and hard
out-of-scope boundaries. Cards under `tasks/cards/` split work into smaller
pieces with owned files and focused tests. Completed milestones keep their
verify bundle in `tasks/milestones/done/`, so future agents can see what was
actually run rather than inferring from commit messages.

The fourth is `reference/`. Reference applications show how Plumego should be
wired by callers. `reference/standard-service` is the canonical application
layout. Scenario references such as `reference/with-rest`, `reference/with-rpc`,
and `reference/with-tenant-admin` demonstrate optional extension families
without moving product behavior into stable roots. This protects the public API:
core packages stay small, while richer examples remain copyable and explicit.

Enforcement lives in `internal/checks/`. The checks are intentionally small Go
programs that can run locally and in CI. `dependency-rules` enforces import and
dependency boundaries. `agent-workflow` validates milestone and card control
plane shape. `module-manifests` checks every `module.yaml`. `reference-layout`
keeps reference applications aligned with the canonical structure. `public-entrypoints-sync`
guards the stable public entrypoint inventory. These tools make review cheaper:
an agent can run the same checks before handoff that CI will run later.

Change recipes in `specs/change-recipes/` make common work reproducible. A bug
fix, new HTTP endpoint, middleware change, symbol change, or review-only task
has a standard checklist for scope, tests, and documentation sync. The recipes
are deliberately conservative. They push agents toward the smallest reversible
change, visible constructor wiring, ordinary `net/http` shapes, and focused
validation.

To adopt this model in another Go project, start with module ownership. Create
per-package manifests that state owner, status, allowed imports, and required
tests. Add a routing spec that maps request types to packages and docs. Add a
small set of checks that fail on the most expensive mistakes, such as forbidden
dependencies, missing manifests, stale public API inventories, or malformed
task specs. Then introduce task cards only for work that benefits from explicit
handoff. Do not start with a large process document. Start with two or three
rules that reduce review risk, make them executable, and expand only when the
repository shows repeated failure patterns.

The important constraint is that agent-first does not mean agent-only. The
repository should remain understandable to a Go developer using normal tools:
`go test`, `go vet`, `gofmt`, `make`, and GitHub review. The control planes make
agent work legible, but they also make human maintenance easier because the
reason for a boundary, validation command, or stop condition is stored next to
the code it protects.
