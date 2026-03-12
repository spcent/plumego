# Webhook Out Migration

`net/webhookout` is legacy.

Use [`x/webhook`](../x-webhook/README.md) instead:

- package path: `github.com/spcent/plumego/x/webhook`
- outbound delivery primitives: `Config`, `Service`, `Target`, `TargetPatch`
- delivery inspection: `Delivery`, `DeliveryFilter`
- in-memory store and queue setup: `NewMemStore`, `NewService`
- explicit app components: `NewWebhookOutComponent`

This page remains only as a migration marker for historical links.

```yaml
# env.example
WEBHOOK_QUEUE_SIZE=2048     # Outbound queue buffer
WEBHOOK_WORKERS=8           # Delivery worker count
WEBHOOK_TIMEOUT_SECONDS=10  # Per-request timeout
WEBHOOK_MAX_RETRIES=5       # Retry attempts
```

## Related Documentation

- [Inbound Webhooks](../webhookin/README.md) — Receiving webhooks
- [Security](../../security/) — HMAC signature verification
- [Scheduler](../../scheduler/README.md) — Scheduled webhook delivery
