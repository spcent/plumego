# Card 1299

Milestone:
Recipe: specs/change-recipes/extension-stability.yaml
Priority: P1
State: done
Primary Module: x/websocket
Owned Files: x/websocket/hub.go, x/websocket/hub_test.go, docs/modules/x-websocket/README.md
Depends On:

Goal:
Make security event dispatch lifecycle and configuration semantics stable and understandable.

Scope:
- Ensure a blocking `SecurityEventHandler` cannot create unbounded lifecycle ambiguity.
- Align event enqueue, dispatcher startup, and `EnableSecurityEvents` naming/documentation.
- Preserve panic containment and bounded dispatch behavior.
- Add tests for disabled events, blocked handlers, and shutdown behavior.

Non-goals:
- Add an external logging or metrics dependency.
- Turn security events into a generic pubsub system.
- Change auth policy decisions.

Files:
- x/websocket/hub.go
- x/websocket/hub_test.go
- docs/modules/x-websocket/README.md

Tests:
- go test -timeout 20s ./x/websocket/...

Docs Sync:
- Document when security events are emitted and how handler blocking is handled.

Done Definition:
- Event dispatch semantics are bounded and clear.
- Disabled events do not start surprising runtime work.
- Module tests pass.

Outcome:
Aligned security event runtime work with `EnableSecurityEvents`: disabled
events no longer allocate a handler queue, start a dispatcher, or enqueue
events. Documented the bounded single-dispatcher lifecycle and the requirement
that application handlers return. Added a focused disabled-events test.
