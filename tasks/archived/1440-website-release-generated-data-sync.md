# Card 1440

Milestone: M-006
Recipe: specs/change-recipes/docs-only.yaml
Priority: P0
State: done
Primary Module: website
Owned Files:
- `website/src/generated/releases.ts`
Depends On:
- 1439

Goal:
- Resolve the current generated release data drift so Codex starts v1.0.1 work
  from a clean workspace.

Scope:
- Confirm whether `website/src/generated/releases.ts` matches `pnpm sync`.
- If the drift is generated truth, commit it with evidence.
- If the drift is accidental, restore it and record the reason.

Non-goals:
- Do not change release source docs.
- Do not change website runtime behavior outside generated release data.
- Do not rewrite release tags.

Files:
- `website/src/generated/releases.ts`

Tests:
- `cd website && pnpm sync`
- `git diff --check`
- `git status --short --branch`

Docs Sync:
- Not required beyond generated data classification.

Done Definition:
- Generated release data is either committed or restored.
- Workspace status is understood before the next maintenance card starts.

Outcome:
- `cd website && pnpm sync` passed and left the same
  `website/src/generated/releases.ts` status normalization diff.
- The generated drift is committed as source-of-truth output from the website
  sync pipeline.
- No release source docs or runtime code were changed.
