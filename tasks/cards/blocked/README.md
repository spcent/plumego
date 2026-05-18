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
