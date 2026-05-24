# Card 2052

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Context Package: implementation
Priority: P0
State: done
Primary Module: reference/workerfleet
Owned Files:
- reference/workerfleet/internal/app/alert_loop.go
- reference/workerfleet/internal/app/alert_loop_test.go
- reference/workerfleet/internal/app/config.go
- reference/workerfleet/docs/alerts.md
- reference/workerfleet/docs/design/technical-design.md
Depends On: 2051

Goal:
Allow notification delivery to drain durable outbox jobs independently from alert evaluation being enabled.

Scope:
Start the notification delivery loop whenever notification delivery is enabled, keep alert evaluation loop separately gated, and document the independent runtime behavior.

Non-goals:
- Do not split workerfleet into separate API and worker binaries.
- Do not add a new process supervisor.
- Do not change notifier sink payloads.

Files:
- reference/workerfleet/internal/app/alert_loop.go
- reference/workerfleet/internal/app/alert_loop_test.go
- reference/workerfleet/internal/app/config.go
- reference/workerfleet/docs/alerts.md
- reference/workerfleet/docs/design/technical-design.md

Acceptance Tests:
- reference/workerfleet/internal/app/alert_loop_test.go: TestStartAlertLoopRunsNotificationDeliveryWithoutAlertEvaluation

Tests:
- Existing alert loop tests.

Docs Sync:
- reference/workerfleet/docs/alerts.md
- reference/workerfleet/docs/design/technical-design.md
- reference/workerfleet/docs/design/technical-design.zh-CN.md

Validation:
- cd reference/workerfleet && go test -timeout 30s ./internal/app
- cd reference/workerfleet && go vet ./internal/app
- gofmt -l reference/workerfleet/internal/app

Done Definition:
- [x] Acceptance Tests pass.
- [x] All Validation commands exit 0.
- [x] gofmt -l . produces no output.
- [x] Docs Sync targets updated (if applicable).

Outcome:
- Split `AlertRunner.Start` gating so alert evaluation and notification delivery loops start from their own runtime switches.
- Added `TestStartAlertLoopRunsNotificationDeliveryWithoutAlertEvaluation`.
- Documented independent notification-drain behavior in alerts and technical design docs.
- Validation:
  - `cd reference/workerfleet && go test -timeout 30s ./internal/app`
  - `cd reference/workerfleet && go vet ./internal/app`
  - `gofmt -l reference/workerfleet/internal/app`
