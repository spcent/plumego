# Card 1271

Milestone:
Recipe: specs/change-recipes/extension-stability.yaml
Priority: P0
State: active
Primary Module: x/websocket
Owned Files: x/websocket/websocket.go, x/websocket/websocket_test.go, docs/modules/x-websocket/README.md
Depends On:

Goal:
Make route registration failure visible without surprising partial success and make admin broadcast report runtime failures.

Scope:
- Reduce or document `RegisterRoutes` partial-registration risk around multi-route setup.
- Use result-returning broadcast APIs in the admin route.
- Map stopped hub, invalid target, and no-accepted-message results to visible HTTP outcomes.
- Add focused registration and admin broadcast tests.

Non-goals:
- Replace the router package.
- Add rollback APIs to stable router.
- Change public broadcast endpoint shape beyond error visibility.

Files:
- x/websocket/websocket.go
- x/websocket/websocket_test.go
- docs/modules/x-websocket/README.md

Tests:
- go test -timeout 20s ./x/websocket/...

Docs Sync:
- Document admin broadcast status semantics if response behavior changes.

Done Definition:
- Static route config is validated before registration.
- Admin broadcast no longer reports success when dispatch failed.
- Module tests pass.

Outcome:
Added route conflict preflight for registrars exposing `Routes()` so common
multi-route registration conflicts fail before any websocket route is added.
Changed the admin broadcast endpoint to dispatch through `TryBroadcast*` and
return visible outcomes for no recipients, stopped hubs, and dropped delivery.
Updated tests and module docs.
