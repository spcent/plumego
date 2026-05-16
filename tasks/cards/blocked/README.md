# Blocked Task Cards

`tasks/cards/blocked/` holds work that still matters but cannot be executed
without an external prerequisite such as a release ref, owner sign-off, product
decision, or unavailable dependency.

Rules:

- keep `State: blocked` in each blocked card
- keep `Depends On:` or an explicit blocker in the card body
- do not run blocked cards from the active queue
- move a card back to `tasks/cards/active/` only when the blocker is resolved
- update linked evidence docs when moving a card between queues

## Blocked Queue

| Card | Priority | Primary module | Blocker |
|---|---|---|---|
| [1367](1367-x-tenant-beta-evidence-closure.md) | P2 | x/tenant | second release ref and owner sign-off |
| [1370](1370-x-ai-stable-tier-beta-evidence-closure.md) | P2 | x/ai | second release refs and owner sign-off |
| [1371](1371-x-data-surface-beta-evidence-closure.md) | P2 | x/data | second release refs and owner sign-off |
| [1372](1372-x-discovery-surface-beta-evidence-closure.md) | P2 | x/gateway/discovery | second release ref and owner sign-off |
| [1373](1373-x-messaging-service-beta-evidence-closure.md) | P2 | x/messaging | second release ref and owner sign-off |
