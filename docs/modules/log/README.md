# log

## Purpose

`log` defines structured logging contracts and base implementations.

## v1 Status

- `GA` in the Plumego v1 support matrix
- Public compatibility is expected for the stable package surface

## Use this module when

- adjusting logger interfaces
- changing base logger behavior
- wiring structured fields at the library boundary

## Do not use this module for

- secret-bearing payload logging
- business event schema ownership
- feature-specific transport logic
- request id generation

## First files to read

- `log/module.yaml`
- `log/*.go`
- `AGENTS.md` security rules

## Stable behavior

- `NewLogger` is the canonical constructor for text, JSON, and discard logging.
- `LoggerConfig.Fields` are copied into each logger at construction.
- Per-entry fields override logger fields.
- When a log call receives multiple `Fields` arguments, they are merged in
  order and later maps override earlier maps.
- Text output sorts field keys and escapes ambiguous keys or values so field
  suffixes stay deterministic and single-line.
- JSON output owns the reserved `time`, `level`, and `msg` keys.
- JSON output stringifies individual field values that `encoding/json` cannot
  encode, preserving the rest of the log entry.
- Context-aware methods keep the request-scoped call shape but do not read
  transport metadata from `context.Context`.

## Canonical change shape

- preserve `StructuredLogger` and the canonical `NewLogger` path
- use `LoggerConfig.Format` to select text/json/discard backends instead of parallel constructors
- keep text logger backend ownership constructor-local; do not route the stable path through a package-global singleton
- keep lifecycle/start-stop hooks and CLI flag bootstrap out of the stable public path
- never log secrets, tokens, signatures, or private keys
- keep reusable test logging helpers out of stable `log`; use `x/observability/testlog`
- attach `request_id` and other transport metadata explicitly at call sites; stable `log` must not read them from context
