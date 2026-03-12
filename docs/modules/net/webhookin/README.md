# Webhook In Migration

`net/webhookin` is legacy.

Use [`x/webhook`](../x-webhook/README.md) instead:

- package path: `github.com/spcent/plumego/x/webhook`
- inbound verification helpers: `VerifyHMAC`, `VerifyGitHub`, `VerifyStripe`
- replay and IP controls: `NewDeduper`, `NewIPAllowlist`, `NewMemoryNonceStore`
- explicit app components: `NewWebhookInComponent`

This page remains only as a migration marker for historical links.
