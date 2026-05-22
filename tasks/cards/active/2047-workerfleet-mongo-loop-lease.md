# Card 2047

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Context Package: implementation
Priority: P0
State: active
Primary Module: reference/workerfleet
Owned Files:
- reference/workerfleet/internal/app
- reference/workerfleet/internal/platform/store/mongo
- reference/workerfleet/docs/design/technical-design.md
- reference/workerfleet/docs/design/technical-design.zh-CN.md
- reference/workerfleet/deploy
Depends On: 2046

Goal:
Use MongoDB-backed distributed leases so workerfleet background loop families can safely run in multi-replica deployments.

Scope:
Add runtime lease config, Mongo lease collection/indexes, a Mongo `LoopLeaseCoordinator`, and wire it into kube sync, status sweep, and alert evaluation loops when Mongo storage is enabled.

Non-goals:
- Do not add a non-Mongo lease backend.
- Do not raise deployment replicas above one in this card.
- Do not add lease behavior to Plumego stable roots.

Files:
- reference/workerfleet/internal/app/config.go
- reference/workerfleet/internal/app/bootstrap.go
- reference/workerfleet/internal/platform/store/mongo/lease_store.go
- reference/workerfleet/internal/platform/store/mongo/indexes.go
- reference/workerfleet/docs/design/technical-design.md

Acceptance Tests:
- reference/workerfleet/internal/app/runtime_loops_test.go: TestLoopRunnerSkipsWorkWhenLeaseNotAcquired
- reference/workerfleet/internal/platform/store/mongo/lease_store_test.go: TestLeaseCoordinatorAcquireRenewSteal

Tests:
- Existing app runtime loop and Mongo store tests.

Docs Sync:
- reference/workerfleet/README.md
- reference/workerfleet/docs/design/technical-design.md
- reference/workerfleet/docs/design/technical-design.zh-CN.md
- reference/workerfleet/deploy/README.md

Validation:
- cd reference/workerfleet && go test -timeout 30s ./internal/app ./internal/platform/store/mongo
- cd reference/workerfleet && go vet ./internal/app ./internal/platform/store/mongo
- gofmt -l reference/workerfleet/internal/app reference/workerfleet/internal/platform/store/mongo

Done Definition:
- [ ] Acceptance Tests pass.
- [ ] All Validation commands exit 0.
- [ ] gofmt -l . produces no output.
- [ ] Docs Sync targets updated (if applicable).

Outcome:
