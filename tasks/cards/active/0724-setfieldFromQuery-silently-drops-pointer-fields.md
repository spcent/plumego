# Card 0724

Priority: P2

Goal:
- Fix `setFieldFromQuery` to handle pointer fields (`*string`, `*int`, etc.)
  so that query binding does not silently drop values into pointer fields.

Problem:

`context_bind.go:228-267` — `setFieldFromQuery` switches on `fv.Kind()` but
has no `reflect.Ptr` case:

```go
switch fv.Kind() {
case reflect.String:   fv.SetString(val)
case reflect.Int, ...: fv.SetInt(n)
// ...
// reflect.Ptr: no case → falls through, returns nil, value is never set
}
return nil
```

A query struct with pointer fields silently receives no value:

```go
type SearchQuery struct {
    Limit  *int    `query:"limit"`    // ← silently dropped
    Filter *string `query:"filter"`   // ← silently dropped
}
```

`c.BindQuery(&q)` returns nil (no error) but `q.Limit` and `q.Filter` remain
nil regardless of what was in the URL query string.

This is a silent data loss bug. The caller gets a success return and assumes
the struct is populated, but the pointer fields are always nil.

Fix: Add a `reflect.Ptr` case that allocates the pointed-to value, calls
`setFieldFromQuery` recursively on the element, and sets the pointer:

```go
case reflect.Ptr:
    // Allocate the pointed-to type
    elem := reflect.New(fv.Type().Elem()).Elem()
    if err := setFieldFromQuery(elem, val, vals); err != nil {
        return err
    }
    ptr := reflect.New(fv.Type().Elem())
    ptr.Elem().Set(elem)
    fv.Set(ptr)
```

Note: if `val == "" && len(vals) == 0`, the early-return at line 230 already
exits before reaching the switch, so pointer fields correctly remain nil when
the query parameter is absent.

Non-goals:
- Do not handle `**T` (double pointer) fields.
- Do not change the slice handling logic.
- Do not change `bindQuery` or the BindQuery/BindAndValidateQuery signatures.

Files:
- `contract/context_bind.go`

Tests:
- Add tests:
  - `*string` field: present query param sets the pointer to the value.
  - `*string` field: absent query param leaves pointer nil.
  - `*int` field: present query param parses and sets the pointer.
  - `*int` field: invalid value returns a BindError.
- `go test ./contract/...`
- `go vet ./...`

Done Definition:
- `setFieldFromQuery` handles `reflect.Ptr` fields.
- Present query params populate pointer fields with allocated values.
- Absent query params leave pointer fields nil.
- Invalid values return a `BindError`.
- All tests pass.
