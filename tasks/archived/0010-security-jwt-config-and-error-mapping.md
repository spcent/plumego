# Card 0010

Priority: P1
State: done
Primary Module: security
Owned Files:
  - security/jwt/jwt.go
  - security/jwt/auth_jwt.go

Depends On: —

Goal:
The `security` package has two independent but related consistency problems:

1. **JWTConfig is missing Validate()**: `security/abuse.Config`, `store/cache.Config`, and
   `store/db.Config` all define a `Validate() error` method and call it in their constructors.
   However, `security/jwt.JWTConfig` (jwt.go:189) has no `Validate()`, so misconfiguration
   is only discovered at runtime.

2. **mapJWTError discards diagnostic information**: `mapJWTError` in
   `security/jwt/auth_jwt.go:128-145` collapses 4 different internal JWT errors
   (ErrTokenExpired, ErrTokenNotYetValid, ErrInvalidIssuer, ErrInvalidAudience) all into
   `authn.ErrInvalidToken`. Callers cannot distinguish between "token has expired" and
   "issuer mismatch" and other failure reasons, which prevents clients from giving correct feedback.
   The correct approach: ErrTokenExpired → `authn.ErrExpiredToken` (already exists); the others
   should each map to a meaningful sentinel or be wrapped as errors carrying a reason, with the
   corresponding ErrorType reflected at the `contract` layer.

Scope:
- Add a `Validate() error` method to `JWTConfig` that checks:
  - At least one of Secret / public-private key pair is non-empty
  - AccessTokenTTL > 0
  - RefreshTokenTTL >= AccessTokenTTL (where applicable)
- Call `config.Validate()` in the `NewJWTManager` constructor for early failure
- Extend `mapJWTError`: ErrTokenNotYetValid → `authn.ErrInvalidToken` (or add
  `authn.ErrTokenNotYetValid` sentinel if the authn package allows); ErrInvalidIssuer /
  ErrInvalidAudience → wrap the error carrying the reason rather than discarding the information;
  keep the existing `ErrTokenExpired` → `authn.ErrExpiredToken` mapping unchanged
- Update `auth_jwt_test.go` to assert the new mapping behavior

Non-goals:
- Do not change the external interface signature of JWTManager
- Do not add a large number of new sentinels to the authn package (only add ones that genuinely need to be distinguished)
- Do not modify security/password or security/csrf

Files:
  - security/jwt/jwt.go (JWTConfig.Validate + NewJWTManager call)
  - security/jwt/auth_jwt.go (mapJWTError extension)
  - security/jwt/auth_jwt_test.go (assertion updates)
  - security/authn/authn.go (if new sentinels are needed)

Tests:
  - go test ./security/jwt/...
  - go test ./security/authn/...

Docs Sync: —

Done Definition:
- `JWTConfig.Validate()` exists; `NewJWTManager` returns an error when the config is invalid
- `mapJWTError` no longer maps ErrTokenExpired to ErrInvalidToken (already correct, kept as-is)
- ErrTokenNotYetValid, ErrInvalidIssuer, and ErrInvalidAudience all have explicit mappings (no more default fall-through)
- `go test ./security/...` passes

Outcome:
- Added `JWTConfig.Validate()` checking AccessExpiration > 0, RefreshExpiration >= AccessExpiration, and a supported Algorithm
- Called `config.Validate()` at the start of `NewJWTManager` for early failure on bad configuration
- Fixed the default arm of `mapJWTError`: unknown errors are now returned as-is rather than being remapped to ErrInvalidToken
- ErrInvalidIssuer, ErrInvalidAudience, ErrUnknownKey, and ErrMissingSubject now use `fmt.Errorf("reason: %w", authn.ErrInvalidToken)` to preserve diagnostic information
