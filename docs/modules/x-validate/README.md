# x/validate

`x/validate` provides explicit request binding helpers for handlers that want
JSON decoding plus caller-owned validation without introducing validation
middleware or a tag language.

Use `BindJSON[T]` when a handler only needs JSON decoding. Use `Bind[T]` when a
handler has a `Validator` implementation and wants validation errors adapted to
the canonical `contract.APIError` shape.

This module is experimental. The root package includes only the standard
library JSON decoder, the `Validator` interface, `Bind`, `BindJSON`, and
`ValidationError`.

The optional `x/validate/playground` submodule adapts
`github.com/go-playground/validator/v10` to the same `Validator` interface.
It has its own `go.mod`, so the main module remains dependency-free beyond the
standard library.
