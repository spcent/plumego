# Card 0751

Milestone:
Recipe: specs/change-recipes/review-only.yaml
Priority: P2
State: done
Primary Module: x/data
Owned Files:
- docs/modules/x-data/README.md
- tasks/cards/active/0751-x-data-stable-readiness-second-gate.md
Depends On:
- 0750-x-data-sharding-metrics-and-watcher-lifecycle

Goal:
Run the second x/data stable-readiness gate and record remaining blockers after cards 0744-0750.

Scope:
- Run x/data tests, race tests, vet, and boundary/manifest checks.
- Update docs with remaining blockers.
- Keep x/data experimental unless all API and operational blockers are resolved.

Non-goals:
- Do not widen scope to unrelated x modules.
- Do not run make gates unless needed by the final conclusion.
- Do not mark stable without explicit evidence.

Files:
- docs/modules/x-data/README.md
- tasks/cards/active/0751-x-data-stable-readiness-second-gate.md

Tests:
- go test -timeout 20s ./x/data/...
- go test -race -timeout 60s ./x/data/...
- go vet ./x/data/...

Docs Sync:
- Update x/data docs with the second gate result and remaining blockers.

Done Definition:
- Validation commands are recorded.
- Remaining NO-GO items are explicit, or experimental status is justified.
- Completed card is archived to done.

Outcome:
- Ran the second x/data stable-readiness gate after cards 0744-0750.
- Recorded the passing validation set in the x/data docs.
- Kept x/data experimental because API freeze decisions and large-object S3
  operational policy remain open.

Validation:
- `go test -timeout 20s ./x/data/...`
- `go test -race -timeout 60s ./x/data/...`
- `go vet ./x/data/...`
- `go run ./internal/checks/dependency-rules`
- `go run ./internal/checks/agent-workflow`
- `go run ./internal/checks/module-manifests`
- `go run ./internal/checks/reference-layout`
