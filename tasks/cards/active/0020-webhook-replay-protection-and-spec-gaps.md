# Card 0020

Priority: P2
State: active
Primary Module: x/webhook
Owned Files:
  - x/webhook/inbound_hmac.go
  - x/webhook/inbound_github.go
  - specs/dependency-rules.yaml

Depends On: —

Goal:
两个相互独立但都属于"边界治理"范畴的问题：

**问题 1：x/webhook HMAC 验证不强制 nonce 检查（可能允许重放攻击）**

`inbound_hmac.go` 定义了 `NonceStore` 接口和 `MemoryNonceStore` 实现，
用于防止已验证 webhook 的重放攻击。但 nonce 检查是**可选的**：
调用方若不提供 `NonceStore`，HMAC 签名验证通过后请求即被接受，
攻击者可在签名有效窗口内重放相同请求。

`inbound_github.go` 的文档（line 13）提到 NonceStore 但只作为"可选配置"说明，
没有警告缺失时的安全风险。大部分使用者会忽略这个配置项。

修复方向：不要求强制提供 NonceStore（会破坏现有用法），但：
1. 当 `NonceStore` 未配置时，在 `Inbound.ServeHTTP` 中记录一条 WARN 日志（仅首次）
2. 在 GitHub / Stripe inbound 的 struct 文档注释中明确标注：
   "不配置 NonceStore 时，HMAC 验证通过的请求可被重放"
3. 在 `x/webhook/doc.go` 或 README 中补充安全模型说明

**问题 2：specs/dependency-rules.yaml 缺少多个扩展模块的显式条目**

当前 specs/dependency-rules.yaml 中有 x/cache、x/tenant 的显式定义，
但以下扩展家族无显式条目，其边界仅靠 catch-all deny 规则隐式约束：

- `x/fileapi`
- `x/webhook`
- `x/scheduler`（仅在 x/messaging 的 allowed_imports 中作为从属提及）
- `x/messaging`（整个家族未显式定义）
- `x/resilience`

缺少显式条目意味着：
- check 工具无法验证这些包的 import 是否合规
- 新加入的开发者无法从 specs 中知道这些包允许哪些依赖
- x/tenant 的 EXPERIMENTAL 状态与 specs 中允许 stable 层导入它的规则存在矛盾
  （x/tenant/module.yaml:4: `status: experimental`，但依赖规则未限制稳定层）

Scope:
- **webhook 警告**：
  - 在 `NewInbound` 或 `Inbound.ServeHTTP` 中，当 hmac config 存在但 NonceStore
    为 nil 时，通过 logger（若已配置）记录一次 WARN：
    "NonceStore not configured: replay attacks are possible for verified webhooks"
  - 在 `GitHubInboundConfig` 和 `StripeInboundConfig`（若有）的 NonceStore 字段注释中
    补充：`// Required for replay protection. Without it, any intercepted delivery can be replayed.`
  - 在 `x/webhook/inbound_hmac.go:41` 的 NonceStore 接口文档中加强安全说明
- **specs 补全**：
  - 在 `specs/dependency-rules.yaml` 中为 x/fileapi、x/webhook、x/messaging、
    x/resilience 各添加一个模块条目，参照 x/tenant 的格式，声明：
    - `allowed_imports`（参考各包的 module.yaml allowed_imports）
    - `status`（experimental/stable）
  - 检查并修正 x/tenant 条目：若 experimental，在 allowed_imports 中明确排除
    stable middleware 等（或在注释中说明为什么 experimental 包可以被 stable 层使用）
  - 在 specs/checks.yaml 中确认 dependency-rules 检查覆盖上述新增模块

Non-goals:
- 不将 NonceStore 改为必填项（破坏现有 API）
- 不重写 specs/dependency-rules.yaml 全部结构
- 不修改 x/webhook 的 HMAC 验证逻辑本身
- 不改变 x/tenant 的 EXPERIMENTAL 状态（由 0019 卡片决定）

Files:
  - x/webhook/inbound_hmac.go（NonceStore 接口文档 + 运行时 warn）
  - x/webhook/inbound_github.go（NonceStore 字段注释加强）
  - specs/dependency-rules.yaml（补充缺失模块条目）
  - specs/checks.yaml（确认新增条目被 lint 覆盖）

Tests:
  - go build ./x/webhook/...
  - go test ./x/webhook/...
  - go run ./internal/checks/dependency-rules（验证新规则生效）

Docs Sync:
  - x/webhook 的 module.yaml doc_paths 若有 README 同步安全说明

Done Definition:
- `NonceStore nil` 时有 WARN 日志（或明确文档说明风险）
- specs/dependency-rules.yaml 含 x/fileapi、x/webhook、x/messaging、x/resilience 条目
- `go run ./internal/checks/dependency-rules` 通过

Outcome:
