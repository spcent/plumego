# Card 0009

Priority: P2
State: active
Primary Module: x/ops
Owned Files:
  - x/ops/healthhttp/handlers.go
  - x/ops/healthhttp/runtime.go
  - reference/standard-service/internal/handler/health.go

Depends On: —

Goal:
`x/ops` 子域内存在三个相互交织的混乱点：

1. **WriteResponse vs WriteJSON 分裂**：`x/ops/ops.go` 用 `contract.WriteResponse`（带
   `{data, meta, request_id}` 信封），而 `x/ops/healthhttp/` 所有 handler 统一用
   `contract.WriteJSON`（无信封）。同一个 `x/ops` 下两种写法共存，没有任何注释说明选择依据。
   健康检查端点绕过信封是合理的（Kubernetes probe 不期望信封），但需要明确并一致执行。

2. **HealthHandler / DebugHealthHandler 命名混淆**：
   - `HealthHandler(manager, debug bool)` → debug=false 时委托给 `DetailedHandler`，
     debug=true 时额外附加 RuntimeInfo。
   - `DebugHealthHandler(manager, debug bool)` → debug=false 时直接返回 404，
     debug=true 时返回诊断 map。
   两个函数的 `debug bool` 参数语义完全不同（一个是"增加字段"，另一个是"控制是否暴露"），
   却共用相似的命名，极易混淆。

3. **reference/standard-service 手写了 HealthHandler**：参考实现
   `reference/standard-service/internal/handler/health.go` 没有使用
   `x/ops/healthhttp`，而是手写了简单的 live/ready handler，且用了 `WriteResponse`
   信封格式（和 healthhttp 的 WriteJSON 相反），不能作为标准示范。

Scope:
- **WriteResponse/WriteJSON 分层文档化**：在 `x/ops/ops.go` 文件顶部和
  `x/ops/healthhttp` package doc 中各加一行注释，说明：
  - `x/ops/healthhttp`：健康端点绕过信封，使用 `WriteJSON`（probe 兼容格式）
  - `x/ops/ops.go`：运维 API 端点遵循标准信封，使用 `WriteResponse`
- **DebugHealthHandler 改名**：将 `DebugHealthHandler(manager, debug bool)` 改名为
  `DiagnosticsHandler(manager Manager, enabled bool)`，参数名 `enabled` 更准确表达语义；
  同时将 `HealthHandler(manager, debug bool)` 的 `debug` 参数改名为 `includeRuntime bool`，
  消除歧义
- **reference/standard-service 更新**：将 `reference/standard-service` 的手写
  `HealthHandler` 替换为使用 `x/ops/healthhttp.LiveHandler()` 和
  `x/ops/healthhttp.ReadinessHandler(manager)`（若已有 manager），或明确说明
  该参考实现有意保持最简（加注释）

Non-goals:
- 不统一 WriteJSON 与 WriteResponse 为同一个函数
- 不改变任何 handler 的实际行为或响应体结构
- 不为 reference 引入 health manager 依赖（若过于复杂，仅加注释说明即可）
- 不修改 Kubernetes probe 兼容性相关逻辑

Files:
  - x/ops/healthhttp/handlers.go（注释 + HealthHandler 参数改名）
  - x/ops/healthhttp/runtime.go（DebugHealthHandler 改名为 DiagnosticsHandler）
  - x/ops/ops.go（注释说明 WriteResponse 用途）
  - reference/standard-service/internal/handler/health.go（更新或注释说明）
  - reference/standard-service/internal/app/routes.go（若 handler 签名变化则联动）

Tests:
  - go test ./x/ops/...
  - go build ./reference/standard-service/...
  - grep -rn "DebugHealthHandler" . --include="*.go" 结果应为空（或仅剩改名后引用）

Docs Sync:
  - x/ops/healthhttp 的 module.yaml doc_paths 若有 API 文档需同步

Done Definition:
- `DebugHealthHandler` 不再存在，改为 `DiagnosticsHandler`
- `HealthHandler` 的 bool 参数名为 `includeRuntime`
- `x/ops/ops.go` 与 `x/ops/healthhttp` 各有注释说明其写法选择
- `go build ./...` 和 `go test ./x/ops/...` 均通过
- reference/standard-service health handler 与 healthhttp 用法对齐，或有注释说明差异

Outcome:
