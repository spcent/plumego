# 0716 - x/websocket room identity validation

Status: active
Priority: P1
Primary module: `x/websocket`

## Problem

Room names come directly from the request query string with no length or
character policy, and room passwords are transported as URL query parameters.
That creates high-cardinality room keys, logging pollution, and credential
leakage risks.

## Scope

- Introduce a single room-name validator and apply it during handshake.
- Replace query room-password handling with an explicit header-based transport.
- Add structured rejection tests for invalid room names and missing/wrong room
  credentials.
- Document the accepted room name and credential transport contract.

## Out of Scope

- Tenant/session ownership.
- Persistent room registry.

## Validation

- `go test -timeout 20s ./x/websocket/...`
- `go vet ./x/websocket/...`

