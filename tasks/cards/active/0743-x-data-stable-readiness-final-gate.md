# Card 0743

Milestone:
Recipe: specs/change-recipes/review-only.yaml
Priority: P2
State: active
Primary Module: x/data
Owned Files:
- x/data/module.yaml
- docs/modules/x-data/README.md
- tasks/cards/active/0743-x-data-stable-readiness-final-gate.md
Depends On:
- 0742-x-data-sharding-api-surface-and-sql-support-freeze

Goal:
Run the final x/data stable-readiness review and record remaining blockers without prematurely marking the module stable.

Scope:
- Re-run x/data tests, race tests, vet, boundary checks, and module manifest checks.
- Update docs with remaining stable blockers or readiness status.
- Leave module status experimental unless every blocker has been resolved.

Non-goals:
- Do not widen scope into unrelated x modules.
- Do not mark x/data stable unless all previous cards and gates justify it.
- Do not edit stable roots.

Files:
- x/data/module.yaml
- docs/modules/x-data/README.md
- tasks/cards/active/0743-x-data-stable-readiness-final-gate.md

Tests:
- go test -timeout 20s ./x/data/...
- go test -race -timeout 60s ./x/data/...
- go vet ./x/data/...

Docs Sync:
- Update x/data docs with implemented readiness and remaining blockers.

Done Definition:
- Final validation commands are recorded.
- Remaining NO-GO items are explicit, or module status is updated only with evidence.
- Completed cards are archived to done.

Outcome:
