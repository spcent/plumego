# 0754 - x/websocket message read contract

Status: done
Priority: P1
Primary module: `x/websocket`

## Goal

Make the inbound message read path and its stable contract clearer and less
allocation-heavy.

## Scope

- Reduce the server read loop from buffer-copy-copy to a single owned message
  allocation where possible.
- Keep `ReadMessageStream` documented as bounded reader, not true streaming.
- Update docs and tests around the server read path if needed.

## Non-goals

- Renaming `ReadMessageStream` in this pass.
- Introducing a new streaming business handler API.
- Changing `Message.Data` ownership semantics.

## Files

- `x/websocket/server.go`
- `x/websocket/stream.go`
- `x/websocket/server_config_test.go`
- `docs/modules/x-websocket/README.md`

## Tests

- `go test -timeout 20s ./x/websocket/...`
- `go vet ./x/websocket/...`
- `go build ./...`

## Docs Sync

Document that `ReadMessageStream` remains bounded-reader API and server
handlers receive owned message bytes.

## Done Definition

- Server read loop avoids avoidable duplicate buffers.
- Stable contract wording is precise.
- Validation passes.

## Outcome

- Replaced the registered server handler read path's pooled buffer plus copy
  with a direct owned `Message.Data` allocation via `io.ReadAll`.
- Kept `ReadMessageStream` documented as a bounded reader, not a true streaming
  or zero-copy API.
- Documented that registered server handlers receive owned bytes while still
  reading the complete message into memory.

## Validations

- `go test -timeout 20s ./x/websocket/...`
- `go vet ./x/websocket/...`
- `go build ./...`
