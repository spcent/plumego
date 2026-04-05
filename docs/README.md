# Docs Index

This directory is the human-readable documentation surface for Plumego.

Use the documents below by purpose instead of reading everything at once.

## Start Here

- `getting-started.md` — smallest runnable example plus the canonical next reads
- `CANONICAL_STYLE_GUIDE.md` — one obvious way to structure handlers, routes, middleware, and responses
- `architecture/AGENT_FIRST_REPO_BLUEPRINT.md` — repository shape, boundaries, and canonical implementation path
- `ROADMAP.md` — active staged roadmap for repository and feature evolution

## Agent Workflow

- `CODEX_WORKFLOW.md` — milestone-driven execution workflow
- `MILESTONE_PIPELINE.md` — `Milestone -> Plan -> Cards -> Verify -> PR` artifact contract
- `DEPRECATION.md` — deprecation and removal policy for stable roots and extension evolution

## Module Primers

- `modules/*/README.md` — package-family primers that should match each module manifest's `doc_paths`

## GitHub Workflow Assets

- `github-workflows/milestone-pr-template.md` — reviewer-facing PR template for milestone work
- `github-workflows/milestone-gates.yml` — CI workflow example for milestone gates

## Archived Drafts

These paths are kept only as historical pointers. They are not current authority.

- `PRODUCT_PRD_AGENT_NATIVE.md` — archived draft; superseded by `ROADMAP.md`, `CODEX_WORKFLOW.md`, and `specs/*`
- `ROADMAP_AGENT_NATIVE.md` — archived draft roadmap; superseded by `ROADMAP.md` and milestone workflow docs

## Notes

- `docs/` explains the architecture and workflow in prose.
- `specs/` remains the machine-readable control plane.
- `tasks/` remains the execution surface for active work.
- When prose and machine-readable rules disagree, prefer `AGENTS.md`, `specs/*`, and the touched module manifest.
