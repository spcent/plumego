# Card 0818

Priority: P1
State: done
Primary Module: reference/standard-service
Owned Files:
- `README.md`
- `README_CN.md`
- `docs/README.md`
- `docs/getting-started.md`
- `reference/standard-service/README.md`
Depends On:

Goal:
- Keep the top-level onboarding docs aligned with the current canonical reference app and repository workflow surface.

Scope:
- Sync the top-level README files, docs index, getting-started guide, and reference-service primer so they describe the same canonical next reads and reference-app role.
- Remove stale onboarding wording that no longer matches the current `reference/standard-service` layout or current docs/specs/tasks split.
- Keep the first-run guidance grounded in implemented behavior and live repository surfaces only.

Non-goals:
- Do not change reference app runtime behavior in this card.
- Do not document planned extensions as part of the default onboarding path.
- Do not broaden this card into module-specific extension primers.

Files:
- `README.md`
- `README_CN.md`
- `docs/README.md`
- `docs/getting-started.md`
- `reference/standard-service/README.md`

Tests:
- `go run ./internal/checks/reference-layout`
- `go test -timeout 20s ./reference/standard-service/...`
- `go vet ./reference/standard-service/...`

Docs Sync:
- Keep onboarding docs aligned with `reference/standard-service`, the current docs/specs/tasks split, and the repository's canonical read order.

Done Definition:
- The top-level onboarding surface points readers at the same canonical reference app and next reads.
- No stale wording suggests a different bootstrap shape or outdated workflow surface.
- The reference-layout and reference-package validation commands stay green.

Outcome:
- Updated `README.md`, `README_CN.md`, `docs/README.md`, `docs/getting-started.md`, and `reference/standard-service/README.md` so they describe the same canonical onboarding order: getting-started, reference app, docs index, then specs/tasks as deeper control surfaces.
- Removed stale ambiguity around whether the docs index or specs/tasks come before the canonical reference app in the default learning path.
- Validation:
  - `go run ./internal/checks/reference-layout`
  - `go test -timeout 20s ./reference/standard-service/...`
  - `go vet ./reference/standard-service/...`
