# Card 1451: Website Doc Path Convergence Sync

Milestone: post-v1-convergence
Recipe: specs/change-recipes/analysis-only.yaml
Priority: P1
State: done
Primary Module: website
Owned Files:
- `website/src/content/docs/**`
Depends On: Card 1450

Goal:
- Align website documentation repo-path references with post-v1 extension
  family convergence.

Scope:
- Replace old subordinate extension paths in website content with their new
  family paths.
- Keep English and Chinese website docs aligned mechanically.
- Preserve page structure and avoid runtime website behavior changes.

Non-goals:
- Do not change Go code.
- Do not promote extension maturity.
- Do not rewrite root docs in this card.

Files:
- `website/src/content/docs/**`

Tests:
- `cd website && pnpm check`
- `git diff --check`
- `GOCACHE=/private/tmp/plumego-gocache GOMODCACHE=/private/tmp/plumego-gomodcache make gates`

Docs Sync:
- This card is website docs sync for already-completed path convergence.

Done Definition:
- `pnpm check` no longer reports missing repo paths for old subordinate roots.
- Website content references converged paths such as `x/data/cache`,
  `x/gateway/discovery`, `x/messaging/pubsub`, and `x/observability/ops`.
- Repo gates pass.

Outcome:
- Mechanically replaced old website documentation repo-path references with
  converged extension family paths:
  - `x/cache` -> `x/data/cache`
  - `x/devtools` -> `x/observability/devtools`
  - `x/ops` -> `x/observability/ops`
  - `x/discovery` -> `x/gateway/discovery`
  - `x/ipc` -> `x/gateway/ipc`
  - `x/mq` -> `x/messaging/mq`
  - `x/pubsub` -> `x/messaging/pubsub`
  - `x/scheduler` -> `x/messaging/scheduler`
  - `x/webhook` -> `x/messaging/webhook`
