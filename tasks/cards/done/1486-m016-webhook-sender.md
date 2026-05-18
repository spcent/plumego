# Card 1563

Milestone: M-016
Recipe: specs/change-recipes/add-http-endpoint.yaml
Priority: P2
State: done
Primary Module: reference/with-events
Owned Files:
- `reference/with-events/internal/webhook/sender.go`
- `reference/with-events/internal/webhook/sender_test.go`
- `reference/with-events/README.md`

Goal:
- Implement the webhook delivery component in reference/with-events using
  x/messaging/webhook with exponential-backoff retry, and write the
  reference/with-events/README.md with architecture diagram and run instructions.

Scope:
- Create internal/webhook/sender.go:
  - WebhookSender struct holding an x/messaging/webhook.Sender and target URL
    read from config.WebhookTargetURL.
  - Send(ctx context.Context, event OrderCreated) error — serialises the event
    as JSON and delivers via webhook.Sender with up to 3 retries using
    exponential backoff (1s, 2s, 4s).
  - If WebhookTargetURL is empty, Send is a no-op (skips delivery silently).
- Create internal/webhook/sender_test.go using httptest.NewServer as the target:
  - Successful delivery on first attempt returns nil.
  - Target returns 500 twice, succeeds on third attempt — retries work.
  - Target always returns 500: Send returns error after MaxRetries.
  - Empty WebhookTargetURL: Send returns nil without making any HTTP call.
- Create reference/with-events/README.md with:
  - ASCII architecture diagram showing: HTTP → handler → publisher → pubsub →
    consumer → (success: done | failure: scheduler → retry → publisher) and
    consumer → webhook sender → external target.
  - How to run: `go run ./...` with optional WEBHOOK_TARGET_URL env var.
  - Pattern explanations: outbox, at-least-once consumer, delayed retry,
    webhook delivery.
  - How to swap in-process pubsub for a real broker.

Non-goals:
- Do not add HMAC signature to webhook delivery (can be added later).
- Do not persist delivery attempts beyond the retry loop.

Files:
- `reference/with-events/internal/webhook/sender.go`
- `reference/with-events/internal/webhook/sender_test.go`
- `reference/with-events/README.md`

Tests:
- `go test -timeout 30s ./reference/with-events/internal/webhook/...`
- `go test -timeout 30s ./reference/with-events/...`
- `go vet ./reference/with-events/...`
- `go build ./reference/with-events/...`

Docs Sync:
- reference/with-events/README.md is written in this card.

Done Definition:
- Retry logic backs off correctly; stops at MaxRetries.
- Empty URL is a no-op.
- All four sender_test.go cases pass.
- reference/with-events/README.md has ASCII diagram and run instructions.
- `go build ./reference/with-events/...` exits 0.

Outcome:
- Implemented reference/with-events webhook delivery with JSON POST, empty-target
  no-op behavior, and retryable HTTP/network failure handling using
  x/messaging/webhook retry rules.
- Wired the order consumer to an optional webhook sender from
  WEBHOOK_TARGET_URL so the reference flow includes consumer -> webhook target
  delivery.
- Added httptest coverage for first-attempt success, two failures then success,
  exhausted 500 retries, and empty target no-op.
- Wrote reference/with-events/README.md with ASCII architecture diagram, run
  instructions, and notes for outbox, at-least-once consumer, delayed retry,
  webhook delivery, and replacing in-process pubsub.
- Validation:
  `GOTOOLCHAIN=go1.24.4 GOCACHE=/private/tmp/plumego-gocache go test -timeout 30s ./internal/webhook/...`
  from reference/with-events; `GOTOOLCHAIN=go1.24.4
  GOCACHE=/private/tmp/plumego-gocache go test -timeout 30s ./...` from
  reference/with-events; `GOTOOLCHAIN=go1.24.4
  GOCACHE=/private/tmp/plumego-gocache go vet ./...` from
  reference/with-events; `GOTOOLCHAIN=go1.24.4
  GOCACHE=/private/tmp/plumego-gocache go build ./...` from
  reference/with-events; `GOTOOLCHAIN=go1.24.4
  GOCACHE=/private/tmp/plumego-gocache go run ./internal/checks/reference-layout`;
  `GOTOOLCHAIN=go1.24.4 GOCACHE=/private/tmp/plumego-gocache go run
  ./internal/checks/agent-workflow`; `gofmt -l .`; `git diff --check`.
