# Card 0017

Priority: P2
State: active
Primary Module: middleware
Owned Files:
  - middleware/concurrencylimit/concurrency_limit.go
  - middleware/auth/contract.go

Depends On: —

Goal:
middleware 各子包的构造函数命名和参数风格高度不一致，调用方无法形成统一心智模型：

| 子包 | 构造函数 | 参数风格 |
|------|----------|----------|
| ratelimit | `AbuseGuard(config AbuseGuardConfig)` | Config 结构体 |
| auth | `Authenticate(authenticator, opts...)` | 接口 + 可变选项 |
| bodylimit | `BodyLimit(maxBytes int64, logger)` | 裸参数 |
| timeout | `Timeout(cfg TimeoutConfig)` | Config 结构体 |
| coalesce | `New(config) *Coalescer` + `Middleware(config)` | 双入口 |
| cors | `Middleware(opts CORSOptions)` | Options 结构体 |
| concurrencylimit | `Middleware(maxConcurrent, queueDepth, queueTimeout, logger)` | 裸参数 |

除命名混乱外，还有两个具体问题：

1. **concurrencylimit 接受 logger 但立即丢弃**（concurrency_limit.go:15: `_ = logger`）：
   接口承诺但实现为空，会误导调用方以为日志是激活的。

2. **middleware/auth/contract.go 文件名误导**：Auth 中间件的全部实现（Authenticate、
   Authorize、AuthorizeFunc、error handler、转换函数）都放在 `contract.go` 中，
   但 `contract.go` 按惯例应只含接口/类型定义，不含实现。

Scope:
- **concurrencylimit**：删除 `logger log.StructuredLogger` 参数（破坏性改动需更新调用方）；
  若确实需要日志，改为在 Config 结构体中提供 `Logger` 可选字段，
  并在队列超时事件上实际记录；执行前 grep 确认调用方数量
- **middleware/auth**：将 `contract.go` 重命名为 `auth.go`；
  若将来确需分离接口与实现，新建 `types.go` 存放 AuthErrorHandler/AuthOption 类型
- **命名规范文档化**：在 `middleware/module.yaml` 的 `agent_hints` 中补充一条：
  "无状态中间件用 `Middleware(config T) middleware.Middleware`；
   有状态中间件用 `New(config T) *Type`，另提供 `(t *Type) Middleware()` 方法；
   Config 结构体均需提供 `WithDefaults()` 或 `Default()` 函数"

Non-goals:
- 不统一所有已有子包的构造函数（改动量过大，仅处理最典型问题）
- 不改变 auth 中间件的行为逻辑
- 不强制改写 bodylimit / cors 等其他子包（等触碰时跟进）

Files:
  - middleware/concurrencylimit/concurrency_limit.go（删除 logger 参数）
  - middleware/auth/contract.go → middleware/auth/auth.go（重命名）
  - middleware/auth/contract_test.go → middleware/auth/auth_test.go（重命名）
  - middleware/module.yaml（补充 agent_hints 规范）
  - 调用 concurrencylimit.Middleware 的所有文件（grep 确认后更新）

Tests:
  - go build ./middleware/...
  - go test ./middleware/...

Docs Sync:
  - middleware/module.yaml agent_hints 补充命名规范

Done Definition:
- `grep -rn "_ = logger" middleware/concurrencylimit/` 结果为空
- `middleware/auth/contract.go` 不存在，`middleware/auth/auth.go` 存在
- `go build ./...` 通过
- middleware/module.yaml 含构造函数命名规范注释

Outcome:
