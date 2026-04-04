# Card 0607

Priority: P1

Goal:
- Complete the Query binding API so it is symmetric with the JSON binding API
  for the operations that make sense for query parameters.

Problem:
Current binding surface (context_bind.go):

| Operation              | JSON                          | Query                        |
|------------------------|-------------------------------|------------------------------|
| Bind                   | `BindJSON`                    | `BindQuery`                  |
| ShouldBind alias       | `ShouldBindJSON`              | **missing**                  |
| BindAndValidate        | `BindAndValidateJSON`         | `BindAndValidateQuery`       |
| BindAndValidateOptions | `BindAndValidateJSONWithOptions` | **missing**                |

`ShouldBindJSON` is a documented alias for `BindJSON` that signals "the
caller handles errors themselves." Having only JSON support this alias makes
the Query API a dead end for callers following that pattern.

`BindAndValidateQueryWithOptions` is needed so callers can inject a custom
validator or disable struct-tag validation for query params, mirroring what
`BindAndValidateJSONWithOptions` provides.

Note: `MaxBodySize` and `DisallowUnknownFields` from `BindOptions` do not
apply to query parameters and must be silently ignored by the query variant.
Only `DisableValidation`, `Validator` (and `Redact` for logging) are relevant.

Scope:
- Add `(c *Ctx) ShouldBindQuery(dst any) error` as a documented alias for
  `BindQuery`, following the same pattern as `ShouldBindJSON`:
  ```go
  // ShouldBindQuery is an alias for BindQuery. It signals that the caller
  // is responsible for handling the returned error without writing a response.
  func (c *Ctx) ShouldBindQuery(dst any) error {
      return c.BindQuery(dst)
  }
  ```

- Add `(c *Ctx) BindAndValidateQueryWithOptions(dst any, opts BindOptions) error`
  that calls `BindQuery` then conditionally validates:
  ```go
  func (c *Ctx) BindAndValidateQueryWithOptions(dst any, opts BindOptions) error {
      if err := c.BindQuery(dst); err != nil {
          return err
      }
      if opts.DisableValidation {
          return nil
      }
      validate := opts.Validator
      if validate == nil {
          validate = validateStruct
      }
      if err := validate(dst); err != nil {
          return &BindError{Status: http.StatusBadRequest, Message: err.Error(), Err: err}
      }
      return nil
  }
  ```

Non-goals:
- Do not add `BindQueryWithOptions` (options that apply to JSON body decoding
  have no meaning for query params; per-field options belong in the query tag,
  not the bind config).
- Do not change `BindQuery` or `BindAndValidateQuery`.

Files:
- `contract/context_bind.go`

Tests:
- Add `TestShouldBindQuery` and `TestBindAndValidateQueryWithOptions` in
  `contract/context_test.go` or a new `context_bind_query_test.go`.
- `go test ./contract/...`
- `go vet ./contract/...`

Done Definition:
- `ShouldBindQuery` and `BindAndValidateQueryWithOptions` exist.
- `BindAndValidateQueryWithOptions` respects `opts.DisableValidation` and
  `opts.Validator`.
- New tests pass; existing tests pass.
