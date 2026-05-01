# Card 0749

Milestone: v1
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P0
State: active
Primary Module: security
Owned Files:
- `security/`
- `middleware/`
- `contract/`
- `x/webhook/`
- `docs/release/v1.0.0-rc.1.md`
Depends On: 0747

Goal:
- Complete the v1 fail-closed security audit for stable security primitives and security-adjacent extension entrypoints.

Problem:
Security regressions are v1 blockers even when they appear in optional adapters. The release needs evidence that auth, verification, policy, and error-writing paths fail closed and do not expose secrets.

Scope:
- Audit stable `security`, `middleware`, and `contract` negative paths.
- Verify token, signature, and policy checks fail closed.
- Verify secret comparisons use timing-safe comparison where secrets are compared.
- Verify error responses and logs do not leak tokens, signatures, or private keys.
- Run focused tests for `x/webhook`, `x/websocket`, and `x/gateway` where security-adjacent behavior is exposed, while keeping those modules experimental.

Non-goals:
- Do not promote any security-adjacent `x/*` package.
- Do not redesign auth or policy APIs.
- Do not add new dependencies.

Files:
- `security/`
- `middleware/`
- `contract/`
- `x/webhook/`
- `docs/release/v1.0.0-rc.1.md`

Tests:
- `go test ./security ./middleware ./contract`
- `go test ./x/webhook ./x/websocket ./x/gateway`
- `go vet ./...`

Docs Sync:
- Required if behavior, failure semantics, or release blockers change.

Done Definition:
- All audited auth, signature, and policy negative paths fail closed.
- No audited error or log path leaks secret material.
- Security-adjacent extension evidence is recorded without changing experimental maturity.
- Any discovered runtime bug is fixed or split into a P0 blocker card.
