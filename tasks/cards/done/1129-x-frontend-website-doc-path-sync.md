# Card 1129: x/frontend Website Doc Path Sync

Milestone: none
Recipe: specs/change-recipes/docs-sync.yaml
Priority: P0
State: done
Primary Module: x/frontend
Owned Files:
- `website/src/content/docs/docs/modules/x-frontend.mdx`
- `website/src/content/docs/zh/docs/modules/x-frontend.mdx`
- `tasks/cards/active/README.md`
Depends On: 0750

Goal:
Fix stale website module primer references that currently break release gates.

Scope:
- Replace nonexistent `x/frontend/embedded_fs.go` and `x/frontend/embedded/`
  references with current `x/frontend` implementation files.
- Keep English and Chinese docs in sync.
- Verify the website docs API symbol check.

Non-goals:
- Do not change runtime code.
- Do not rewrite the whole website primer.

Files:
- `website/src/content/docs/docs/modules/x-frontend.mdx`
- `website/src/content/docs/zh/docs/modules/x-frontend.mdx`
- `tasks/cards/active/README.md`

Tests:
- `pnpm check:docs-api`
- `env GOCACHE=/private/tmp/plumego-gocache make gates`

Docs Sync:
This is a docs sync card.

Done Definition:
- Website x/frontend primers reference only existing repo paths.
- Website docs API symbol check passes.
- Release gate no longer fails on stale x/frontend website paths.

Outcome:
- Updated English and Chinese website `x/frontend` primers to remove stale
  `x/frontend/embedded_fs.go` and `x/frontend/embedded/` references.
- Replaced the stale embedded-helper guidance with current `RegisterFS` and
  existing implementation file references.
- Installed website dependencies locally to run the full website checks; no
  lockfile change was required.
- Validation passed:
  - `pnpm check:docs-api`
  - `pnpm check`
  - `env GOCACHE=/private/tmp/plumego-gocache make gates`
