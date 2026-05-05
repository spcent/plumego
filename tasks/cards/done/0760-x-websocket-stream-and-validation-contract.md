# 0760 - x/websocket stream and validation contract

Status: done
Priority: P1
Primary module: `x/websocket`

## Goal

Remove misleading stable-facing language around message reading and text
validation.

## Scope

- Keep `ReadMessageStream` as a bounded-reader API and document that it is not
  a low-memory streaming API.
- Keep current read implementation but make comments precise about frame
  buffering and complete-message memory use.
- Reword text validation comments away from XSS guarantees and toward
  transport-level validation.

## Non-goals

- Implementing true streaming reads.
- Renaming public APIs.
- Changing message validation defaults.

## Files

- `x/websocket/stream.go`
- `x/websocket/validation.go`
- `docs/modules/x-websocket/README.md`
- `docs/extension-evidence/x-websocket.md`

## Tests

- `go test -timeout 20s ./x/websocket/...`
- `go vet ./x/websocket/...`
- `go build ./...`

## Docs Sync

Keep docs explicit that stable high-throughput large-message support is not yet
claimed.

## Done Definition

- Comments and docs no longer imply true streaming or XSS protection.
- Current bounded-reader behavior remains tested.
- Validation passes.

## Outcome

- Updated stream comments to state that `ReadMessageStream` is bounded but not
  low-memory or zero-copy.
- Reworded text validation comments as transport-level validation, not XSS or
  application content policy.
- Updated README and evidence language to avoid true-streaming claims.

## Validations

- `go test -timeout 20s ./x/websocket/...`
- `go vet ./x/websocket/...`
- `go build ./...`
