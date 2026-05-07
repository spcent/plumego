# Card 0729

Milestone:
Recipe: specs/change-recipes/analysis-only.yaml
Priority: P2
State: done
Primary Module: x/frontend
Owned Files: x/frontend/module.yaml, docs/extension-evidence/x-frontend.md
Depends On: 0728

Goal:
Record the stable runtime contract decisions that must be owner-accepted before promotion.

Scope:
- Update the x/frontend evidence ledger with the current decisions for:
  - precompressed downgrade observability
  - custom FS lazy probing and explicit mitigation
  - directory-backed static artifact immutability
  - AddRoute-only registrar partial registration
  - nil receiver inspection methods
- Keep unresolved owner approval marked as missing.
- Do not claim stable readiness until owner sign-off exists.

Non-goals:
- Do not change runtime behavior.
- Do not promote module status.
- Do not fabricate owner sign-off.

Files:
- `x/frontend/module.yaml`
- `docs/extension-evidence/x-frontend.md`

Tests:
- `go run ./internal/checks/extension-beta-evidence`

Docs Sync:
- Evidence ledger only; do not duplicate prose in user docs unless behavior changed in earlier cards.

Done Definition:
- Evidence clearly distinguishes implemented runtime decisions from missing owner approval.
- Beta evidence check still reports release/owner blockers honestly.

Outcome:
- Recorded implemented runtime contract decisions in the x/frontend evidence ledger.
- Updated module risk wording to track implemented hook/plan regressions rather than unresolved decision points.
- Kept owner sign-off and release evidence blockers unresolved.
- Validation Run:
  - `go run ./internal/checks/extension-beta-evidence`
