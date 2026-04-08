# Card 0847

Priority: P1
State: active
Primary Module: core
Owned Files:
- `core/app_helpers.go`
- `core/http_handler.go`
Depends On:

Goal:
- Remove the dead unexported `logError` method and fix the non-canonical error code in `ServeHTTP` so that every HTTP error written by the stable kernel follows `contract` code conventions.

Problem:

**Dead method:**
- `App.logError` (`app_helpers.go:72`) is an unexported method that takes a message, error, and log fields. It has never been called anywhere in the package or in the repository. Unexported methods cannot be used outside `core`, making this unreachable dead code that misleads maintainers into thinking it serves some lifecycle role.

**Non-canonical error code:**
- `App.ServeHTTP` (`http_handler.go:97`) writes a 503 response when the handler is nil:
  ```go
  contract.NewErrorBuilder().Status(http.StatusServiceUnavailable).
      Code("handler_not_configured").Message("handler not configured").Build()
  ```
- The inline string `"handler_not_configured"` is lowercase snake_case, inconsistent with every other `contract` error code constant which uses SCREAMING_SNAKE_CASE (e.g. `CodeUnavailable = "SERVICE_UNAVAILABLE"`).
- The correct code for a 503 is `contract.CodeUnavailable`. If a more specific code is needed, it belongs in `contract/error_codes.go` as a named constant, not as an inline string literal in a different package.

Scope:
- Delete `App.logError` from `core/app_helpers.go`.
- Replace the inline `Code("handler_not_configured")` string in `core/http_handler.go:97` with `contract.CodeUnavailable`.

Non-goals:
- Do not add a new error code constant; `CodeUnavailable` is sufficient for this case.
- Do not change `wrapCoreError` (it wraps internal Go errors, not HTTP error codes).
- Do not touch any other lifecycle path.

Files:
- `core/app_helpers.go`
- `core/http_handler.go`

Tests:
- `go test -timeout 20s ./core/...`
- `go vet ./core/...`

Docs Sync:
- None required.

Done Definition:
- `App.logError` no longer exists.
- `App.ServeHTTP` uses `contract.CodeUnavailable` instead of an inline lowercase string.
- All tests pass.

Outcome:
- Pending.
