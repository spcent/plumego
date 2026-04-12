# Card 0951: Log Derived Logger Semantics Alignment

Priority: P1
State: done
Primary Module: log

## Goal

Make derived stable loggers preserve the same behavior as their parent logger so text and JSON backends do not diverge on verbosity semantics.

## Problem

- `log/logger.go` implements `defaultLogger.WithFields(...)` by returning a new `defaultLogger` with the backend and merged fields, but it drops `respectVerbosity`.
- `log/json.go` preserves `respectVerbosity` when deriving a child logger.
- That means the two stable backends do not behave the same:
  - a text logger created with `RespectVerbosity: true` can lose debug gating after `With(...)` or `WithFields(...)`
  - a JSON logger keeps the gating
- This is exactly the kind of silent semantic drift that the stable `log` root should avoid.

## Scope

- Fix `defaultLogger.WithFields(...)` and any related text-logger derivation path so child loggers preserve parent semantics.
- Add regression tests covering:
  - text logger + `RespectVerbosity: true` + derived child logger
  - parity between text and JSON child logger behavior for the same configuration
- Review other derived logger state in the stable `log` root and preserve any additional parent-owned semantics that should not reset on `With(...)`.

## Non-Goals

- Do not redesign `gLogger`.
- Do not add global logger state or new backend formats.
- Do not change fatal/exit semantics.

## Files

- `log/logger.go`
- `log/logger_semantics_test.go`
- `log/json_test.go`

## Tests

- `go test -timeout 20s ./log/...`
- `go test -race -timeout 60s ./log/...`
- `go vet ./log/...`

## Docs Sync

- None expected unless the public logger configuration story changes.

## Done Definition

- Derived text loggers preserve `RespectVerbosity` and any other parent-owned stable semantics.
- Text and JSON stable loggers behave consistently after `With(...)` / `WithFields(...)`.
- Regression tests fail before the fix and pass after it.

## Outcome

- Fixed `defaultLogger.WithFields(...)` so derived text loggers preserve `respectVerbosity`.
- Added a regression test that exercises derived child loggers across both text and JSON formats.
- Verified that child loggers now keep both field attachment and verbosity gating semantics.

## Validation Run

```bash
gofmt -w log/logger.go log/logger_semantics_test.go
go test -timeout 20s ./log/...
go test -race -timeout 60s ./log/...
go vet ./log/...
```
