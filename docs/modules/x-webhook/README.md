# X Webhook

> **Package**: `github.com/spcent/plumego/x/webhook` | **Layer**: Extension

`x/webhook` is the application-facing webhook surface for Plumego.

Use this package for:
- inbound verification helpers such as `VerifyHMAC`, `VerifyGitHub`, and `VerifyStripe`
- outbound delivery primitives such as `Config`, `Service`, `Target`, and `NewMemStore`
- explicit component mounting through `NewWebhookInComponent` and `NewWebhookOutComponent`
- PubSub bridging through `WebhookBridge`

Do not import legacy `net/webhookin` or `net/webhookout` from new application code.

Related docs:
- [Net Webhook In](../net/webhookin/README.md)
- [Net Webhook Out](../net/webhookout/README.md)
