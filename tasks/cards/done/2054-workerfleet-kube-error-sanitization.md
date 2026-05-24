# Card 2054

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Context Package: implementation
Priority: P1
State: done
Primary Module: reference/workerfleet
Owned Files:
- reference/workerfleet/internal/platform/kube/watch.go
- reference/workerfleet/internal/platform/kube/watch_test.go
- reference/workerfleet/docs/design/technical-design.md
- reference/workerfleet/docs/design/technical-design.zh-CN.md
Depends On: 2053

Goal:
Prevent Kubernetes watch ERROR status messages from propagating raw server text that could contain sensitive data.

Scope:
Return only low-cardinality Kubernetes watch error code/reason details, preserve expired resource version handling, and document the sanitized error contract.

Non-goals:
- Do not change list/watch request authentication.
- Do not add client-go or controller-runtime.
- Do not expose raw Kubernetes status bodies in logs or metrics.

Files:
- reference/workerfleet/internal/platform/kube/watch.go
- reference/workerfleet/internal/platform/kube/watch_test.go
- reference/workerfleet/docs/design/technical-design.md
- reference/workerfleet/docs/design/technical-design.zh-CN.md

Acceptance Tests:
- reference/workerfleet/internal/platform/kube/watch_test.go: TestClientWatchErrorDoesNotIncludeStatusMessage

Tests:
- Existing kube watch tests.

Docs Sync:
- reference/workerfleet/docs/design/technical-design.md
- reference/workerfleet/docs/design/technical-design.zh-CN.md

Validation:
- cd reference/workerfleet && go test -timeout 30s ./internal/platform/kube
- cd reference/workerfleet && go vet ./internal/platform/kube
- gofmt -l reference/workerfleet/internal/platform/kube

Done Definition:
- [x] Acceptance Tests pass.
- [x] All Validation commands exit 0.
- [x] gofmt -l . produces no output.
- [x] Docs Sync targets updated (if applicable).

Outcome:
- Changed Kubernetes watch ERROR handling to return only status code and reason for non-expired errors.
- Preserved resource-version expiry relist behavior.
- Added `TestClientWatchErrorDoesNotIncludeStatusMessage`.
- Documented the sanitized watch error contract in English and Chinese technical design docs.
- Validation:
  - `cd reference/workerfleet && go test -timeout 30s ./internal/platform/kube`
  - `cd reference/workerfleet && go vet ./internal/platform/kube`
  - `gofmt -l reference/workerfleet/internal/platform/kube`
