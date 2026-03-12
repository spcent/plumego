# X WebSocket

> **Package**: `github.com/spcent/plumego/x/websocket` | **Layer**: Extension

`x/websocket` is the application-facing websocket surface for Plumego.

Use this package for:
- websocket hub primitives such as `Hub`, `HubConfig`, `NewHub`, and `NewHubWithConfig`
- auth helpers such as `NewSimpleRoomAuth` and `NewSecureRoomAuth`
- server entrypoints such as `ServeWSWithAuth` and `ServeWSWithConfig`
- explicit component mounting through `NewComponent`

Do not import legacy `net/websocket` from new application code.

Related docs:
- [Net WebSocket](../net/websocket/README.md)
