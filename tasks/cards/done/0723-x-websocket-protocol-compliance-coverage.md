# 0723 - x/websocket protocol compliance coverage

Status: done
Priority: P1
Primary module: `x/websocket`

## Problem

Handshake validation does not require `Sec-WebSocket-Version: 13`, and stable
readiness needs stronger RFC6455 negative and boundary coverage for handshake
and frame validation.

## Scope

- Reject missing or non-13 websocket versions during handshake.
- Add focused protocol negative tests for handshake headers and frame boundary
  validation.
- Add fuzz or table-driven boundary coverage where it fits the existing test
  style.

## Out of Scope

- Third-party websocket dependency adoption.
- Full Autobahn test-suite integration.

## Validation

- `go test -timeout 20s ./x/websocket/...`
- `go vet ./x/websocket/...`

## Outcome

- Added handshake rejection for missing or non-13 `Sec-WebSocket-Version`.
- Added structured negative handshake coverage for unsupported versions.
- Updated existing valid handshake fixtures and module primer notes.
