# Card 0930: Log Logger Semantics Convergence

Priority: P1
State: active
Primary Module: log

## Goal

Make `log.NewLogger` produce concrete logger implementations with one clear behavioral contract across text, JSON, and discard formats.

## Problem

`log` already has one canonical constructor path through `NewLogger(LoggerConfig)`, but the concrete implementations do not fully honor the same semantics:

- `StructuredLogger.Fatal` and `FatalCtx` document that they log and then call `os.Exit(1)`.
- `jsonLogger` and `discardLogger` exit on fatal calls.
- `defaultLogger` writes a `FATAL` entry but does not exit, violating the public interface contract.
- `LoggerConfig.ErrorOutput` is honored by JSON output but ignored by the text logger.
- `LoggerConfig.RespectVerbosity` changes JSON debug behavior, while the text logger uses its backend verbosity path regardless of that flag.

This creates format-dependent runtime behavior behind a single stable constructor.

## Scope

- Enumerate all `StructuredLogger` concrete implementations and fatal tests before editing.
- Make `Fatal` and `FatalCtx` semantics consistent across text, JSON, and discard implementations.
- Align common `LoggerConfig` fields across text and JSON formats where the field is format-independent: `Level`, `Fields`, `Verbosity`, `RespectVerbosity`, `Output`, and `ErrorOutput`.
- Keep discard format intentionally side-effect-free except for the fatal-exit contract.
- Add tests that prove fatal semantics and config behavior do not diverge by format.
- Update log docs or module hints if any config field remains intentionally format-specific.

## Non-Goals

- Do not add external logging dependencies.
- Do not add package-global default logger facades.
- Do not add context-derived transport metadata to `log`; request metadata remains explicitly supplied by callers.
- Do not change the `StructuredLogger` method set unless the full exported-symbol protocol is followed.

## Expected Files

- `log/logger.go`
- `log/glog.go`
- `log/json.go`
- `log/noop.go`
- `log/*_test.go`
- `docs/modules/log/README.md` if public behavior guidance changes

## Validation

Run focused gates first:

```bash
go test -timeout 20s ./log
go test -race -timeout 60s ./log
go vet ./log
```

Then run the required repo-wide gates before committing.

## Done Definition

- All concrete `StructuredLogger` implementations obey the documented fatal behavior.
- Shared `LoggerConfig` fields have the same meaning across supported formats, or a format-specific exception is explicitly documented.
- Tests cover fatal calls and config behavior for text, JSON, and discard formats.
- Focused gates and repo-wide gates pass.
