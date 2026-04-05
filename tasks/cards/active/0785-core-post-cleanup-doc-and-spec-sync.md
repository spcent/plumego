# Card 0785

Priority: P2
State: active
Primary Module: core
Owned Files:
- `docs/modules/core/README.md`
- `docs/CANONICAL_STYLE_GUIDE.md`
- `README.md`
- `README_CN.md`
- `core/module.yaml`
- `AGENTS.md`
- `CLAUDE.md`
Depends On:
- `0781-core-runtime-state-contract-collapse.md`
- `0782-core-server-settings-projection-deduplication.md`
- `0783-core-logger-lifecycle-ownership-removal.md`
- `0784-core-introspection-surface-pruning.md`

Goal:
- Bring repo-native docs and spec metadata into exact alignment with the final
  post-cleanup `core` surface.

Problem:
- The next cleanup steps change exported runtime state semantics, logger
  lifecycle ownership, and kernel introspection boundaries.
- These changes will invalidate current core module docs, canonical examples,
  and module metadata unless they are updated together at the end.
- Leaving docs half-synced would recreate the same confusion the cleanup is
  trying to remove.

Scope:
- Sync core module docs, canonical examples, top-level READMEs, and module
  metadata to the final cleaned-up kernel contract.
- Remove stale wording about logger lifecycle ownership or middleware-list
  introspection if those surfaces are gone.
- Keep the repo's source-of-truth docs aligned with the actual code.

Non-goals:
- Do not preserve legacy notes for removed contracts.
- Do not rewrite unrelated architecture guidance.
- Do not broaden the kernel API to match stale documentation.

Files:
- `docs/modules/core/README.md`
- `docs/CANONICAL_STYLE_GUIDE.md`
- `README.md`
- `README_CN.md`
- `core/module.yaml`
- `AGENTS.md`
- `CLAUDE.md`

Tests:
- `go run ./internal/checks/module-manifests`
- `go test -timeout 20s ./core/...`
- `go vet ./...`

Docs Sync:
- This card is the sync pass; these files are the primary target.

Done Definition:
- Core docs, examples, and module metadata describe the same final kernel
  surface.
- No stale runtime-state, logger-lifecycle, or middleware-list guidance
  remains.
- Repo-native source-of-truth files are aligned with the implemented cleanup.
