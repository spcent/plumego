# Card 2007

Milestone: —
Recipe: specs/change-recipes/analysis-only.yaml
Priority: P1
State: done
Primary Module: reference/workerfleet/internal/platform/notifier
Owned Files:
- `reference/workerfleet/internal/platform/notifier/feishu.go`
- `reference/workerfleet/internal/platform/notifier/webhook.go`
- `reference/workerfleet/internal/platform/notifier/dispatcher.go`
- `reference/workerfleet/internal/platform/notifier/notifier_test.go`
- `reference/workerfleet/docs/notifiers.md`
Depends On:
- `tasks/cards/active/2006-workerfleet-alert-engine.md`

Goal:
- Deliver alert events to Feishu and generic Webhook endpoints with dedupe-aware fire and recovery notifications.

Scope:
- Feishu and Webhook notifier clients.
- Dispatcher logic for firing and recovery notifications.
- App-local notifier configuration and retry-safe delivery interface.

Non-goals:
- Do not add email, Slack, or multi-channel routing policy.
- Do not introduce hidden background globals.
- Do not store secrets in logs or payload history.

Files:
- `reference/workerfleet/internal/platform/notifier/feishu.go`
- `reference/workerfleet/internal/platform/notifier/webhook.go`
- `reference/workerfleet/internal/platform/notifier/dispatcher.go`
- `reference/workerfleet/internal/platform/notifier/notifier_test.go`
- `reference/workerfleet/docs/notifiers.md`

Tests:
- `go test ./reference/workerfleet/internal/platform/notifier/...`
- Tests for message rendering, retry-safe dispatch, fire vs recovery behavior, and secret redaction.

Docs Sync:
- Document notifier configuration, secret handling, and payload examples in `reference/workerfleet/docs/notifiers.md`.

Done Definition:
- Alert events can be sent to Feishu and Webhook destinations without leaking secrets.
- Recovery notifications reuse the same dedupe-aware alert model as firing notifications.
- The notifier layer remains app-local to `reference/workerfleet`.

Outcome:
- Added app-local Feishu and generic webhook notifiers with a shared dispatcher.
- Added focused notifier tests for Feishu payloads, webhook envelopes, fan-out behavior, and secret-safe error handling.
- Documented notifier behavior and delivery rules in `reference/workerfleet/docs/notifiers.md`.
- Validation run: `go test ./reference/workerfleet/internal/platform/notifier/... ./reference/workerfleet/internal/...`
