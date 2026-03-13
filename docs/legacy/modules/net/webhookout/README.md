# Webhook Out Migration

`net/webhookout` is legacy.

Use [`x/webhook`](/Users/bingrong.yan/projects/go/plumego/docs/modules/x-webhook/README.md) instead:

- package path: `github.com/spcent/plumego/x/webhook`
- outbound delivery primitives: `Config`, `Service`, `Target`, `TargetPatch`
- delivery inspection: `Delivery`, `DeliveryFilter`
- in-memory store and queue setup: `NewMemStore`, `NewService`
- explicit app components: `NewWebhookOutComponent`

This page remains only as an archived migration marker for historical links.

```yaml
# env.example
WEBHOOK_QUEUE_SIZE=2048
WEBHOOK_WORKERS=8
WEBHOOK_TIMEOUT_SECONDS=10
WEBHOOK_MAX_RETRIES=5
```

## Related Documentation

- [Inbound Webhooks](/Users/bingrong.yan/projects/go/plumego/docs/legacy/modules/net/webhookin/README.md) - Receiving webhooks
- [Security](/Users/bingrong.yan/projects/go/plumego/docs/modules/security/README.md) - HMAC signature verification
