# Card 0007

Priority: P1
State: done
Primary Module: internal
Owned Files:
  - internal/httputil/html.go
  - internal/httputil/http_response.go
  - internal/httpx/client_ip.go
  - x/gateway/websocket.go
  - internal/validator/validator.go

Depends On: —

Goal:
Fix three independent but related "internal hygiene" issues in the internal utility packages:
1. **Inconsistent package name**: All files in the `internal/httputil` directory declare `package utils`, which does not match the directory name, and `utils` is an anti-pattern package name.
2. **Duplicate ClientIP logic**: `internal/httpx.ClientIP` already implements complete IP extraction logic, but `x/gateway/websocket.go:241-262` defines a private `getClientIP` with identical logic that does not reuse it.
3. **No tests for internal/httpx**: `internal/httpx/client_ip.go` has no test file, despite the function handling security-sensitive header parsing logic.
4. **internal/validator/validator.go is too large**: A single file with 2133 lines containing unrelated logic including rule definitions, struct validation, HTTP binding, and utility functions — difficult to navigate and modify locally.

Scope:
- Rename the `internal/httputil` package from `package utils` to `package httputil`,
  and update the sole caller `internal/nethttp/response_recorder.go`
- Replace `x/gateway/websocket.go:getClientIP` with a call to `internal/httpx.ClientIP`
- Add a new test file for `internal/httpx/client_ip.go` covering:
  X-Forwarded-For (including multi-IP chains), X-Real-IP, RemoteAddr, and nil request
- Split `internal/validator/validator.go` by responsibility into:
  - `validator.go` (core types: Rule, Validator, ValidationError, FieldErrors, ~300 lines)
  - `rules.go` (all built-in Rule implementations: Required, Min, Max, Email, etc.)
  - `http.go` (BindJSON, ValidateRequest, and other HTTP binding functions)
  The split must not change any public API — it is a file-level reorganization only

Non-goals:
- Do not modify the functional logic of internal/httputil
- Do not change the public interface of internal/validator
- Do not promote internal/httpx to a public package
- Do not handle calls in internal/config/validator.go (those call a local `validator` variable, not a package import)

Files:
  - internal/httputil/html.go (rename package)
  - internal/httputil/http_response.go (rename package)
  - internal/httputil/html_test.go (rename package)
  - internal/httputil/http_response_test.go (rename package)
  - internal/nethttp/response_recorder.go (update import alias or reference)
  - internal/httpx/client_ip.go
  - internal/httpx/client_ip_test.go (new file)
  - x/gateway/websocket.go (replace getClientIP, add httpx dependency)
  - internal/validator/validator.go (reduced to core types)
  - internal/validator/rules.go (new file, rule implementations moved in)
  - internal/validator/http.go (new file, HTTP binding moved in)

Tests:
  - go test ./internal/httputil/...
  - go test ./internal/httpx/...
  - go test ./internal/validator/...
  - go test ./x/gateway/...
  - go build ./internal/nethttp/...

Docs Sync: —

Done Definition:
- All files in `internal/httputil` declare `package httputil`, `go build` passes
- `x/gateway/websocket.go` no longer contains a private `getClientIP`, uses `httpx.ClientIP` instead
- `internal/httpx/client_ip_test.go` exists and `go test ./internal/httpx/...` passes
- `internal/validator/validator.go` is ≤ 400 lines, with the remaining logic distributed in new files
- All existing validator tests continue to pass without modification

Outcome:
- Renamed the package declaration in all httputil files from `package utils` to `package httputil`
- Removed the private `getClientIP` function from `x/gateway/websocket.go`; the code now calls `httpx.ClientIP`
- Added `internal/httpx/client_ip_test.go` with `TestClientIP` covering X-Forwarded-For, X-Real-IP, RemoteAddr, and nil request
- Split `validator.go` (2133 lines) into: `validator.go` (313 lines, core types), `rules.go` (1818 lines, all Rule constructors), and `http.go` (20 lines, BindJSON helpers)
