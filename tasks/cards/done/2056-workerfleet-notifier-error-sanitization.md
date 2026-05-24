# Card 2056

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Context Package: implementation
Priority: P0
State: done
Primary Module: reference/workerfleet
Owned Files:
- reference/workerfleet/internal/platform/notifier/dispatcher.go
- reference/workerfleet/internal/platform/notifier/feishu.go
- reference/workerfleet/internal/platform/notifier/webhook.go
- reference/workerfleet/internal/platform/notifier/notifier_test.go
- reference/workerfleet/docs/notifiers.md
Depends On: 2055

Goal:
Prevent notifier delivery errors from persisting or surfacing raw remote response bodies that could contain secrets or high-cardinality data.

Scope:
Change HTTP notifier error construction to keep sink name, status code, error class, and permanence while omitting raw response payload text from `Error()`. Keep response bodies drained with bounded reads so HTTP connections remain reusable. Add tests proving a remote response body containing a configured header secret does not appear in the error string.

Non-goals:
- Do not change notifier request payloads.
- Do not add redaction regexes or secret scanners.
- Do not change outbox retry classification.

Files:
- reference/workerfleet/internal/platform/notifier/dispatcher.go
- reference/workerfleet/internal/platform/notifier/feishu.go
- reference/workerfleet/internal/platform/notifier/webhook.go
- reference/workerfleet/internal/platform/notifier/notifier_test.go
- reference/workerfleet/docs/notifiers.md

Acceptance Tests:
- reference/workerfleet/internal/platform/notifier/notifier_test.go: TestNotifierErrorDoesNotIncludeRemoteBody

Tests:
- Existing notifier tests.

Docs Sync:
- reference/workerfleet/docs/notifiers.md

Validation:
- cd reference/workerfleet && go test -timeout 30s ./internal/platform/notifier
- cd reference/workerfleet && go vet ./internal/platform/notifier
- gofmt -l reference/workerfleet/internal/platform/notifier

Done Definition:
- [x] Acceptance Tests pass.
- [x] All Validation commands exit 0.
- [x] gofmt -l . produces no output.
- [x] Docs Sync targets updated (if applicable).

Outcome:
- Sanitized HTTP notifier delivery errors so remote response bodies are drained but omitted from `Error()`.
- Preserved sink, status code, error class, and permanence for retry classification.
- Added `TestNotifierErrorDoesNotIncludeRemoteBody`.
- Documented that notifier errors omit raw remote response bodies.
- Validation:
  - `cd reference/workerfleet && go test -timeout 30s ./internal/platform/notifier`
  - `cd reference/workerfleet && go vet ./internal/platform/notifier`
  - `gofmt -l reference/workerfleet/internal/platform/notifier`
