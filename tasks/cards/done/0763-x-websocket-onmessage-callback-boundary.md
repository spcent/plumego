# Card 0763

Milestone:
Recipe: specs/change-recipes/extension-stability.yaml
Priority: P0
State: active
Primary Module: x/websocket
Owned Files: x/websocket/server.go, x/websocket/server_test.go, docs/modules/x-websocket/README.md
Depends On:

Goal:
Make application message callbacks fail contained and stop slow callbacks from implicitly defining the connection read loop contract.

Scope:
- Add a panic boundary around `ServerConfig.OnMessage`.
- Define whether callback execution is synchronous or bounded asynchronous.
- Preserve explicit app callback ownership without introducing hidden globals.
- Add tests for panic containment and slow callback behavior.

Non-goals:
- Add a generic event bus.
- Change websocket auth or room authorization behavior.
- Add non-standard-library dependencies.

Files:
- x/websocket/server.go
- x/websocket/server_test.go
- docs/modules/x-websocket/README.md

Tests:
- go test -timeout 20s ./x/websocket/...

Docs Sync:
- Document the `OnMessage` callback panic/blocking contract if behavior changes.

Done Definition:
- Callback panic cannot crash the process.
- Slow callback behavior is explicit and covered by focused tests.
- Module tests pass.

Outcome:
Implemented a per-connection bounded message callback dispatcher. `OnMessage`
panics are recovered and close only the affected connection, while slow
callbacks fill a bounded queue and close the connection instead of creating
unbounded backlog or implicitly blocking the read path. Added focused tests and
updated module docs.
