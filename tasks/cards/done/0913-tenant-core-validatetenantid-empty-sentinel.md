# Card 0913

Priority: P2
State: done
Primary Module: x/tenant/core
Owned Files:
- `x/tenant/core/middleware.go`
Depends On: —

Goal:
- Fix the wrong sentinel returned by `ValidateTenantID` for empty input so callers
  can distinguish "tenant not found" from "tenant ID is syntactically invalid".

---

## Problem

**File:** `x/tenant/core/middleware.go` lines 20–24

```go
func ValidateTenantID(id string) error {
    if id == "" {
        return ErrTenantNotFound  // ← wrong sentinel
    }
    ...
```

`ErrTenantNotFound` means "no tenant with this ID exists in the system". An empty
string is not a missing tenant — it is a malformed input that never identifies a tenant.
The correct sentinel is `ErrInvalidTenantID`, which the function already returns for all
other invalid inputs (line 27, 30, 34).

Returning `ErrTenantNotFound` for `""` causes callers to mishandle the error:
- A caller that checks `errors.Is(err, ErrTenantNotFound)` on an empty header will
  treat it as "tenant does not exist" and might attempt a lookup or log a spurious miss,
  when the real problem is a missing/empty header.
- A caller that checks `errors.Is(err, ErrInvalidTenantID)` to detect malformed input
  will miss the empty-string case entirely.

The package already has the right sentinel (`ErrInvalidTenantID`). The only fix needed
is to use it.

---

## Fix

```go
func ValidateTenantID(id string) error {
    if id == "" {
        return ErrInvalidTenantID  // ← consistent with all other invalid-input branches
    }
    ...
```

Callers that need to treat "empty tenant ID" as "tenant not found" can check for
`ErrInvalidTenantID` specifically or wrap accordingly in their own layer.

---

Scope:
- One-line change to `ValidateTenantID`.
- Update the test that currently asserts `ErrTenantNotFound` for empty input to assert
  `ErrInvalidTenantID` instead.

Non-goals:
- Do not change extraction logic in `FromHeader`, `FromQuery`, etc. (those already
  return `ErrTenantNotFound` when the header/param is absent, which is semantically
  correct — absence is "not found").
- Do not change `ErrInvalidTenantID` or `ErrTenantNotFound` definitions.

Files:
- `x/tenant/core/middleware.go`
- `x/tenant/core/middleware_test.go`

Tests:
- `go test -timeout 20s ./x/tenant/...`
- `go vet ./x/tenant/...`

Docs Sync: —

Done Definition:
- `ValidateTenantID("")` returns `ErrInvalidTenantID`.
- All tests pass, including the updated empty-string test assertion.

Outcome:
