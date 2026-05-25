# Card 1513

Milestone: M-022
Recipe: specs/change-recipes/symbol-change.yaml
Context Package: implementation
Priority: P1
State: done
Primary Module: x/rpc/gateway
Owned Files:
- `x/rpc/gateway/transcoder.go`
- `x/rpc/gateway/transcoder_test.go`
- `x/rpc/gateway/module.yaml`
- `specs/dependency-rules.yaml`
Depends On: 1505

Goal:
- Remove the local `x/rpc/gateway.ErrHandlerNil` duplication and route nil
  handler construction failures through the canonical `contract.ErrHandlerNil`
  path.

Scope:
- Update the transcoder registration path to return `contract.ErrHandlerNil`.
- Update local tests to assert against the canonical contract error.
- Align the module manifest and dependency rules so `x/rpc/gateway` can depend
  on `contract` for that canonical transport error path.

Non-goals:
- Do not change any other RPC gateway error behavior.
- Do not widen `contract` with RPC-specific error types.

Files:
- `x/rpc/gateway/transcoder.go`
- `x/rpc/gateway/transcoder_test.go`
- `x/rpc/gateway/module.yaml`
- `specs/dependency-rules.yaml`

Acceptance Tests:
- `go test -timeout 20s ./x/rpc/gateway`

Tests:
- `go test -timeout 20s ./x/rpc/...`
- `go run ./internal/checks/dependency-rules`

Docs Sync:
- None expected.

Validation:
- `go test -timeout 20s ./x/rpc/...`
- `go run ./internal/checks/dependency-rules`
- `gofmt -l .`

Done Definition:
- [x] Acceptance Tests pass.
- [x] All Validation commands exit 0.
- [x] gofmt -l . produces no output.
- [x] Docs Sync targets updated (if applicable).

Outcome:
- `x/rpc/gateway.HTTPTranscoder.Register` now returns
  `contract.ErrHandlerNil` for nil handlers instead of maintaining a local
  duplicate transport-construction error.
- Updated the local RPC gateway test to assert against the canonical contract
  error, removed the stale `ErrHandlerNil` manifest entry, and aligned module
  and repo dependency rules so `x/rpc/gateway` can import `contract`.
- Validation:
  - `go test -timeout 20s ./x/rpc/...`
  - `go run ./internal/checks/dependency-rules`
  - `go run ./internal/checks/module-manifests`
  - `go run ./internal/checks/public-entrypoints-sync`
  - `gofmt -l .`
