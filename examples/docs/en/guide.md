# Plumego documentation map (English)

This directory now mirrors the plumego code layout so each module has its own focused guide. Start here and jump into the module-specific pages as needed.

## Module index
- [Core](core.md): lifecycle wiring, configuration options, components, and troubleshooting.
- [Router](router.md): trie matcher, route groups, parameters, and middleware inheritance.
- [Middleware](middleware.md): built-in guards, composition rules, and customization template.
- [Pub/Sub & WebSocket](pubsub-websocket.md): in-process bus, WebSocket hub, and real-time fan-out patterns.
- [Metrics & Health](metrics-health.md): Prometheus/OpenTelemetry adapters plus readiness/build endpoints.
- [Security & Webhook](security-webhook.md): inbound signature checks and outbound delivery manager.

## How to use the examples
All snippets reference the demo app in `examples/reference`. Run it directly to validate behavior:

```bash
go run ./examples/reference
```

## Notes on style
- Keep boundaries clear: business logic lives in your own `main`, plumego supplies the scaffolding.
- Register routes and middleware before calling `Boot()`; late registrations panic by design.
- Prefer configuration via `core.With...` options or the environment variables documented in `env.example`.
