# Card 2049

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Context Package: implementation
Priority: P1
State: done
Primary Module: reference/workerfleet
Owned Files:
- reference/workerfleet/internal/platform/kube
- reference/workerfleet/internal/app/runtime_loops.go
- reference/workerfleet/docs/design/technical-design.md
- reference/workerfleet/docs/design/technical-design.zh-CN.md
- reference/workerfleet/docs/metrics.md
Depends On: 2048

Goal:
Upgrade the app-local Kubernetes adapter to model list/watch lifecycle, relist recovery, and deleted pod events clearly.

Scope:
Add watch event handling for `ADDED`, `MODIFIED`, `DELETED`, `BOOKMARK`, and Kubernetes `ERROR`, recover from expired resource versions with relist, and keep the implementation stdlib-only.

Non-goals:
- Do not introduce `controller-runtime` or Kubernetes client-go.
- Do not build a generic Kubernetes controller framework.
- Do not change Plumego stable roots.

Files:
- reference/workerfleet/internal/platform/kube/watch.go
- reference/workerfleet/internal/platform/kube/discovery.go
- reference/workerfleet/internal/platform/kube/watch_test.go
- reference/workerfleet/internal/app/runtime_loops.go
- reference/workerfleet/docs/design/technical-design.md

Acceptance Tests:
- reference/workerfleet/internal/platform/kube/watch_test.go: TestInventoryWatchRelistsAfterExpiredResourceVersion
- reference/workerfleet/internal/platform/kube/watch_test.go: TestInventoryWatchAppliesDeletedPodEvent

Tests:
- Watch stream exits on context cancellation.
- Bearer token is not included in error messages or metric labels.

Docs Sync:
- reference/workerfleet/docs/design/technical-design.md
- reference/workerfleet/docs/design/technical-design.zh-CN.md
- reference/workerfleet/docs/metrics.md

Validation:
- cd reference/workerfleet && go test -timeout 30s ./internal/platform/kube ./internal/app
- cd reference/workerfleet && go vet ./internal/platform/kube ./internal/app
- gofmt -l reference/workerfleet/internal/platform/kube reference/workerfleet/internal/app

Done Definition:
- [x] Acceptance Tests pass.
- [x] All Validation commands exit 0.
- [x] gofmt -l . produces no output.
- [x] Docs Sync targets updated (if applicable).

Outcome:
Upgraded the app-local Kubernetes adapter from list-only sync plus raw watch callback into an explicit list/watch lifecycle. The kube sync now lists first, watches from `resourceVersion`, handles `ADDED`, `MODIFIED`, `DELETED`, `BOOKMARK`, and Kubernetes `ERROR`, relists on expired resource versions, marks deleted pod snapshots, exits cleanly on cancellation, and avoids reflecting bearer tokens through list/watch HTTP errors.

Validation:
- `cd reference/workerfleet && go test -timeout 30s ./internal/platform/kube ./internal/app`
- `cd reference/workerfleet && go vet ./internal/platform/kube ./internal/app`
- `gofmt -l reference/workerfleet/internal/platform/kube reference/workerfleet/internal/app`
- `git diff --check`
