# Card 0790

Priority: P2
State: done
Primary Module: core
Owned Files:
- `README.md`
- `README_CN.md`
- `docs/getting-started.md`
- `docs/CANONICAL_STYLE_GUIDE.md`
- `docs/modules/core/README.md`
- `core/module.yaml`
- `AGENTS.md`
- `CLAUDE.md`
Depends On:
- `0786-core-internal-preparation-state-unification.md`
- `0787-core-router-policy-sync-side-effect-removal.md`
- `0788-core-constructor-dependency-surface-normalization.md`
- `0789-core-named-route-surface-pruning.md`

Goal:
- Bring repo-native docs and manifest metadata into exact alignment with the
  next round of cleaned-up `core` APIs.

Problem:
- The remaining cleanup items will change internal state semantics, constructor
  shape, and route helper surface.
- Those changes will invalidate current examples and manifest metadata unless
  they are synchronized together after implementation.
- Leaving docs half-updated would recreate the same ambiguity the cleanup is
  trying to remove.

Scope:
- Sync top-level READMEs, getting-started, canonical style guide, module docs,
  and manifest metadata to the final cleaned-up `core` surface.
- Remove stale references to deleted constructor or route helper contracts.
- Keep repo-native source-of-truth files aligned with actual code.

Non-goals:
- Do not preserve legacy notes for removed APIs.
- Do not broaden `core` to match stale docs.
- Do not rewrite unrelated architecture guidance.

Files:
- `README.md`
- `README_CN.md`
- `docs/getting-started.md`
- `docs/CANONICAL_STYLE_GUIDE.md`
- `docs/modules/core/README.md`
- `core/module.yaml`
- `AGENTS.md`
- `CLAUDE.md`

Tests:
- `go run ./internal/checks/module-manifests`
- `go test -timeout 20s ./core/...`
- `go vet ./...`

Docs Sync:
- This card is the sync pass; these files are the target.

Done Definition:
- Docs, examples, and manifest metadata describe the same final `core` API.
- No stale constructor, route-helper, or state-language remains.
- Repo-native source-of-truth files are aligned with the implemented cleanup.
