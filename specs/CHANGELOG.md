# specs/ Changelog

Track rule changes here so agents can detect when a previously-learned rule
has changed. Update this file whenever a rule in specs/ changes in a way that
would affect agent behavior.

Format: `## YYYY-MM-DD — <brief description>` followed by bullet points.

---

## 2026-05-16 — Initial changelog; task_affinity and landing_zones added to module.yaml

- All `module.yaml` files now declare `task_affinity` (maps task type → landing
  sub-path) and `landing_zones` (maps code role → sub-directory).
- `middleware/module.yaml` now declares `selection_guide` mapping use-case keywords
  to the correct sub-package.
- `specs/gate-profiles.yaml` added: machine-readable gate profile definitions used
  by `plumego agents validate-diff` and `make validate-diff`.
- `make bundle` now writes `.agent-bundle.yaml` with complete task execution context.
- `plumego generate endpoint|service|repo` added for canonical code scaffolding.

---

## How to update this file

1. After merging a spec change, add an entry at the TOP (newest first).
2. Reference the changed spec file(s) explicitly.
3. Describe what an agent relying on the old rule needs to change.
