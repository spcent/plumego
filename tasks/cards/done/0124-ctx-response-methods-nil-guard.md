# Card 0124

Priority: P2
State: done
Primary Module: contract
Owned Files: contract/context_response.go

Goal:
- Add nil-receiver guards to `Ctx.JSON`, `Ctx.Text`, `Ctx.Bytes`,
  `Ctx.Redirect`, and `Ctx.ErrorJSON` so that calling any response method
  on a nil `*Ctx` returns an error rather than panicking.

Problem:

`context_response.go` currently mixes guarded and unguarded methods:

| Method        | nil check? | panics on nil? |
|---------------|-----------|----------------|
| `Response`    | yes (line 35) | no |
| `JSON`        | no        | yes (`c.W.Header()`) |
| `Text`        | no        | yes (`c.W.Header()`) |
| `Bytes`       | no        | yes (`c.W.Header()`) |
| `Redirect`    | no        | yes (`c.W`) |
| `ErrorJSON`   | no        | yes (`c.TraceID`, `c.W`) |

`Ctx.Close` (`context_core.go:294`) and `Ctx.CollectedErrors` (line 338)
also guard nil. The inconsistency means callers cannot rely on the contract
"any Ctx method is safe on a nil receiver" — they must know which methods
are safe. A nil `*Ctx` is a plausible programming error in middleware chains,
so panic is a worse outcome than a returned sentinel.

Fix — add to each unguarded method:
```go
if c == nil {
    return ErrContextNil
}
```

`ErrContextNil` is already defined at `context_core.go:115`.

Methods to update in `context_response.go`:

```go
func (c *Ctx) JSON(status int, data any) error {
    if c == nil {
        return ErrContextNil
    }
    ...
}

func (c *Ctx) Text(status int, text string) error {
    if c == nil {
        return ErrContextNil
    }
    ...
}

func (c *Ctx) Bytes(status int, data []byte) error {
    if c == nil {
        return ErrContextNil
    }
    ...
}

func (c *Ctx) Redirect(status int, location string) error {
    if c == nil {
        return ErrContextNil
    }
    ...
}

func (c *Ctx) ErrorJSON(status int, errCode string, message string, details map[string]any) error {
    if c == nil {
        return ErrContextNil
    }
    ...
}
```

Non-goals:
- Do not add nil guards to `SetCookie`, `Cookie`, `File`, `SafeRedirect` in
  this card (those follow the same pattern but can be addressed separately if
  desired; keep this card small).
- Do not change any existing logic beyond adding the guard.

Files:
- `contract/context_response.go`

Tests:
- Add a sub-table test in `contract/context_test.go` or
  `contract/context_extended_test.go`:
  ```go
  func TestCtxResponseMethodsNilSafe(t *testing.T) {
      var c *Ctx
      if err := c.JSON(200, nil); !errors.Is(err, ErrContextNil) {
          t.Errorf("JSON: want ErrContextNil, got %v", err)
      }
      if err := c.Text(200, ""); !errors.Is(err, ErrContextNil) {
          t.Errorf("Text: want ErrContextNil, got %v", err)
      }
      if err := c.Bytes(200, nil); !errors.Is(err, ErrContextNil) {
          t.Errorf("Bytes: want ErrContextNil, got %v", err)
      }
      if err := c.Redirect(302, "/"); !errors.Is(err, ErrContextNil) {
          t.Errorf("Redirect: want ErrContextNil, got %v", err)
      }
      if err := c.ErrorJSON(400, "CODE", "msg", nil); !errors.Is(err, ErrContextNil) {
          t.Errorf("ErrorJSON: want ErrContextNil, got %v", err)
      }
  }
  ```
- `go test ./contract/...`
- `go vet ./contract/...`

Done Definition:
- `JSON`, `Text`, `Bytes`, `Redirect`, `ErrorJSON` all return `ErrContextNil`
  when the receiver is nil.
- No panics on nil `*Ctx` for any of the above methods.
- All existing tests pass.

Outcome:
- Completed by adding nil-receiver guards to `JSON`, `Text`, `Bytes`,
  `Redirect`, and `ErrorJSON`, plus a table-driven regression test covering all
  five methods.

Validation Run:
- `gofmt -w contract/context_response.go contract/context_response_nil_test.go`
- `go test -timeout 20s ./contract/...`
- `go vet ./contract/...`
