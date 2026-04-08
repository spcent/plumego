# Card 0840

Priority: P1
State: active
Primary Module: contract
Owned Files:
- `contract/auth.go`
- `contract/observability_policy.go`
- `contract/module.yaml`
- `docs/modules/contract/README.md`
- `middleware/auth`
- `middleware/requestid`
- `middleware/accesslog`
- `security`
- `x/tenant`

Goal:
- Remove non-transport auth and observability policy ownership from stable `contract`.
- Converge identity, auth sentinel errors, and middleware observability policy onto their owning stable modules instead of keeping them in the transport contract layer.

Problem:
- `contract/auth.go` exports `Principal`, `Authenticator`, `Authorizer`, and auth/session error sentinels, even though these are security and session semantics rather than transport contracts.
- `contract/observability_policy.go` exports `ObservabilityPolicy`, request-id attach helpers, redaction rules, and middleware log-field construction, which is middleware observability behavior rather than a transport payload contract.
- `contract/module.yaml` and `docs/modules/contract/README.md` describe transport helpers, error writing, and request metadata, not security identity contracts or middleware logging policy.
- Current consumers in `middleware/auth`, `security/jwt`, `middleware/requestid`, `middleware/accesslog`, and `x/tenant` depend on these leaked symbols, so the stable boundary is still blurred across transport, security, and observability concerns.

Scope:
- Move auth identity contracts and auth/session sentinel errors to a stable security-owned package.
- Move middleware observability policy, field redaction, and request-id attach/read policy to the middleware observability ownership boundary.
- Keep only true transport metadata carriers in `contract` (for example request-id context accessors and trace context carrier types).
- Update `middleware/auth`, `security/jwt`, `middleware/requestid`, `middleware/accesslog`, and `x/tenant` in the same change.
- Sync contract docs and manifest to the reduced transport-only boundary.

Non-goals:
- Do not redesign JWT claim models or tenant session lifecycle in this card.
- Do not reintroduce deprecated compatibility aliases in `contract`.
- Do not move request-id context storage out of `contract` if it remains the canonical transport carrier.
- Do not add app-specific auth policies or business ACL helpers.

Files:
- `contract/auth.go`
- `contract/observability_policy.go`
- `contract/module.yaml`
- `docs/modules/contract/README.md`
- `middleware/auth`
- `middleware/requestid`
- `middleware/accesslog`
- `security`
- `x/tenant`

Tests:
- `go test -timeout 20s ./contract/... ./middleware/... ./security/... ./x/tenant/...`
- `go test -race -timeout 60s ./contract/... ./middleware/... ./security/... ./x/tenant/...`
- `go vet ./contract/... ./middleware/... ./security/... ./x/tenant/...`

Docs Sync:
- Keep the contract manifest and primer aligned on the rule that stable `contract` owns transport carriers and error/response helpers only, while auth identity belongs to `security` and middleware observability policy belongs to `middleware`.

Done Definition:
- Stable `contract` no longer exports auth identity interfaces, auth/session sentinel errors, or middleware observability policy helpers.
- Security and middleware own the symbols they actually use, with no residual references to removed `contract` exports.
- Request-id and trace carriers remain explicit and transport-focused in `contract`.
- Contract docs and manifest describe the same reduced transport-only boundary the code implements.

Outcome:
- Completed. Auth identity primitives and auth/session sentinel errors now live in `security/authn`, middleware observability policy helpers now live in `middleware/internal/observability`, and stable `contract` is back to transport carriers plus response/error helpers only.
