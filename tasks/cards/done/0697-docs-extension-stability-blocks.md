# 0697 - Standardize Extension Stability Blocks

State: done
Priority: P1
Primary module: website docs

## Goal

Make extension adoption risk explicit across `x/*` primer pages by adding a
consistent stability block that points readers to `module.yaml`, release posture,
and extension maturity before they depend on an experimental package.

## Scope

- Add a short, consistent stability section to the main extension primer pages.
- Keep `x/ai` as the richer example because it has subpackage stability tiers.
- Avoid implying that checked-in extension code is part of the default stable
  service path.
- Mirror updates in Chinese pages.

## Non-goals

- Do not change `module.yaml` files.
- Do not change release policy.
- Do not create a new generated sync system unless the current component
  patterns already make it straightforward.

## Page Revision Checklist

| Target page | Modification goal | Acceptance points |
| --- | --- | --- |
| `website/src/content/docs/docs/modules/x-rest.mdx` | Add a stability block near the top. | The block states that `x/rest` is an extension surface and readers must check `x/rest/module.yaml`, Release Posture, and Extension Maturity before production adoption. |
| `website/src/content/docs/zh/docs/modules/x-rest.mdx` | Add the localized stability block. | The warning is clear in Chinese and does not overstate API compatibility. |
| `website/src/content/docs/docs/modules/x-tenant.mdx` | Add a stability and production-adoption block. | The block distinguishes tenant isolation importance from stability guarantees; it recommends app-local interfaces for production adoption when needed. |
| `website/src/content/docs/zh/docs/modules/x-tenant.mdx` | Add the localized stability block. | The Chinese copy preserves the fail-closed isolation caution without implying stable-root status. |
| `website/src/content/docs/docs/modules/x-websocket.mdx` | Add a stability block for WebSocket transport adoption. | The block tells readers to check `x/websocket/module.yaml` and avoid treating stateful connection lifecycle as a stable-root concern. |
| `website/src/content/docs/zh/docs/modules/x-websocket.mdx` | Add the localized WebSocket stability block. | The page keeps package names unchanged and localizes guidance text. |
| `website/src/content/docs/docs/modules/x-messaging.mdx` | Add a family-level stability block. | The block says to start from `x/messaging` first, then verify subordinate package maturity before depending on `x/pubsub`, `x/mq`, `x/scheduler`, or `x/webhook`. |
| `website/src/content/docs/zh/docs/modules/x-messaging.mdx` | Add the localized family-level stability block. | Chinese text explains primary family vs subordinate package maturity. |
| `website/src/content/docs/docs/modules/x-gateway.mdx` | Add a gateway stability block. | The block warns that edge/proxy behavior should be wrapped behind application-local config where compatibility matters. |
| `website/src/content/docs/zh/docs/modules/x-gateway.mdx` | Add the localized gateway stability block. | Chinese page preserves the proxy/edge distinction. |
| `website/src/content/docs/docs/modules/x-observability.mdx` | Add an observability adapter stability block. | The block distinguishes stable `middleware` request instrumentation from higher-level exporter/adapter work in `x/observability`. |
| `website/src/content/docs/zh/docs/modules/x-observability.mdx` | Add the localized observability stability block. | Chinese page keeps the stable middleware vs extension adapter boundary clear. |
| `website/src/content/docs/docs/modules/x-ai.mdx` | Review the existing stability table for consistency with the new wording. | The current richer table remains intact; any new wording aligns with other extension primers. |
| `website/src/content/docs/zh/docs/modules/x-ai.mdx` | Review the localized `x/ai` stability table. | The Chinese table remains the model for subpackage-tier explanations. |

## Files

- `website/src/content/docs/docs/modules/x-rest.mdx`
- `website/src/content/docs/zh/docs/modules/x-rest.mdx`
- `website/src/content/docs/docs/modules/x-tenant.mdx`
- `website/src/content/docs/zh/docs/modules/x-tenant.mdx`
- `website/src/content/docs/docs/modules/x-websocket.mdx`
- `website/src/content/docs/zh/docs/modules/x-websocket.mdx`
- `website/src/content/docs/docs/modules/x-messaging.mdx`
- `website/src/content/docs/zh/docs/modules/x-messaging.mdx`
- `website/src/content/docs/docs/modules/x-gateway.mdx`
- `website/src/content/docs/zh/docs/modules/x-gateway.mdx`
- `website/src/content/docs/docs/modules/x-observability.mdx`
- `website/src/content/docs/zh/docs/modules/x-observability.mdx`
- `website/src/content/docs/docs/modules/x-ai.mdx`
- `website/src/content/docs/zh/docs/modules/x-ai.mdx`

## Tests

- `cd website && pnpm check:docs-api` - passed
- `cd website && pnpm check` - passed
- `cd website && pnpm build` - passed

## Planned Tests

- `cd website && pnpm check:docs-api`
- `cd website && pnpm check`

## Docs Sync

This is docs-only work. Keep English and Chinese warning language consistent in
meaning, not word-for-word identical.

## Done Definition

- Primary extension primer pages have consistent stability/adoption language.
- `x/ai` remains the canonical example for subpackage stability tiers.
- No primer implies an experimental extension has a stable-root compatibility promise.
