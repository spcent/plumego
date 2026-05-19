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

Third-party validation adapters are intentionally not shipped under `x/*`.
Applications that want `github.com/go-playground/validator/v10` or another
validator should keep the adapter in their own module, as shown by
`reference/with-rest`, or publish it as an external module.
