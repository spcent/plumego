# Card 2059

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Context Package: implementation
Priority: P2
State: active
Primary Module: reference/workerfleet
Owned Files:
- reference/workerfleet/internal/platform/metrics/instrumentation.go
- reference/workerfleet/internal/platform/metrics/instrumentation_test.go
- reference/workerfleet/internal/platform/notifier/dispatcher.go
- reference/workerfleet/docs/metrics.md
- reference/workerfleet/docs/notifiers.md
Depends On: 2058

Goal:
Make runtime error metrics preserve notifier delivery error classes without introducing high-cardinality labels.

Scope:
Expose a small error-class contract from notifier delivery errors and teach `runtimeErrorClass` to map those classes into `workerfleet_runtime_errors_total`. Keep non-notifier errors on existing low-cardinality classes such as `deadline_exceeded`, `canceled`, and `operation_failed`.

Non-goals:
- Do not put raw error messages, URLs, response bodies, worker IDs, or task IDs into metric labels.
- Do not change notification retry semantics.
- Do not change metric names.

Files:
- reference/workerfleet/internal/platform/metrics/instrumentation.go
- reference/workerfleet/internal/platform/metrics/instrumentation_test.go
- reference/workerfleet/internal/platform/notifier/dispatcher.go
- reference/workerfleet/docs/metrics.md
- reference/workerfleet/docs/notifiers.md

Acceptance Tests:
- reference/workerfleet/internal/platform/metrics/instrumentation_test.go: TestRuntimeErrorObserverUsesNotifierErrorClass

Tests:
- Existing metrics and notifier tests.

Docs Sync:
- reference/workerfleet/docs/metrics.md
- reference/workerfleet/docs/notifiers.md

Validation:
- cd reference/workerfleet && go test -timeout 30s ./internal/platform/metrics ./internal/platform/notifier
- cd reference/workerfleet && go vet ./internal/platform/metrics ./internal/platform/notifier
- gofmt -l reference/workerfleet/internal/platform/metrics reference/workerfleet/internal/platform/notifier

Done Definition:
- [ ] Acceptance Tests pass.
- [ ] All Validation commands exit 0.
- [ ] gofmt -l . produces no output.
- [ ] Docs Sync targets updated (if applicable).

Outcome:
