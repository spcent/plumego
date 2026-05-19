# 0800 - x/websocket validation scope and log safety

Status: done
Priority: P2
State: done
Primary module: `x/websocket`

## Problem

Logging sanitization comments and behavior do not match: newline/tab are still
allowed. `ContainsDangerousPatterns` is an application-content heuristic inside
transport code and cannot be promised as stable websocket behavior.

## Scope

- Make log sanitization remove or encode newline/tab consistently.
- Remove transport-level dangerous-pattern scanning from default message
  validation, or explicitly isolate it as an opt-in helper.
- Update tests for message validation and logging safety.

## Out of Scope

- Application-level content moderation.
- SQL/XSS policy engines.

## Validation

- `go test -timeout 20s ./x/websocket/...`
- `go vet ./x/websocket/...`

## Outcome

- `SanitizeForLogging` now replaces newline and tab characters so sanitized
  values stay single-line.
- `ContainsDangerousPatterns` is documented as an opt-in heuristic helper and
  not part of default transport validation.
- Updated focused validation tests and module primer wording.
