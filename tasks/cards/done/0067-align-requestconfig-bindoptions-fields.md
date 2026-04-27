# Card 0067

Priority: P1

Goal:
- Align field names between `RequestConfig` and `BindOptions`, and remove the
  redundant `Logger` field from `BindOptions` given that `Ctx` already carries
  a `Logger`.

Problem:
- `RequestConfig.MaxBodySize` and `BindOptions.MaxBodyBytes` name the same
  concept differently, causing silent misconfiguration when developers set one
  but not the other.
- `BindOptions.Logger` duplicates `Ctx.Logger`; when binding is invoked through
  `Ctx`, callers pass in a second logger unnecessarily.

Scope:
- Rename `BindOptions.MaxBodyBytes` → `MaxBodySize` to match `RequestConfig`.
- Remove `BindOptions.Logger`; bind operations invoked via `Ctx` should fall
  back to `Ctx.Logger` automatically.
  - For the standalone `BindAndValidateJSONWithOptions(dst, opts)` path (no
    Ctx), callers that need logging should pass a logger through `opts.Validator`
    closure or a future `BindContext`; document this in a comment.
- Update all call sites.

Non-goals:
- Do not merge `RequestConfig` and `BindOptions` into one struct — they have
  different scopes (request lifecycle vs. single bind call).
- Do not change bind behaviour, only field naming and Logger wiring.

Files:
- `contract/context_bind.go`
- `contract/bind_helpers.go`
- `contract/context_core.go`
- All callers that set `BindOptions.MaxBodyBytes` or `BindOptions.Logger`

Tests:
- `go test ./contract/...`
- `go vet ./contract/...`
- `go build ./...`

Done Definition:
- `BindOptions.MaxBodyBytes` field does not exist; `MaxBodySize` is used.
- `BindOptions.Logger` field is removed.
- Bind operations through `Ctx` use `Ctx.Logger` without caller intervention.
- All tests pass.
