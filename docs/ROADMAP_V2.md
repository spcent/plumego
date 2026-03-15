# Plumego v1.x → v2.0 路线图

> **文档版本**: v2.0 草案 | **基准日期**: 2026-03-15 | **基于代码**: origin/main (8356160)
>
> 本文档取代 `ROADMAP_V1.md`，基于当前真实代码状态进行全面重写。

---

## 一、ROADMAP_V1.md 合理性分析

### 1.1 背景与历史偏差

`ROADMAP_V1.md` 写于项目早期，彼时 plumego 被定位为"轻量 Go HTTP 工具包"。随着迭代推进，项目实际演进路径与原路线图存在**重大偏离**。

| 原路线图阶段 | 规划内容 | 实际现状 | 评估 |
|---|---|---|---|
| Phase 1：核心再造 | 重塑 core.App、统一 plumego.Run、标准化 Component | Component 接口已实现（RegisterRoutes/RegisterMiddleware/Start/Stop/Health），BaseComponent 提供默认实现；但 `plumego.Run` 统一入口**未实现**，config 包仍有全局状态 | 60% 完成，存在遗留问题 |
| Phase 2：路由与中间件 | 显式路由、标准中间件签名、精简 Context | 均已完成：trie 路由、`func(http.Handler) http.Handler`、`contract.Ctx` 轻量包装 | **100% 完成** |
| Phase 3：AI 网关与工具链 | ai.Provider 抽象、ai.Tool 框架、SSE 支持 | **严重低估**：实际已有 provider/tool/orchestration/distributed/semanticcache/multimodal/marketplace/session/sse/streaming/filter/prompt/instrumentation/llmcache/ratelimit/resilience/circuitbreaker/metrics 共 18+ 子包 | 规划远落后于实现 |
| Phase 4：生产力与可观测性 | 统一 observability、security 中间件、性能基准 | observability 部分完成（Prometheus/OTel 适配器存在），security 包有 jwt/headers/input/abuse guard 但**缺 CSRF**；benchmark 存在于 router 但未纳入 CI | 50% 完成 |
| Phase 5：文档与发布 | AGENTS.md、风格指南、examples/v1、质量门禁、v1.0.0 标签 | AGENTS.md 已完善，CANONICAL_STYLE_GUIDE.md 存在，examples 有 15+ 示例项目（agents/api-gateway/ai-agent-gateway/multi-tenant-saas 等），当前版本为 v1.0.0-rc.1 | 80% 完成 |

### 1.2 核心问题：定位已经演变

原路线图将 plumego 定位为**轻量 HTTP 工具包**。而当前代码实质上已是一个**AI-Native SaaS 基础设施框架**：

```
原定位: Go web toolkit ≈ 路由 + 中间件 + 标准库封装
当前实质: AI-First SaaS platform = HTTP层 + AI网关层 + 多租户层 + 数据层
```

AI 子系统规模已远超 HTTP 层：
- AI 层：18+ 子包，含分布式执行引擎、语义缓存、多模态、工具市场等
- 数据层：DB sharding、读写分离、Redis 集群、文件存储（本地/S3）
- 租户层：配额管理、策略引擎、隔离 DB 查询

### 1.3 原路线图的合理之处

1. **设计原则正确**：AI-First、显式优于隐式、拥抱标准库、组件化、零全局状态——这五条原则仍然完全适用
2. **模块边界清晰**：contract/router/middleware/security/tenant/store/ai 的职责划分准确
3. **质量门禁方向**：go test -race、go vet、gofmt 是正确选择

### 1.4 原路线图的主要问题

1. **AI 层规划严重不足**：仅规划了 Provider 抽象和 SSE，但实际已建成庞大 AI 基础设施，且多个子模块处于并行开发（distributed/SPRINT5_PLAN.md、semanticcache/SPRINT2_ENHANCEMENTS.md）
2. **缺乏子模块稳定性分级**：18 个 AI 子包参差不齐，无稳定等级标注，外部使用者无法判断哪些 API 已稳定
3. **遗漏关键基础设施**：MCP 协议支持、Agent 框架、数据库迁移 runner 均未规划
4. **发布节奏不清晰**：v1.0.0-rc.1 但无明确 GA 标准定义
5. **测试覆盖不均衡**：核心包测试充分，AI 子包覆盖差异大

---

## 二、Plumego 的准确定位

基于当前代码状态，plumego 的准确定位应为：

> **Plumego 是一个以标准库为基础的 AI-Native Go Web 框架，专为构建多租户 AI 应用和 AI 网关而设计。它提供从 HTTP 路由到 AI 编排的全栈基础设施，同时保持与 net/http 生态的完全兼容。**

### 核心竞争优势

| 维度 | 差异化价值 |
|---|---|
| **AI-First 设计** | AI 提供商统一接口、工具框架、语义缓存、多模态、分布式编排 |
| **零外部依赖（核心）** | main module 仅依赖标准库，AI 层按需引入 |
| **多租户原生** | 配额/策略/速率限制/数据隔离开箱即用 |
| **生产级数据层** | DB sharding、读写分离、分布式缓存 |
| **标准库兼容** | 任何 net/http 组件均可接入 |

---

## 三、现状差距分析（Gap Analysis）

### 3.1 核心层（High Stability）

| 问题 | 严重程度 | 说明 |
|---|---|---|
| `config` 全局状态未清除 | 中 | `config/global.go` 仍持有包级全局实例，违反零全局状态原则 |
| `plumego.Run` 统一入口缺失 | 低 | README 显示 `core.New()` + `app.Boot()`，无顶层 `plumego.Run` 封装 |
| Component 接口 Start/Stop 未在 BaseComponent 中提供默认实现 | 低 | 需要实现者必须手写 Start/Stop，增加样板代码 |

### 3.2 AI 层（Medium Stability，最大差距）

| 子包 | 稳定性 | 核心问题 |
|---|---|---|
| `ai/provider` | 中 | Claude/OpenAI 实现存在，但无完整的错误规范化、无重试策略文档 |
| `ai/tool` | 中 | FuncTool 存在，但缺乏**声明式 Go 结构体 → Tool Schema** 自动生成 |
| `ai/orchestration` | 低 | 单测覆盖不足，API 未冻结 |
| `ai/distributed` | 低 | SPRINT5_PLAN.md 表明仍在开发中；持久化层、恢复机制需验证 |
| `ai/semanticcache` | 低 | SPRINT2_ENHANCEMENTS.md 表明仍在迭代；向量存储后端仅有内存实现 |
| `ai/marketplace` | 低 | 实验性，安装/验证流程未生产验证 |
| `ai/multimodal` | 低 | Provider 适配仅部分完成 |
| **MCP 协议** | 缺失 | Model Context Protocol 支持完全缺失，这是 AI Agent 生态关键协议 |
| **Agent 框架** | 缺失 | 无统一 `ai.Agent` 类型和 Agentic Loop 抽象 |
| **流式代理转发** | 缺失 | Provider → SSE → Client 的完整流式透传管道无封装 |

### 3.3 安全层（Critical）

| 问题 | 严重程度 |
|---|---|
| CSRF 防护缺失（原路线图 Phase 4.2 提及但未实现） | 高 |
| JWT 续签（Refresh Token 轮换）完整流程未在 security/jwt 中封装 | 中 |
| AI 层输入内容安全（Prompt Injection 防护）缺失 | 高 |

### 3.4 可观测性层

| 问题 | 严重程度 |
|---|---|
| `observability` 统一组件未完成（分散在 core/middleware/metrics）| 中 |
| AI 请求追踪（provider → orchestration → tool call）链路不完整 | 高 |
| 性能基准未纳入 CI | 低 |

### 3.5 数据层

| 问题 | 严重程度 |
|---|---|
| DB 迁移 runner 缺失（仅有 SQL 文件，无 Runner） | 中 |
| `store/file` S3 实现依赖外部 SDK（与 stdlib-only 原则冲突） | 中 |

### 3.6 文档与 DX

| 问题 | 严重程度 |
|---|---|
| AI 子包无统一入门文档（18 个子包，无导航） | 高 |
| examples/ 质量参差不齐（部分仅有框架无完整运行逻辑） | 中 |
| `plumego dev` CLI 文档不完整 | 低 |
| CHANGELOG.md 缺失 | 低 |

---

## 四、v1.0.0 GA 发布计划（短期：0-6 周）

**目标**：完成当前 rc.1 → GA 的最后冲刺，确保核心层 API 冻结并通过质量门禁。

### M1：质量收敛（第 1-2 周）

- [ ] **M1.1 全包质量门禁通过**
  - `go test -race -timeout 60s ./...` 零失败
  - `go vet ./...` 零警告
  - `gofmt -w .` 零差异
  - 输出覆盖率报告，目标：核心包（contract/router/middleware/security）≥ 85%

- [ ] **M1.2 核心 API 冻结声明**
  - 在 `contract/`、`router/`、`middleware/`、`security/` 包内标注 `// Stable API` 注释
  - 在 AGENTS.md 中更新稳定性等级表

- [ ] **M1.3 config 全局状态清理**
  - 将 `config/global.go` 中的包级实例改为显式初始化
  - 提供迁移指南（BREAKING CHANGE 注释）

### M2：安全加固（第 2-3 周）

- [ ] **M2.1 CSRF 中间件**
  - 实现 `security/csrf` 包，提供 Double Submit Cookie 模式
  - 签名：`func CSRFProtection(opts CSRFOptions) func(http.Handler) http.Handler`
  - 测试矩阵：`csrf_negative_matrix_test.go`

- [ ] **M2.2 Prompt Injection 防护（基础版）**
  - 在 `ai/filter` 中添加 `PromptInjectionFilter`
  - 基于关键词规则和结构检测，不引入外部 ML 模型
  - 集成到 `ai/provider` 请求管道

- [ ] **M2.3 JWT 续签流程封装**
  - 在 `security/jwt` 中提供 `RefreshTokenManager` 完整实现
  - 确保与 `contract/auth.go` 的 `RefreshManager` 接口对齐

### M3：文档与示例（第 3-5 周）

- [ ] **M3.1 CHANGELOG.md 创建**
  - 从 git log 整理 v1.0.0-rc.1 的主要变更

- [ ] **M3.2 AI 子包导航文档**
  - 新建 `docs/ai/README.md`，建立 18 个 AI 子包的稳定性分级表和使用路径图
  - 为每个子包在 `doc.go` 中添加使用示例

- [ ] **M3.3 核心示例质量保证**
  - 确保以下示例可直接 `go run`：
    - `examples/reference`（HTTP 全功能参考）
    - `examples/ai-agent-gateway`（AI 网关参考）
    - `examples/multi-tenant-saas`（多租户参考）
  - 所有示例通过 `go build ./examples/...` 零错误

### M4：v1.0.0 发布（第 5-6 周）

- [ ] **M4.1 质量门禁脚本**
  - 完善 `scripts/v1-release-readiness-check.sh`
  - 加入 benchmark 基准回归检测

- [ ] **M4.2 版本标签与 Release**
  - 更新 README.md badge 为 v1.0.0
  - 更新 README_CN.md 同步
  - 创建 `v1.0.0` Git 标签
  - 发布 GitHub Release 附带 CHANGELOG

---

## 五、AI 层深化计划（中期：1-3 个月）

**目标**：将分散的 AI 子包整合为内聚、稳定、可生产使用的 AI 网关层。

### A1：AI 层稳定性分级与 API 冻结

按以下优先级推进稳定化：

**第一批（v1.1，立即稳定）**：
- `ai/provider` — 核心接口冻结，统一错误规范
- `ai/tool` — Tool 接口冻结
- `ai/sse` — SSE 协议实现稳定
- `ai/streaming` — 流式读写接口稳定

**第二批（v1.2，3-6 周稳定）**：
- `ai/session` — 会话管理
- `ai/ratelimit` — AI 请求速率控制
- `ai/resilience` — 弹性包装
- `ai/metrics` — AI 指标采集

**第三批（v1.3，实验转稳定）**：
- `ai/orchestration` — 完善测试
- `ai/llmcache` — 缓存一致性保证
- `ai/filter` — 内容过滤完善

**长期实验**：
- `ai/distributed` — 完成 SPRINT5_PLAN
- `ai/semanticcache` — 完成 SPRINT2_ENHANCEMENTS
- `ai/marketplace` — 设计评审后决定是否保留
- `ai/multimodal` — 依赖 Provider 适配进度

### A2：声明式 Tool Schema 框架（重要差距）

当前 Tool 定义需要手写 JSON Schema，开发体验差：

```go
// 当前方式（繁琐）
tool := provider.Tool{
    Function: provider.FunctionDef{
        Name: "search",
        Parameters: map[string]any{"type": "object", "properties": ...},
    },
}

// 目标方式（声明式）
type SearchInput struct {
    Query  string `tool:"description=搜索关键词,required"`
    Limit  int    `tool:"description=返回数量,default=10"`
}
tool := ai.NewTool[SearchInput]("search", "执行搜索", func(ctx context.Context, input SearchInput) (any, error) {
    // ...
})
```

**任务**：
- [ ] **A2.1** 在 `ai/tool` 中实现 `NewTool[T any]` 泛型工厂函数
  - 使用 Go reflect 将结构体字段自动转换为 JSON Schema
  - 自定义 struct tag `tool:"description=...,required,default=..."`
- [ ] **A2.2** 自动处理工具调用循环（Agentic Loop）
  - `ai/tool` 提供 `ToolExecutor.RunLoop(ctx, provider, messages, tools)` 自动执行 tool use 循环直到 `end_turn`
- [ ] **A2.3** 工具并行调用支持
  - `ai/tool/parallel` 已存在，完善并集成到主流程

### A3：MCP（Model Context Protocol）支持

MCP 是 AI Agent 生态的关键协议，当前完全缺失。这是 plumego 定位 AI-First 框架的重要短板。

- [ ] **A3.1** 新建 `ai/mcp` 包，实现 MCP Server 端
  - 支持 tools/list、tools/call 端点
  - 基于 `net/http` 实现，无外部依赖
- [ ] **A3.2** 集成 `router.Router` 挂载 MCP 端点
  - `app.MountMCP("/mcp", toolRegistry)` 一行挂载
- [ ] **A3.3** MCP Client 端（按需）
  - 允许 plumego 应用作为 MCP 客户端调用外部 MCP Server 提供的工具

### A4：统一 Agent 框架

当前 AI 能力分散，缺乏统一的 Agent 抽象：

- [ ] **A4.1** 新建 `ai/agent` 包，定义核心抽象：
  ```go
  type Agent interface {
      Run(ctx context.Context, input string) (*AgentResult, error)
      Stream(ctx context.Context, input string) (<-chan AgentEvent, error)
  }
  ```
- [ ] **A4.2** 实现 `ReActAgent`：基于 Reasoning + Acting 循环
- [ ] **A4.3** 实现 `PlanAndExecuteAgent`：先规划后执行
- [ ] **A4.4** 提供 Agent 的持久化（依赖 `ai/session`）
- [ ] **A4.5** 与多租户层集成：每个租户的 Agent 受配额/策略约束

### A5：Provider 生态扩展

- [ ] **A5.1** Google Gemini Provider 实现（高优先级）
- [ ] **A5.2** 本地模型支持（Ollama 协议适配）
- [ ] **A5.3** Provider 健康检查与自动切换（failover）
  - 基于 `ai/resilience` 的 circuit breaker 实现
- [ ] **A5.4** Provider 费用追踪
  - 基于 token 用量自动计算费用，集成到 `ai/metrics`

---

## 六、多租户层完善计划（中期：1-3 个月）

**目标**：将实验性的多租户能力提升为生产级，成为 plumego 的核心竞争力之一。

### T1：API 稳定与测试加固

- [ ] **T1.1** `tenant/` 包 API 冻结
  - 在接口文件上标注 `// Stable`
  - 补充压力测试：10000 租户并发配额检查
- [ ] **T1.2** 完善隔离性测试
  - 跨租户数据泄漏测试矩阵
  - `store/db/tenant_isolation_test.go` 扩展

### T2：租户感知 AI 层

- [ ] **T2.1** 将 AI 请求自动关联到租户上下文
  - `ai/provider` 请求携带租户 ID，受租户速率限制约束
  - Token 用量记录到租户配额
- [ ] **T2.2** 租户级 AI 策略
  - 允许的 Provider、Model、最大 Token 数的租户级配置
  - 集成到 `tenant/policy.go`

### T3：数据库分片与迁移

- [ ] **T3.1** DB 迁移 Runner
  - 新建 `store/db/migrate` 包
  - 支持 up/down 迁移，记录版本历史
  - 集成到 `core.App` 生命周期（`app.WithMigrations(db, fs)`）
- [ ] **T3.2** 多租户 DB Sharding 策略
  - `store/db/sharding` 与 `tenant` 层集成
  - 支持租户 → 分片映射

---

## 七、可观测性完善计划（中期：1-2 个月）

**目标**：端到端可观测性，涵盖 HTTP 层 → AI 编排层 → Provider 调用。

### O1：统一 Observability 组件

- [ ] **O1.1** 完善 `core.ConfigureObservability` 为一行配置
  ```go
  app.ConfigureObservability(core.ObservabilityConfig{
      Metrics: core.MetricsConfig{Enabled: true, Namespace: "myapp"},
      Tracing: core.TracingConfig{Enabled: true, Exporter: "otlp"},
      Logging: core.LoggingConfig{Level: slog.LevelInfo},
  })
  ```
- [ ] **O1.2** 确保所有中间件自动注入 trace_id 和 span_id

### O2：AI 链路追踪

- [ ] **O2.1** 在 `ai/instrumentation` 中完善 Provider 调用链追踪
  - Span: provider.complete, tool.execute, orchestration.step
  - 属性: model, token_usage, latency_ms, tenant_id
- [ ] **O2.2** 与 OpenTelemetry 导出器集成

### O3：性能基准 CI 集成

- [ ] **O3.1** 建立基准基线
  - Router：10 万次/秒路由匹配
  - Middleware Chain（5 层）：添加 ≤ 5μs 开销
  - JSON 序列化：≤ 2μs/请求
- [ ] **O3.2** 将 `router/router_bench_test.go` 纳入 CI
  - 基准回归阈值：±10%

---

## 八、开发者体验计划（中期：1-3 个月）

**目标**：让 plumego 成为 AI 应用开发者的首选框架。

### D1：plumego CLI 完善

- [ ] **D1.1** `plumego new` 脚手架命令
  - 模板：web-api、ai-gateway、multi-tenant-saas
  - 生成完整可运行的项目结构
- [ ] **D1.2** `plumego dev` 稳定性
  - 完善 `cmd/plumego/DEV_SERVER.md`
  - 热重载在 Linux/macOS/Windows 三平台测试通过
- [ ] **D1.3** `plumego check` 质量检查
  - 封装 `scripts/v1-release-readiness-check.sh` 为 CLI 命令

### D2：示例完善

优先完善以下高价值示例：

- [ ] **D2.1** `examples/ai-agent-gateway` — AI 网关完整示例
  - 多 Provider 负载均衡、速率限制、租户隔离、流式响应
- [ ] **D2.2** `examples/agents` — 多 Agent 协作示例
  - ReAct Agent + Tool Use + Session 持久化
- [ ] **D2.3** `examples/mcp-server` — MCP Server 示例（待 A3 完成后）
- [ ] **D2.4** 所有示例通过 `go test -run TestExample ./examples/...`

### D3：文档体系

- [ ] **D3.1** `docs/ai/` 目录结构
  ```
  docs/ai/
    README.md          # AI 层全貌，子包导航
    provider.md        # Provider 接入指南
    tool.md            # Tool 开发指南
    agent.md           # Agent 开发指南
    mcp.md             # MCP 协议指南
    tenant.md          # 多租户 AI 配置
    observability.md   # AI 可观测性
  ```
- [ ] **D3.2** 更新 README.md 的 AI 层章节
  - 反映 18+ AI 子包的实际能力
  - 添加 AI 应用快速开始示例
- [ ] **D3.3** README_CN.md 同步更新

---

## 九、v2.0 长远规划（远期：3-12 个月）

这些方向依赖 v1.x 的稳定化，按成熟度逐步推进。

### V2.1：多 Agent 编排框架（正式版）

- 基于 `ai/distributed` 的生产级多 Agent 分布式调度
- 支持 Agent 间消息传递、状态共享、并行执行
- 与 `pubsub` 集成实现 Agent 事件流

### V2.2：向量数据库原生集成

- `ai/semanticcache` 向量存储支持持久化后端
  - pgvector（PostgreSQL 扩展）
  - 内置高性能 HNSW 内存索引
- RAG（检索增强生成）工具链

### V2.3：AI 应用安全框架

- Prompt Injection 高级检测（基于语义分析）
- AI 输出内容过滤（有害内容、PII 检测）
- AI 请求审计日志（独立于应用日志）

### V2.4：边缘部署支持

- 支持 WebAssembly 编译目标（核心层）
- 轻量化配置，适配边缘计算场景

### V2.5：生态集成

- Kubernetes Operator 部署模式
- Prometheus Operator 监控集成
- Helm Chart 发布

---

## 十、优先级矩阵与执行顺序

```
                    ┌─────────────────────────────┐
高价值+低工作量     │  M1 质量收敛                 │ ← 立即开始
                    │  M2.1 CSRF 防护              │
                    │  M3.2 AI 文档导航            │
                    └─────────────────────────────┘
                    ┌─────────────────────────────┐
高价值+中工作量     │  A2 声明式 Tool Schema        │ ← 1-2 月内
                    │  A3 MCP 支持                 │
                    │  T1 多租户 API 冻结           │
                    │  O1 统一 Observability        │
                    └─────────────────────────────┘
                    ┌─────────────────────────────┐
高价值+高工作量     │  A4 统一 Agent 框架           │ ← 2-4 月内
                    │  T3 DB 迁移 Runner            │
                    │  D1 CLI 脚手架                │
                    │  A5 Provider 生态扩展         │
                    └─────────────────────────────┘
                    ┌─────────────────────────────┐
中价值+任意工作量   │  V2.x 长远规划               │ ← v2.0 周期
                    └─────────────────────────────┘
```

---

## 十一、关键设计约束（不可违反）

以下约束在整个路线图执行过程中必须严格遵守：

1. **主模块零外部依赖**：`go.mod` 中 `require` 块必须为空（stdlib only）。AI 子包可以有独立 go.mod。
2. **handler 签名唯一性**：`func(http.ResponseWriter, *http.Request)` 是唯一合法的 handler 形态。
3. **错误写入单一路径**：永远通过 `contract.WriteError` 写错误响应，禁止其他方式。
4. **中间件职责单一**：中间件只做传输层（认证、日志、限流），不注入业务逻辑。
5. **构造函数 DI**：依赖项在构造函数传入，禁止通过 context 做服务定位。
6. **安全失败关闭**：认证/验证失败时默认拒绝请求，不允许降级放行。
7. **不记录敏感信息**：任何日志中禁止出现 token、secret、key、password。

---

## 十二、里程碑时间线

```
2026-03       ████ M1 质量收敛
              ████ M2 安全加固
2026-04       ████ M3 文档示例
              ████ M4 v1.0.0 GA ←── 当前目标
2026-05       ████ A1 AI 稳定性分级
              ████ A2 声明式 Tool Schema
2026-06       ████ A3 MCP 支持
              ████ T1 多租户稳定化
2026-07       ████ A4 Agent 框架
              ████ O1 统一 Observability
              ████ D1 CLI 完善
2026-09       ████ v1.2 发布（AI 层稳定）
2026-12       ████ v2.0 规划启动
```

---

## 附录 A：子包稳定性速查表

| 包路径 | 稳定性 | 说明 |
|---|---|---|
| `contract/` | ✅ Stable | API 已冻结 |
| `router/` | ✅ Stable | API 已冻结 |
| `middleware/` | ✅ Stable | API 已冻结 |
| `security/jwt` | ✅ Stable | 可生产使用 |
| `security/headers` | ✅ Stable | 可生产使用 |
| `security/abuse` | ✅ Stable | 可生产使用 |
| `security/input` | ✅ Stable | 可生产使用 |
| `security/csrf` | ❌ 缺失 | 待实现 |
| `config/` | ⚠️ 存在全局状态 | 待清理 |
| `health/` | ✅ Stable | 可生产使用 |
| `tenant/` | ⚠️ Beta | API 待冻结 |
| `store/db` | ✅ Stable | 可生产使用 |
| `store/db/rw` | ✅ Stable | 可生产使用 |
| `store/db/sharding` | ⚠️ Beta | 需进一步测试 |
| `store/cache` | ✅ Stable | 可生产使用 |
| `store/file` | ⚠️ Beta | S3 依赖问题待解 |
| `store/kv` | ✅ Stable | 可生产使用 |
| `store/idempotency` | ✅ Stable | 可生产使用 |
| `ai/provider` | ⚠️ Beta | API 接近稳定 |
| `ai/tool` | ⚠️ Beta | 待声明式增强 |
| `ai/sse` | ⚠️ Beta | 接近稳定 |
| `ai/streaming` | ⚠️ Beta | 接近稳定 |
| `ai/session` | ⚠️ Beta | 待完善 |
| `ai/ratelimit` | ⚠️ Beta | 待完善 |
| `ai/resilience` | ⚠️ Beta | 待完善 |
| `ai/metrics` | ⚠️ Beta | 待完善 |
| `ai/orchestration` | 🧪 Alpha | 测试不足 |
| `ai/llmcache` | 🧪 Alpha | 一致性待验证 |
| `ai/filter` | 🧪 Alpha | 规则待完善 |
| `ai/distributed` | 🧪 Alpha | SPRINT5 未完成 |
| `ai/semanticcache` | 🧪 Alpha | SPRINT2 未完成 |
| `ai/multimodal` | 🧪 Alpha | Provider 适配不完整 |
| `ai/marketplace` | 🧪 Alpha | 实验性 |
| `ai/mcp` | ❌ 缺失 | 待实现 |
| `ai/agent` | ❌ 缺失 | 待实现 |
| `validator/` | ✅ Stable | 可生产使用 |
| `utils/` | ✅ Stable | 低稳定性要求 |

---

*本路线图随项目演进动态更新。任何 API 变更必须更新本文档对应章节。*
