# 0759 - x/websocket hub broadcast inputs

Status: active
Priority: P1
Primary module: `x/websocket`

## Goal

Make public Hub broadcast APIs apply the same input validation style as join
APIs.

## Scope

- Validate room names in `TryBroadcastRoom`.
- Validate text/binary opcode input in `TryBroadcastRoom` and
  `TryBroadcastAll`.
- Keep fire-and-forget `BroadcastRoom`/`BroadcastAll` wrappers intentionally
  discarding errors.
- Add tests for invalid room/opcode paths.

## Non-goals

- Returning errors from `BroadcastRoom`/`BroadcastAll`.
- Changing queue-full behavior.
- Changing room join limits.

## Files

- `x/websocket/hub.go`
- `x/websocket/hub_lifecycle_test.go`
- `docs/modules/x-websocket/README.md`
- `x/websocket/module.yaml`

## Tests

- `go test -timeout 20s ./x/websocket/...`
- `go vet ./x/websocket/...`
- `go build ./...`

## Docs Sync

Document result-returning broadcast validation behavior.

## Done Definition

- Result-returning broadcast APIs fail visibly on invalid inputs.
- Wrapper APIs remain fire-and-forget.
- Validation passes.
