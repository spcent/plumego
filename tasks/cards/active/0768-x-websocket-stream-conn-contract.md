# Card 0768

Milestone:
Recipe: specs/change-recipes/extension-stability.yaml
Priority: P1
State: active
Primary Module: x/websocket
Owned Files: x/websocket/stream.go, x/websocket/conn.go, docs/modules/x-websocket/README.md, docs/extension-evidence/x-websocket-public-api-inventory.md
Depends On:

Goal:
Make bounded stream reading and server-side connection semantics explicit before API freeze.

Scope:
- Clarify `ReadMessageStream` is bounded-reader behavior, not low-memory streaming.
- Clarify `Conn`/`NewConnE` are server-side websocket primitives.
- Add or adjust comments and evidence inventory notes so owner review sees the exact contract.
- Avoid exported symbol churn unless necessary.

Non-goals:
- Implement true streaming frame IO.
- Rename exported symbols without owner approval.
- Change RFC6455 frame parsing behavior.

Files:
- x/websocket/stream.go
- x/websocket/conn.go
- docs/modules/x-websocket/README.md
- docs/extension-evidence/x-websocket-public-api-inventory.md

Tests:
- go test -timeout 20s ./x/websocket/...

Docs Sync:
- Sync package comments and module docs for stable API contract language.

Done Definition:
- Stream and Conn contracts are no longer implicit or overstated.
- Evidence inventory reflects the remaining owner decision.
- Module tests pass.

Outcome:

