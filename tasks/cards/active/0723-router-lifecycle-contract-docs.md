# Card 0723

Milestone: Router stable readiness
Recipe: specs/change-recipes/stable-root-boundary-review.yaml
Priority: P2
State: active
Primary Module: router
Owned Files: docs/modules/router/README.md, router/module.yaml, tasks/cards/active/README.md
Depends On: 0722-router-internal-cleanup

Goal:
Document the stable lifecycle contract for direct Router use versus core-owned Router use.

Scope:
- Document build-time registration and Freeze-before-serving expectations for direct Router use.
- Clarify that core Prepare/ServeHTTP freezes its owned router.
- Keep active task queue accurate when this batch is done.

Non-goals:
- Code behavior changes.
- Core lifecycle changes.
- New public APIs.

Files:
- docs/modules/router/README.md
- router/module.yaml
- tasks/cards/active/README.md

Tests:
- go run ./internal/checks/module-manifests
- go run ./internal/checks/agent-workflow
- go vet ./router/...

Docs Sync:
- Required.

Done Definition:
- Router docs describe stable lifecycle expectations.
- Manifest/workflow checks and router vet pass.

Outcome:

