# 0719 - x/websocket close and stream semantics

Status: done
Priority: P2
Primary module: `x/websocket`

## Problem

`WriteClose` claims a proper closing handshake but writes a close frame and then
immediately closes TCP. `ReadMessageStream`/reader wording implies streaming even
though continuation payloads are bounded in memory.

## Scope

- Align close API comments and docs with best-effort close-frame behavior, or
  implement a real handshake if the existing lifecycle allows it.
- Rename or document bounded-reader semantics clearly.
- Add tests around close frame emission and bounded read behavior.

## Out of Scope

- New protocol engine.
- Zero-copy streaming reader implementation.

## Validation

- `go test -timeout 20s ./x/websocket/...`
- `go vet ./x/websocket/...`

## Outcome

- Clarified `WriteClose` as a best-effort close frame followed by TCP close.
- Clarified `ReadMessageStream` as a bounded reader rather than zero-copy
  streaming.
- Added focused close-frame test coverage.
- Updated module primer semantics.
