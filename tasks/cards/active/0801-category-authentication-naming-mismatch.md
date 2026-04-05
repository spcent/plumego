# Card 0801

Milestone: contract cleanup
Priority: P1
State: active
Primary Module: contract
Owned Files:
- `contract/errors.go`
Depends On: —

Goal:
- Eliminate the naming mismatch between the `CategoryAuthentication` constant name
  and its wire value `"auth_error"`.

Problem:
Every other `ErrorCategory` constant follows the pattern `Category{Word}` = `"{word}_error"`:
- `CategoryClient` = `"client_error"` ✓
- `CategoryServer` = `"server_error"` ✓
- `CategoryBusiness` = `"business_error"` ✓
- `CategoryTimeout` = `"timeout_error"` ✓
- `CategoryValidation` = `"validation_error"` ✓
- `CategoryRateLimit` = `"rate_limit_error"` ✓
- `CategoryAuthentication` = `"auth_error"` ✗ — Go name uses the full word
  "Authentication", wire value uses the abbreviation "auth". They do not correspond.

This forces a reader to do a mental translation every time they encounter the constant.
It also causes friction when searching the codebase: code that talks about "auth errors"
must know to use `CategoryAuthentication`, not `CategoryAuth`.

Scope:
- Rename `CategoryAuthentication` → `CategoryAuth` in `contract/errors.go`.
- Update all callers across the repo (`grep -rn 'CategoryAuthentication' . --include='*.go'`).
- Do NOT change the wire value `"auth_error"` — it is part of the public API contract.
- Do NOT change `HTTPStatusFromCategory` logic, just the constant name.

Non-goals:
- No change to wire values.
- No change to other `ErrorCategory` constants.
- No change to `ErrorType` constants.

Files:
- `contract/errors.go`
- `contract/errors_test.go`
- `contract/context_extended_test.go`
- `security/jwt/jwt.go`
- `x/websocket/server.go`
- `x/websocket/websocket.go`
- `x/ops/ops.go`
- `x/tenant/transport/transport_test.go`
- `x/tenant/session/middleware.go`
- `x/tenant/policy/middleware.go`
- `x/tenant/resolve/middleware.go`
- `x/pubsub/distributed.go`
- `middleware/auth/auth.go`
- `middleware/auth/contract.go`

Tests:
- `go build ./...`
- `go test -timeout 20s ./contract/...`
- `go test -timeout 20s ./security/... ./x/... ./middleware/...`
- `go vet ./...`
- After renaming, `grep -rn 'CategoryAuthentication' . --include='*.go'` must return empty.

Docs Sync: —

Done Definition:
- `CategoryAuthentication` does not exist anywhere in the codebase.
- `CategoryAuth` is the single constant for the `"auth_error"` category.
- All tests pass.

Outcome:
