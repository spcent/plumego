# Plan for M-021: Agent-First Ecosystem Visibility

Milestone: `M-021`
Branch: `milestone/M-021-agent-first-visibility`
Affected Modules: docs, GitHub workflows, README.

## Phase Map

1. Orient: read the agent workflow specs, existing checks, and workflow files.
2. Implement: add `docs/AGENT_FIRST.md`, update `quality-gates.yml`, and add README links.
3. Validate: parse workflow YAML, check guide length and links, and run manifest/format checks.
4. Ship: record verification evidence.

## Card Inventory

- External agent-first architecture guide.
- Consolidated `quality-gates.yml` updates for gates, benchmarks, and spec validation.
- README and README_CN visibility updates.

## Dependency Edges

- Workflow commands depend on the actual benchmark directory from M-011, currently `benchmark/`.
- README links depend on `docs/AGENT_FIRST.md` existing.

## Parallel Groups

- Guide drafting.
- Workflow consolidation.
- README updates.

## Risk Register

- Workflow references stale milestone paths. Mitigation: use existing `benchmark/` and current `internal/checks/` commands.
- YAML syntax drift. Mitigation: parse `quality-gates.yml` with Python YAML.
- Overlong or under-specified guide. Mitigation: keep `docs/AGENT_FIRST.md` between 600 and 900 words and link concrete repo files.

## Verification Strategy

- `go run ./internal/checks/module-manifests`
- `gofmt -l .`
- Python YAML parse for `quality-gates.yml`.
- Word count and README link checks.

## Exit Condition

M-021 is complete when the guide, consolidated workflow, and README links exist and the acceptance checks pass.
