# X Webhook

> **Package**: `github.com/spcent/plumego/x/webhook` | **Layer**: Extension

`x/webhook` is the application-facing webhook surface for Plumego.

Use this package for:
- inbound verification helpers such as `VerifyHMAC`, `VerifyGitHub`, and `VerifyStripe`
- outbound delivery primitives such as `Config`, `Service`, `Target`, and `NewMemStore`
- explicit component mounting through `NewWebhookInComponent` and `NewWebhookOutComponent`
- PubSub bridging through `WebhookBridge`

Do not import legacy `net/webhookin` or `net/webhookout` from new application code.

Migration note:
- older links under `docs/modules/net/webhookin` and `docs/modules/net/webhookout` now redirect here conceptually
