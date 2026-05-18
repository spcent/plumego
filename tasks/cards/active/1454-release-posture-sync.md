# Card 1454

Milestone: M-003
Recipe: specs/change-recipes/analysis-only.yaml
Priority: P1
State: active
Primary Module: docs
Owned Files:
- `README.md`
- `README_CN.md`
- `docs/ROADMAP.md`
- `docs/release/v1.0.0.md`
- `website/src/generated`
Depends On: tasks/cards/active/1453-cli-gate-timeout-stability.md

Goal:
- Align user-facing release posture with the recorded `v1.0.0` release evidence.

Scope:
- Replace stale `pre-v1` user-facing status with wording that matches the tagged `v1.0.0` release.
- Keep the extension support matrix split between stable roots, beta extensions, and experimental extensions.
- Regenerate website data if docs sources change generated content.

Non-goals:
- Do not change Go APIs.
- Do not promote experimental extensions.
- Do not rewrite the adoption path.

Files:
- `README.md`
- `README_CN.md`
- `docs/ROADMAP.md`
- `docs/release/v1.0.0.md`
- `website/src/generated/*`

Tests:
- `bash scripts/check-doc-snippets-compile.sh`
- `cd website && pnpm check`
- `cd website && pnpm build`

Docs Sync:
- `README.md` and `README_CN.md` must stay aligned in scope and meaning.

Done Definition:
- Repository entrypoints no longer claim `pre-v1` while release docs and tags claim `v1.0.0`.

Outcome:
-
