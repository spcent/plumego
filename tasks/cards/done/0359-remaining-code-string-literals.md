# Card 0359: Remaining Non-Canonical Code Strings — Value-Changing Replacements

Priority: P3
State: done
Recipe: specs/change-recipes/fix-bug.yaml
Primary Module: x/messaging, x/pubsub, x/websocket, x/scheduler

## Goal

Several `x/` modules still use uppercase string literals in `.Code("…")` calls
where:
1. A canonical constant exists but with a **different string value** — changing
   is a deliberate API response change and must be reviewed consciously.
2. No canonical constant exists at all — a new constant should be added to
   `contract/error_codes.go` before replacing the literal.

## Category A — Canonical constant exists, value differs (API breaking)

These replacements change the `code` field in JSON error responses.  Coordinate
with API consumers before applying.

| File | Line | Current literal | Canonical constant | Canonical value |
|---|---|---|---|---|
| `x/messaging/api.go` | 96 | `"MISSING_ID"` | `contract.CodeRequired` | `"REQUIRED_FIELD_MISSING"` |
| `x/messaging/api.go` | 105 | `"NOT_FOUND"` | `contract.CodeResourceNotFound` | `"RESOURCE_NOT_FOUND"` |
| `x/messaging/api.go` | 188 | `"REQUEST_TIMEOUT"` | `contract.CodeTimeout` | `"TIMEOUT"` |

Note: `x/messaging/api.go:188` additionally uses `.Type(contract.TypeTimeout).Status(http.StatusGatewayTimeout)` which is intentional — TypeTimeout sets Category/Status, then Status(504) overrides to gateway-timeout semantics.  The Code literal is the only non-canonical part.

## Category B — No canonical constant; add new constant first

These strings are domain-specific but reusable enough to warrant a named
constant in `contract/error_codes.go`:

| File | Literal | Proposed constant | Value |
|---|---|---|---|
| `x/pubsub/distributed.go` | `"INVALID_PAYLOAD"` | `CodeInvalidPayload` | `"INVALID_PAYLOAD"` |
| `x/pubsub/distributed.go` | `"INVALID_MESSAGE"` | `CodeInvalidMessage` | `"INVALID_MESSAGE"` |
| `x/messaging/api.go` | `"PROVIDER_ERROR"` | `CodeProviderError` | `"PROVIDER_ERROR"` |
| `x/messaging/api.go` | `"QUOTA_EXCEEDED"` | `CodeQuotaExceeded` | `"QUOTA_EXCEEDED"` |
| `x/messaging/api.go` | `"TASK_EXPIRED"` | `CodeTaskExpired` | `"TASK_EXPIRED"` |
| `x/messaging/api.go` | `"SEND_ERROR"` | `CodeSendError` | `"SEND_ERROR"` |
| `x/messaging/api.go` | `"EMPTY_BATCH"` | `CodeEmptyBatch` | `"EMPTY_BATCH"` |
| `x/messaging/api.go` | `"STATS_ERROR"` | `CodeStatsError` | `"STATS_ERROR"` |
| `x/resilience/circuitbreaker/middleware.go` | `"CIRCUIT_OPEN"` | `CodeCircuitOpen` | `"CIRCUIT_OPEN"` |

## Out of Scope (domain-specific, intentionally not canonicalized)

- `x/scheduler/admin_http.go` `"TRIGGER_FAILED"` — operation result, not an API error code
- `x/websocket/server.go` `"JOIN_DENIED"` — websocket domain, not a generic error
- `x/devtools/pubsubdebug/` `"not_supported"` — internal tooling
- `x/devtools/devtools.go` `"env_reload_failed"` — internal tooling

## Tests

```bash
go test -timeout 20s ./x/messaging/... ./x/pubsub/... ./x/resilience/...
go vet ./x/messaging/... ./x/pubsub/... ./x/resilience/...
```

## Done Definition

- All Category A replacements are applied with awareness of API consumer impact.
- All Category B constants are added to `contract/error_codes.go` and the
  literals replaced.
- All tests pass; `go vet` clean.

## Outcome

Completed. Added 11 new constants to `contract/error_codes.go` and replaced all freeform literals:
- Category A (value-changing): "MISSING_ID" → CodeRequired, "NOT_FOUND" → CodeResourceNotFound, "REQUEST_TIMEOUT" → CodeTimeout
- Category B (new constants): CodeProviderError, CodeQuotaExceeded, CodeDuplicateMessage, CodeTaskExpired, CodeSendError, CodeEmptyBatch, CodeStatsError, CodeInvalidPayload, CodeInvalidMessage, CodeCircuitOpen
- Updated `x/messaging/api_test.go` to expect canonical "TIMEOUT" code for context timeout case
All tests pass; `go vet` clean.
