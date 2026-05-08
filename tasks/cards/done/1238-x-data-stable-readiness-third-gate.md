# Card 1238

Milestone:
Recipe: specs/change-recipes/review-only.yaml
Priority: P2
State: done
Primary Module: x/data
Owned Files:
- docs/modules/x-data/README.md
- tasks/cards/done/1238-x-data-stable-readiness-third-gate.md
Depends On:
- 0760-x-data-observability-redaction-boundary

Goal:
Run the third x/data stable-readiness gate and record remaining blockers after cards 0752-0760.

Scope:
- Run x/data tests, race tests, vet, and boundary/manifest checks.
- Update docs with remaining stable blockers.
- Keep x/data experimental unless all API, durability, and operational blockers are resolved.

Non-goals:
- Do not mark stable without explicit go/no-go evidence.
- Do not widen scope to unrelated active websocket cards.
- Do not run make gates unless needed by the final conclusion.

Files:
- docs/modules/x-data/README.md
- tasks/cards/done/1238-x-data-stable-readiness-third-gate.md

Tests:
- go test -timeout 20s ./x/data/...
- go test -race -timeout 60s ./x/data/...
- go vet ./x/data/...

Docs Sync:
- Update x/data docs with the third gate result and remaining blockers.

Done Definition:
- Validation commands are recorded.
- Remaining NO-GO items are explicit, or experimental status is justified.
- Completed card is archived to done.

Outcome:
- Ran the third x/data stable-readiness gate after cards 0752-0760.
- Recorded the passing validation set in the x/data docs.
- Kept x/data experimental because API surface freeze decisions, kvengine
  auto-detect option cleanup, large-object S3 policy, and repo-wide gates remain
  open.

Validation:
- `go test -timeout 20s ./x/data/...`
- `go test -race -timeout 60s ./x/data/...`
- `go vet ./x/data/...`
- `go run ./internal/checks/dependency-rules`
- `go run ./internal/checks/agent-workflow`
- `go run ./internal/checks/module-manifests`
- `go run ./internal/checks/reference-layout`
