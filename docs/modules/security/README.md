# security

## Purpose

`security` contains reviewable authentication, header, input-safety, and abuse-guard primitives.

## v1 Status

- `GA` in the Plumego v1 support matrix
- Public compatibility is expected for the stable package surface

## Use this module when

- verifying tokens or signatures
- enforcing security headers
- validating hostile or malformed input

## Do not use this module for

- app bootstrap
- business-specific authorization policy hidden in middleware
- logging secrets or tokens

## First files to read

- `security/module.yaml`
- the target package under `security/*`
- `AGENTS.md` security rules

## Canonical change shape

- fail closed
- keep verification explicit
- add negative tests for invalid input and invalid credentials
- keep `Principal`, `Authenticator`, `Authorizer`, and the canonical `WithPrincipal(...)` / `PrincipalFromContext(...)` accessors in `security/authn`
- parse bearer credentials with an exact `Bearer` scheme and whitespace delimiter; query-string tokens remain ignored
- keep JWT, header, and signature logic in `security/*` as primitives and policies
- keep session revocation, token-version invalidation, and tenant-session sentinel errors in `x/tenant/session`, not in stable `security/*`
- keep reusable resilience primitives such as circuit breakers in `x/resilience`, not in stable `security/*`
- route HTTP adapter wiring through `middleware/auth` and `middleware/security`

## Production Profile Relationship

Use `security/*` for reviewable primitives and policies:

- v1 compatibility paths are retained only where explicitly documented here and
  in `specs/deprecation-inventory.yaml`; prefer the strict or explicit
  alternatives in new production wiring.
- `security/authn` owns principals, authenticators, authorizers, and context accessors.
- `security/authn` context helpers defensively copy mutable principal fields.
- `security/authn.StaticToken` compares fixed credentials through fixed-length digest comparison.
- `security/headers` owns header policies consumed by `middleware/security`.
- `security/headers.Policy.Validate` reports unsafe configured header names and values before runtime; `Policy.ApplyChecked` validates before writing any headers, while compatibility `Policy.Apply` skips unsafe values and unsupported standard header values.
- Production-facing examples should prefer `middleware/security.Middleware(security.Config{...})` or direct `Policy.ApplyChecked`; use `Policy.Apply` only when lenient skip behavior is explicitly desired.
- `middleware/security.Middleware` validates custom policies during construction instead of applying a partially invalid policy at runtime.
- `security/headers` treats requests as HTTPS only for direct TLS, an all-HTTPS `X-Forwarded-Proto` chain, or an all-HTTPS RFC `Forwarded` proto chain; `X-Forwarded-Ssl` alone is ignored.
- `security/headers.CSPBuilder` filters unsafe directive values so caller-provided semicolons or controls cannot create extra directives; production configuration should use `BuildChecked` or `Validate` to fail closed when a source was rejected.
- `security/input` owns input-safety checks and rejects unsafe HTTP header names or values before they reach transport adapters.
- `security/input.ValidateURL` rejects embedded userinfo credentials, non-decimal or out-of-range URL ports, and unsafe relative URL forms such as network-path, raw-control, or raw-backslash paths; it is a URL shape and scheme check, not an SSRF boundary.
- `security/input.ValidatePublicURL` is the SSRF-sensitive helper for absolute HTTP(S) fetch targets with literal localhost, private, link-local, metadata, multicast, and unspecified IP targets rejected.
- `security/input.ValidateEmail` applies DNS-style domain label checks.
- `security/input` sanitizer helpers are lossy best-effort defense-in-depth utilities, not replacements for context-aware sanitizers, parameterized queries, or output encoding.
- `security/input.BestEffortSanitizeHTML` covers script blocks and quoted or unquoted inline event handlers as a basic defense-in-depth helper; `SanitizeHTML` is a legacy compatibility alias, and new code should prefer the `BestEffort*` name.
- `security/input.BestEffortSanitizeSQL` removes line comments, semicolons, common SQL keywords, and single-line or multiline block comments as a defense-in-depth helper only; `SanitizeSQL` is a legacy compatibility alias, and new code should prefer the `BestEffort*` name.
- `security/abuse` owns abuse guard decisions consumed by `middleware/ratelimit`.
- `security/abuse.Config.Validate` and `NewLimiterWithConfig` treat zero fields as omitted defaults and reject explicit negative values; `NewLimiterWithConfig` remains the canonical production constructor, while `NewLimiter` is the lenient compatibility constructor and falls back to defaults on invalid explicit values.
- `security/abuse.AllowKey` is the fail-closed limiter check for callers that require a non-empty subject/IP key; `Allow` remains the compatibility path and maps empty keys to the shared `unknown` bucket.
- `security/abuse` reports limiter bucket metrics from the same accounting path used for eviction and cleanup decisions.
- `security/jwt` and `security/password` own token and password primitives.
- `security/jwt` verification fails closed when configured issuer, configured audience, or subject claims are missing or mismatched.
- `security/jwt` rejects malformed or oversized compact token envelopes before payload decoding.
- `security/jwt` verification requires a valid JWT header type, matching header/payload key IDs, and valid persisted signing key material.
- `security/jwt.JWTManager` accepts a minimal key-store interface instead of binding to a concrete stable store implementation.
- `security/jwt.JWTManager` treats an absent active signing-key marker as recoverable, but fails startup when a persisted active marker cannot be read.
- `security/jwt.JWTManager.RotateKey` returns signing-key metadata only; HMAC secrets and Ed25519 private keys remain internal.
- `security/jwt.JWTConfig.Validate` rejects negative rotation intervals and clock skew values.
- `security/jwt` verification is read-only with respect to signing key rotation; key rotation happens while issuing tokens or through explicit rotation.
- `security/jwt` verification rejects tokens that omit issued-at, not-before, or expiration claims; issued-at values in the future beyond configured clock skew are invalid.
- `security/jwt` generation and verification honor canceled caller contexts before expensive work.
- `security/jwt` generation, verification, and key rotation evaluate time through a single manager clock path so boundary behavior stays consistent.
- `security/jwt` context and principal helpers defensively copy mutable role and permission slices.
- `security/jwt` authorizer helpers fail closed for empty policies or empty permission targets unless callers set the explicit `AllowEmpty` compatibility flag.
- `security/password` exposes sentinel errors for invalid cost, invalid stored hash, and password mismatch so callers can classify failures with `errors.Is`.
- `security/password` enforces `MinimumCost` and `MaximumCost` bounds for generated and stored PBKDF2 hashes, rejects plaintext password inputs over `MaxPasswordLength` with `ErrPasswordTooLong`, and validates stored salt/hash lengths before verification.

HTTP request wiring belongs in `middleware/auth`, `middleware/security`, and
`middleware/ratelimit`. Application-specific authorization decisions should be
constructor-injected into those adapters rather than hidden behind package
globals.
