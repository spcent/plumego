# Card 0020

Priority: P2
State: done
Primary Module: x/webhook
Owned Files:
  - x/webhook/inbound_hmac.go
  - x/webhook/inbound_github.go
  - specs/dependency-rules.yaml

Depends On: —

Goal:
Two independent problems both belonging to the "boundary governance" category:

**Problem 1: x/webhook HMAC verification does not enforce nonce check (potential replay attack)**

`inbound_hmac.go` defines a `NonceStore` interface and `MemoryNonceStore` implementation to
prevent replay attacks on verified webhooks. However, the nonce check is **optional**: if the
caller does not provide a `NonceStore`, requests are accepted after HMAC signature verification
passes — an attacker can replay the same request within the signature's validity window.

The documentation in `inbound_github.go` (line 13) mentioned NonceStore only as "optional
configuration" without warning about the security risk of omitting it. Most users would skip
this config entirely.

Fix direction: do not require a NonceStore (would break existing usage), but:
1. When `NonceStore` is not configured, log a single WARN in `Inbound.ServeHTTP` (first
   occurrence only)
2. Explicitly annotate the NonceStore field in GitHub / Stripe inbound struct doc comments:
   "Without NonceStore, HMAC-verified requests can be replayed"
3. Strengthen security model explanation in `x/webhook/doc.go` or README

**Problem 2: specs/dependency-rules.yaml missing explicit entries for several extension modules**

The current specs/dependency-rules.yaml had explicit definitions for x/cache and x/tenant,
but the following extension families had no explicit entries — their boundaries were only
implicitly constrained by a catch-all deny rule:

- `x/fileapi`
- `x/webhook`
- `x/scheduler` (mentioned only as a subordinate in x/messaging allowed_imports)
- `x/messaging` (entire family undefined)
- `x/resilience`

Missing explicit entries meant:
- The check tool could not verify whether imports into these packages were compliant
- New developers had no way to discover allowed dependencies from specs alone
- x/tenant's EXPERIMENTAL status conflicted with rules that allowed stable layers to import
  it (x/tenant/module.yaml:4 `status: experimental`, but dependency rules placed no restriction
  on stable-layer imports)

Scope:
- **webhook warning**:
  - In `NewInbound` or `Inbound.ServeHTTP`, when an HMAC config exists but NonceStore is nil,
    log a one-time WARN via the logger (if configured):
    "NonceStore not configured: replay attacks are possible for verified webhooks"
  - Add to `GitHubInboundConfig` and `StripeInboundConfig` (if present) NonceStore field
    comments: `// Required for replay protection. Without it, any intercepted delivery can be replayed.`
  - Strengthen security language in the NonceStore interface doc at
    `x/webhook/inbound_hmac.go:41`
- **specs completion**:
  - Add module entries for x/fileapi, x/webhook, x/messaging, and x/resilience in
    `specs/dependency-rules.yaml`, following the x/tenant format, declaring:
    - `allowed_imports` (referencing each package's module.yaml allowed_imports)
    - `status` (experimental/stable)
  - Review and correct the x/tenant entry: if experimental, explicitly exclude stable
    middleware in allowed_imports (or add a comment explaining why experimental packages
    may be imported by stable layers)
  - Confirm in specs/checks.yaml that the dependency-rules check covers the newly added modules

Non-goals:
- Do not make NonceStore a required field (would break existing API)
- Do not rewrite the full structure of specs/dependency-rules.yaml
- Do not modify x/webhook's HMAC verification logic itself
- Do not change x/tenant's EXPERIMENTAL status (decided by card 0019)

Files:
  - x/webhook/inbound_hmac.go (NonceStore interface doc + runtime warn)
  - x/webhook/inbound_github.go (NonceStore field comment strengthened)
  - specs/dependency-rules.yaml (missing module entries added)
  - specs/checks.yaml (confirm new entries are covered by lint)

Tests:
  - go build ./x/webhook/...
  - go test ./x/webhook/...
  - go run ./internal/checks/dependency-rules (verify new rules take effect)

Docs Sync:
  - x/webhook module.yaml doc_paths: sync security explanation to README if present

Done Definition:
- `NonceStore nil` triggers a WARN log (or explicit documentation of the risk)
- specs/dependency-rules.yaml contains entries for x/fileapi, x/webhook, x/messaging,
  x/resilience
- `go run ./internal/checks/dependency-rules` passes

Outcome:
- Added `replayWarnOnce sync.Once` to `Inbound` struct; `webhookInGitHub` and
  `webhookInStripe` log a one-time WARN on first request:
  "NonceStore not configured: replay attacks are possible for verified webhooks"
- Strengthened `NonceStore` interface and `HMACReplayConfig.NonceStore` doc comments with
  explicit replay-attack risk language
- Enhanced `GitHubVerifyOptions.NonceStore` comment
- Added explicit dependency-rules.yaml entries for x/fileapi, x/webhook, x/messaging,
  x/resilience with allow/deny lists
- Annotated x/tenant entry explaining global rule prevents stable→x/tenant imports
