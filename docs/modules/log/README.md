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

## Canonical change shape

- preserve `StructuredLogger` and the canonical `NewLogger` path
- use `LoggerConfig.Format` to select text/json/discard backends instead of parallel constructors
- never log secrets, tokens, signatures, or private keys
- keep reusable test logging helpers out of stable `log`; use `x/observability/testlog`
- attach `request_id` and other transport metadata explicitly at call sites; stable `log` must not read them from context
