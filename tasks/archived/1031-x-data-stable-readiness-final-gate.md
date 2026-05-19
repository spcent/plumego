# Card 1031

Milestone:
Recipe: specs/change-recipes/review-only.yaml
Priority: P2
State: done
Primary Module: x/data
Owned Files:
- x/data/module.yaml
- docs/modules/x-data/README.md
- tasks/cards/done/1031-x-data-stable-readiness-final-gate.md
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
- tasks/cards/done/1031-x-data-stable-readiness-final-gate.md

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
- Ran final x/data validation after cards 0737-0742.
- Left x/data/module.yaml status as experimental.
- Updated docs/modules/x-data/README.md with 2026-05-04 stable-readiness
  blockers covering API freeze, SQL support policy, kvengine option shape,
  large-object S3 guarantees, and repo-wide gate requirements.

Validation:
- GOCACHE=/private/tmp/plumego-go-build go test -timeout 20s ./x/data/...
- GOCACHE=/private/tmp/plumego-go-build go test -race -timeout 60s ./x/data/...
- GOCACHE=/private/tmp/plumego-go-build go vet ./x/data/...
- GOCACHE=/private/tmp/plumego-go-build go run ./internal/checks/dependency-rules
- GOCACHE=/private/tmp/plumego-go-build go run ./internal/checks/agent-workflow
- GOCACHE=/private/tmp/plumego-go-build go run ./internal/checks/module-manifests
- GOCACHE=/private/tmp/plumego-go-build go run ./internal/checks/reference-layout
- gofmt -l x/data docs/modules/x-data
