# Card 2242

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P2
State: done
Primary Module: contract
Owned Files:
- contract/context_bind.go
- contract/context_test.go
- contract/bind_helpers_test.go
Depends On: 2241

Goal:
Reject negative request-body size limits in binding configuration instead of silently disabling size protection.

Scope:
- Treat negative `RequestConfig.MaxBodySize` as invalid binding configuration.
- Treat negative `BindOptions.MaxBodySize` as invalid binding configuration.
- Add focused tests proving errors wrap `ErrInvalidParam`.

Non-goals:
- Do not change zero size behavior.
- Do not change body cache behavior.
- Do not change the `BindOptions` public shape.

Files:
- `contract/context_bind.go`
- `contract/context_test.go`
- `contract/bind_helpers_test.go`

Tests:
- `go test -timeout 20s ./contract/...`
- `go vet ./contract/...`

Docs Sync:
- None expected; this hardens existing size-limit behavior.

Done Definition:
- Negative body-size limits fail explicitly and do not read the body.
- Existing binding tests continue to pass.

Outcome:
- Negative request-level and per-bind max body sizes now return a `bindError` wrapping `ErrInvalidParam`.
- Invalid body-size configuration is rejected before the request body is read.
- `BindErrorToAPIError` now classifies invalid parameter binding errors as `CodeInvalidRequest` instead of body read failures.
