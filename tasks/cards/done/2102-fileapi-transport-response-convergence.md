# Card 2102: Converge x/fileapi Transport Responses

Priority: P1
State: done
Recipe: specs/change-recipes/http-endpoint-bugfix.yaml
Primary Module: x/fileapi
Owned Files:
- x/fileapi/handler.go
- x/fileapi/handler_test.go
- x/fileapi/context.go
- x/fileapi/module.yaml
- docs/modules/x-fileapi/README.md
Depends On: 2101

## Goal

`x/fileapi` is the app-facing file transport surface, but its JSON success path
still bypasses the canonical envelope in several handlers:

- `Upload`, `GetInfo`, `Delete`, `List`, and `GetURL` call
  `contract.WriteJSON` directly.
- `Download` correctly streams bytes and can keep direct header/status writes,
  but the JSON endpoints should use the same `contract.WriteResponse` shape as
  other app-facing handlers.
- Several error responses include raw storage or parse errors in the client
  message, e.g. `"upload failed: %v"`, `"metadata error: %v"`, and
  `"failed to generate url: %v"`.
- Query parsing silently ignores invalid `page`, `page_size`, `start_time`,
  and `end_time` values, which makes request validation incomplete.

## Scope

- Replace direct `contract.WriteJSON` calls in JSON endpoints with
  `contract.WriteResponse`.
- Keep `Download` streaming behavior explicit and unchanged, including
  `Content-Type`, `Content-Length`, `Content-Disposition`, and cache headers.
- Normalize file handler errors through small package-local helper functions
  with stable codes and non-leaking client messages.
- Treat invalid query parameters as validation errors instead of silently
  falling back when the value is malformed.
- Keep tenant identity extraction from `x/tenant/core.TenantIDFromContext`.

## Non-goals

- Do not change storage interfaces in `store/file`.
- Do not move tenant-aware storage behavior out of `x/data/file`.
- Do not add route registration helpers to `x/fileapi`.
- Do not change the streamed file body contract for `Download`.

## Files

- `x/fileapi/handler.go`
- `x/fileapi/handler_test.go`
- `x/fileapi/context.go`
- `x/fileapi/module.yaml`
- `docs/modules/x-fileapi/README.md`

## Tests

```bash
go test -race -timeout 60s ./x/fileapi/...
go test -timeout 20s ./x/fileapi/...
go vet ./x/fileapi/...
```

## Docs Sync

Update `docs/modules/x-fileapi/README.md` only if the public response shape or
validation behavior changes.  Mention that JSON endpoints use
`contract.WriteResponse`, while `Download` remains a streaming response.

## Done Definition

- No production `x/fileapi` JSON endpoint calls `contract.WriteJSON` directly.
- JSON success tests assert the canonical response envelope.
- Negative tests cover malformed pagination/time query values and storage
  errors without leaking backend error text to clients.
- Download streaming tests still pass.
- The listed validation commands pass.

## Outcome

Completed.

- Replaced JSON success responses in `Upload`, `GetInfo`, `Delete`, `List`,
  and `GetURL` with `contract.WriteResponse`.
- Kept `Download` as an explicit streaming response with direct headers and
  status writes.
- Added stable validation errors for malformed `page`, `page_size`,
  `start_time`, and `end_time` query parameters.
- Sanitized storage and metadata failure messages so backend error text is not
  returned to clients.
- Updated focused tests to assert response envelopes, invalid query errors, and
  non-leaking internal error responses.
- Updated the x/fileapi module primer to document the JSON-vs-streaming response
  policy.

Validation:

```bash
go test -race -timeout 60s ./x/fileapi/...
go test -timeout 20s ./x/fileapi/...
go vet ./x/fileapi/...
```
