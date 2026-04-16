# Card 0010

Priority: P1
State: active
Primary Module: security
Owned Files:
  - security/jwt/jwt.go
  - security/jwt/auth_jwt.go

Depends On: —

Goal:
`security` 包存在两个独立但相关的一致性问题：

1. **JWTConfig 缺少 Validate()**：`security/abuse.Config`、`store/cache.Config`、
   `store/db.Config` 均定义了 `Validate() error` 方法并在构造函数中调用。但
   `security/jwt.JWTConfig`（jwt.go:189）无 `Validate()`，错误配置要到运行时才被发现。

2. **mapJWTError 吞掉诊断信息**：`security/jwt/auth_jwt.go:128-145` 的 `mapJWTError`
   将 JWT 内部的 4 种不同错误（ErrTokenExpired、ErrTokenNotYetValid、ErrInvalidIssuer、
   ErrInvalidAudience）全部折叠为 `authn.ErrInvalidToken`。
   调用方无法区分"token 已过期"和"issuer 不匹配"等不同失败原因，妨碍客户端给出正确提示。
   正确做法：ErrTokenExpired → `authn.ErrExpiredToken`（已存在），其余各自映射到有意义的
   sentinel 或包装成带原因的错误，同时在 `contract` 层体现对应的 ErrorType。

Scope:
- 为 `JWTConfig` 添加 `Validate() error` 方法，校验：
  - Secret/公私钥至少一项非空
  - AccessTokenTTL > 0
  - RefreshTokenTTL >= AccessTokenTTL（如适用）
- 在 `NewJWTManager` 构造函数中调用 `config.Validate()`，早失败
- 扩展 `mapJWTError`：ErrTokenNotYetValid → `authn.ErrInvalidToken`（可新增
  `authn.ErrTokenNotYetValid` sentinel 若 authn 包允许）；ErrInvalidIssuer /
  ErrInvalidAudience → 包装错误携带原因，而非丢弃信息；保持现有 `ErrTokenExpired`
  → `authn.ErrExpiredToken` 的映射不变
- 更新 `auth_jwt_test.go` 断言新映射行为

Non-goals:
- 不改变 JWTManager 的对外接口签名
- 不向 authn 包添加大量新 sentinel（仅补充确实需要区分的）
- 不修改 security/password 或 security/csrf

Files:
  - security/jwt/jwt.go（JWTConfig.Validate + NewJWTManager 调用）
  - security/jwt/auth_jwt.go（mapJWTError 扩展）
  - security/jwt/auth_jwt_test.go（断言更新）
  - security/authn/authn.go（若需要新增 sentinel）

Tests:
  - go test ./security/jwt/...
  - go test ./security/authn/...

Docs Sync: —

Done Definition:
- `JWTConfig.Validate()` 存在，`NewJWTManager` 在 config 无效时返回 error
- `mapJWTError` 不再将 ErrTokenExpired 映射为 ErrInvalidToken（已经正确，保持）
- ErrTokenNotYetValid、ErrInvalidIssuer、ErrInvalidAudience 均有明确映射（不再默认 fall-through）
- `go test ./security/...` 通过

Outcome:
