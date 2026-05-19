# x/validate

`x/validate` provides explicit request binding helpers for handlers that want
JSON decoding plus caller-owned validation without introducing validation
middleware or a tag language.

Use `BindJSON[T]` when a handler only needs JSON decoding. Use `Bind[T]` when a
handler has a `Validator` implementation and wants validation errors adapted to
the canonical `contract.APIError` shape.

This module is experimental. The first release includes only the standard
library JSON decoder, the `Validator` interface, `Bind`, `BindJSON`, and
`ValidationError`.

Validation adapters for third-party libraries live outside the `x/` tree.
