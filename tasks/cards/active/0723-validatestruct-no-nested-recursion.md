# Card 0723

Priority: P2

Goal:
- Make `validateStruct` recurse into nested and embedded struct fields so that
  validation tags on inner structs are not silently ignored.

Problem:

`validation.go:32-79` — `validateStruct` iterates over the top-level fields of
a struct. When a field's type is itself a struct (or a pointer to a struct),
the loop skips it if it has no `validate` tag, and only applies rules to the
outer field if it does have one (e.g. `validate:"required"`).

This means validation tags on inner struct fields are never reached:

```go
type Address struct {
    Street string `validate:"required"`   // ← never validated
    City   string `validate:"required"`   // ← never validated
}

type User struct {
    Name    string  `validate:"required"` // ← validated ✓
    Address Address // ← inner fields silently skipped ✗
}
```

`validateStruct(&User{Name: "Alice", Address: Address{}})` returns nil (passes)
even though `Address.Street` and `Address.City` are empty and marked required.

Same issue applies to pointer-to-struct fields:
```go
type Order struct {
    Shipping *Address `validate:"required"` // required check works
    // but if non-nil, Address.Street is NOT validated
}
```

Fix: After applying rules to a field, check if the field value is a struct or
pointer-to-struct. If so, recurse into `validateStruct`. Collect all inner
`FieldError` values and include them in the outer result, prefixing field names
with the parent field name (e.g. `"Address.Street"`).

```go
// After rule validation for field i:
if fv.Kind() == reflect.Struct {
    if err := validateStruct(fv.Addr().Interface()); err != nil {
        // collect with "FieldName." prefix
    }
} else if fv.Kind() == reflect.Ptr && !fv.IsNil() && fv.Elem().Kind() == reflect.Struct {
    if err := validateStruct(fv.Interface()); err != nil {
        // collect with "FieldName." prefix
    }
}
```

Non-goals:
- Do not validate map or slice elements in this card.
- Do not add cycle detection for recursive structs (avoid infinite loops via
  a depth limit or visited-type guard — include in implementation).
- Do not change the external `validateStruct` signature.

Files:
- `contract/validation.go`

Tests:
- Add tests for:
  - Nested struct with failing inner field returns the inner field error.
  - Pointer-to-struct field: nil pointer skips inner validation.
  - Pointer-to-struct field: non-nil pointer validates inner fields.
  - Field names are prefixed correctly (`"Address.Street"`).
- `go test ./contract/...`
- `go vet ./...`

Done Definition:
- `validateStruct` recurses into struct and pointer-to-struct fields.
- Inner `FieldError.Field` values are prefixed with the parent field name.
- A depth limit (e.g. 10) prevents infinite recursion on self-referential types.
- All tests pass.
