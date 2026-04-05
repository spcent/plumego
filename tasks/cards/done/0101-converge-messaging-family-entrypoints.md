# Card 0101

Priority: P1

Goal:
- Make the messaging family consistently discoverable from `x/messaging` while keeping sibling primitive packages explicitly subordinate.

Scope:
- messaging family metadata
- messaging family primers
- app-facing entrypoint guidance

Non-goals:
- Do not move runtime code between packages in this card.
- Do not redesign queue or broker internals.

Files:
- `specs/extension-taxonomy.yaml`
- `x/messaging/module.yaml`
- `x/mq/module.yaml`
- `x/pubsub/module.yaml`
- `x/webhook/module.yaml`

Tests:
- `go run ./internal/checks/module-manifests`
- `go run ./internal/checks/agent-workflow`

Docs Sync:
- Sync the relevant primers in `docs/modules/x-messaging/README.md` and `docs/modules/x-webhook/README.md` only if the metadata wording changes materially.

Done Definition:
- `x/messaging` is the unambiguous app-facing family entrypoint.
- Sibling packages are explicitly documented as subordinate or primitive.
- Specs, manifests, and primers say the same thing.
