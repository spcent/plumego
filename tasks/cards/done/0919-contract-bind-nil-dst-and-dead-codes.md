# Card 0919

Priority: P2
State: done
Primary Module: contract
Owned Files:
- `contract/context_bind.go`
- `contract/bind_helpers.go`
- `contract/error_codes.go`
Depends On: —

Goal:
- Fix two related bind-layer issues: a nil-destination footgun in `BindJSON` that
  produces a misleading "invalid JSON" error, and two dead error codes that were defined
  but never wired into `BindErrorToAPIError`.

---

## Issue A — `BindJSON(nil)` produces "invalid JSON payload" instead of a clear destination error

**File:** `contract/context_bind.go` → `decodeJSONBody`

When `dst` is nil, `json.Decoder.Decode(nil)` returns an `InvalidUnmarshalError`
(`json: Unmarshal(nil)`). This gets wrapped as:

```go
&bindError{Status: http.StatusBadRequest, Message: ErrInvalidJSON.Error(), ...}
```

The client receives `"invalid JSON payload"`. The JSON payload is fine — the problem is
that the application code passed a nil destination. The misleading message sends
developers chasing a JSON formatting problem that does not exist.

**Fix:** Guard `dst == nil` at the top of `decodeJSONBody`:

```go
func decodeJSONBody(data []byte, dst any, disallowUnknown bool) error {
    if dst == nil {
        return &bindError{Status: http.StatusBadRequest, Message: "bind destination must not be nil", Err: ErrInvalidBindDst}
    }
    ...
```

Where `ErrInvalidBindDst` is a new sentinel:

```go
// ErrInvalidBindDst is returned when a bind destination is nil or otherwise invalid.
var ErrInvalidBindDst = errors.New("invalid bind destination")
```

This surfaces the real problem immediately during development.

---

## Issue B — `CodeInvalidQuery` and `CodeInvalidBindDst` are defined but never used

**File:** `contract/error_codes.go` lines 45–46

```go
CodeInvalidQuery   = "INVALID_QUERY"
CodeInvalidBindDst = "INVALID_BIND_DESTINATION"
```

`grep -rn 'CodeInvalidQuery\|CodeInvalidBindDst' . --include='*.go'` returns only the
definition — neither constant is used in `BindErrorToAPIError` or anywhere else.

Meanwhile `bindQuery` returns two kinds of errors that SHOULD use these codes:

| Error | Message | Should Use |
|-------|---------|-----------|
| Non-pointer/nil dst | `"bind destination must be a non-nil pointer to a struct"` | `CodeInvalidBindDst` |
| Invalid query param | `"invalid query parameter %q: ..."` | `CodeInvalidQuery` |

Both fall through to the `default` case in `BindErrorToAPIError`, receiving `CodeRequestBindError`.

**Fix:** Wire the codes into `BindErrorToAPIError`:

```go
// Add alongside other sentinel checks
case errors.Is(err, ErrInvalidBindDst):
    code = CodeInvalidBindDst
    message = "invalid bind destination"
```

For query parameter errors, add a sentinel `ErrInvalidQueryParam` and wrap query parse
errors with it in `bindQuery`:

```go
var ErrInvalidQueryParam = errors.New("invalid query parameter")

// in bindQuery setFieldFromQuery failure:
return &bindError{
    Status:  http.StatusBadRequest,
    Message: fmt.Sprintf("invalid query parameter %q: %v", name, err),
    Err:     fmt.Errorf("%w: %w", ErrInvalidQueryParam, err),
}
```

Then in `BindErrorToAPIError`:
```go
case errors.Is(err, ErrInvalidQueryParam):
    code = CodeInvalidQuery
    message = "invalid query parameter"
```

---

Scope:
- Add `ErrInvalidBindDst` sentinel to `context_core.go` and guard `decodeJSONBody`.
- Add `ErrInvalidQueryParam` sentinel; wrap query parse errors.
- Wire both codes into `BindErrorToAPIError`.
- Add tests for each new branch.

Non-goals:
- Do not change `BindJSON`, `BindQuery`, or `WriteBindError` signatures.
- Do not add new bind modes or query coercion rules.

Files:
- `contract/context_core.go` (new sentinel vars)
- `contract/context_bind.go` (nil guard + sentinel wrapping)
- `contract/bind_helpers.go` (`BindErrorToAPIError` wiring)
- `contract/error_codes.go` (no change — constants already defined)
- `contract/context_bind_test.go` / `contract/write_bind_error_test.go`

Tests:
- `BindJSON` with nil dst returns `ErrInvalidBindDst`-based error, NOT `ErrInvalidJSON`.
- `BindErrorToAPIError` for nil-dst error uses `CodeInvalidBindDst`.
- `BindQuery` with invalid param value uses `CodeInvalidQuery`.
- `grep -rn 'CodeInvalidQuery\|CodeInvalidBindDst' . --include='*.go'` returns more than just the definition.
- `go test -timeout 20s ./contract/...`
- `go vet ./contract/...`

Docs Sync: —

Done Definition:
- `BindJSON(c, nil)` returns a clear "bind destination must not be nil" error.
- `CodeInvalidQuery` and `CodeInvalidBindDst` are used in `BindErrorToAPIError`.
- All tests pass.

Outcome:
