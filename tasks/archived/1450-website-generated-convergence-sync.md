# Card 1450: Website Generated Convergence Sync

Milestone: post-v1-convergence
Recipe: specs/change-recipes/analysis-only.yaml
Priority: P1
State: done
Primary Module: website
Owned Files:
- `website/src/generated/modules.ts`
- `website/src/generated/releases.ts`
- `website/src/generated/roadmap.ts`
Depends On: Card 1449

Goal:
- Commit generated website facts after post-v1 extension family convergence.

Scope:
- Accept `pnpm sync` output that reflects the current source docs and specs.
- Keep generated module facts aligned with canonical extension family paths.
- Keep release and roadmap generated facts aligned with current docs.

Non-goals:
- Do not edit source docs by hand in this card.
- Do not change website runtime behavior.
- Do not promote extension maturity.

Files:
- `website/src/generated/modules.ts`
- `website/src/generated/releases.ts`
- `website/src/generated/roadmap.ts`

Tests:
- `cd website && pnpm sync`
- `git diff --check`
- `GOCACHE=/private/tmp/plumego-gocache GOMODCACHE=/private/tmp/plumego-gomodcache make gates`

Docs Sync:
- Not required beyond generated output; source docs already own the prose.

Done Definition:
- Website generated files match `pnpm sync`.
- Repo gate no longer fails on stale website generated files.
- No source docs or runtime code are changed in this card.

Outcome:
- `pnpm sync` removed old subordinate extension roots from generated module
  facts and updated generated release/roadmap facts to the converged family
  paths.
