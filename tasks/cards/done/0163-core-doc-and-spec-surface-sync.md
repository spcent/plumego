# Card 0163

Priority: P2
State: done
Primary Module: core
Owned Files:
- `docs/CANONICAL_STYLE_GUIDE.md`
- `docs/modules/core/README.md`
- `core/module.yaml`
Depends On:
- `0157-core-http-observer-shadow-state-removal.md`
- `0158-core-router-policy-config-normalization.md`
- `0159-core-route-surface-convergence.md`
- `0162-core-runtime-tooling-and-readme-migration.md`

Goal:
- Bring canonical style and module metadata into exact alignment with the
  cleaned-up `core` surface.

Problem:
- `docs/CANONICAL_STYLE_GUIDE.md` still shows `core.New()` with no typed config
  and still includes snippets that ignore route-registration errors.
- `docs/modules/core/README.md` still mentions `AttachHTTPObserver(...)` as a
  kernel capability.
- `core/module.yaml` still advertises observer attachment and still lists
  `health` as an allowed import even though `core` no longer owns either.
- These repo-native source-of-truth files therefore describe an older kernel
  than the codebase actually contains.

Scope:
- Update canonical style examples to the current config-first, explicit-error
  core API.
- Remove outdated observer/health claims from module docs and manifest
  metadata.
- Keep the manifest and style guide synchronized with the post-cleanup kernel.

Non-goals:
- Do not rewrite unrelated architecture guidance.
- Do not broaden the core public surface to match stale docs.
- Do not leave partial "legacy note" language behind.

Files:
- `docs/CANONICAL_STYLE_GUIDE.md`
- `docs/modules/core/README.md`
- `core/module.yaml`

Tests:
- `go run ./internal/checks/module-manifests`
- `go test -timeout 20s ./core/...`
- `go vet ./...`

Docs Sync:
- These files are themselves the sync target; they should become the canonical
  source for the cleaned-up core surface.

Done Definition:
- Canonical style, module docs, and module manifest all describe the same
  current `core` API.
- No stale no-arg `core.New()` or observer-attachment guidance remains.
- Module metadata no longer advertises removed dependencies or responsibilities.

Outcome:
- Updated the canonical style guide to the current config-first bootstrap and
  explicit route-registration contract.
- Removed stale observer-attachment and `health` import claims from
  `core/module.yaml`.
- Brought style guide, module docs, and manifest metadata back into alignment
  with the cleaned-up `core` surface.
