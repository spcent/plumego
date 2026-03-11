
# Plumego v1.0 发布路线图

> Status: Historical proposal, not the current GA execution plan.
>
> For the current v1 assessment and execution path, see:
> `docs/other/ROADMAP_V1_REVIEW_AND_EXECUTION_PLAN_2026-03-11.md`
> and `docs/other/V1_GA_EXECUTION_PLAN_2026-03-10.md`.

本文档是驱动 Plumego 框架达到 v1.0 版本的核心开发计划。
我们的目标是创建一个现代、极致简化、由 AI Agent 驱动开发和迭代的 Go Web 工具包。
在此过程中，我们将打破向后兼容性，以建立一个逻辑清晰、高度可组合的全新架构。

## 核心设计原则

1.  **AI 优先 (AI-First)**: 所有 API 设计都必须追求极致的简洁和明确，便于 AI Agent 理解、生成和修改。
2.  **显式优于隐式 (Explicit over Implicit)**: 框架不应有任何“黑魔法”。控制流必须清晰、可追踪。
3.  **拥抱标准库 (Embrace Standard Library)**: 深度整合并兼容 `net/http`，非必要不引入三方依赖。
4.  **组件化架构 (Component-Based Architecture)**: 所有独立功能（路由、数据库、AI 网关）都应被封装为可插拔、可配置的组件。
5.  **零全局状态 (Zero Global State)**: 彻底消除全局变量和单例模式，强制推行依赖注入。

---

## 详细行动计划 (Action Plan)

### 第一阶段：核心再造 (Phase 1: Core Reinvention) - 奠定基石

**目标**: 建立一个极致精简、职责单一的内核，为后续所有功能提供一个稳固、清晰的平台。

-   [ ] **任务 1.1: 重塑 `core.App` 为纯粹的协调器**
    -   [ ] 剥离 `core.App` 中的路由、中间件管理等具体实现。
    -   [ ] 使其只负责：组件注册、生命周期管理（启动/关闭）、依赖关系协调。

-   [ ] **任务 1.2: 统一应用启动入口 `plumego.Run`**
    -   [ ] 创建唯一的启动函数 `plumego.Run(http.Handler, ...Option)`。
    -   [ ] 定义 `Option` 类型，用于配置日志、生命周期钩子等应用级参数。
    -   [ ] 废弃并移除项目中所有其他的应用启动方式。

-   [ ] **任务 1.3: 标准化 `core.Component` 接口**
    -   [ ] 定义一个清晰的组件接口，例如：
        ```go
        type Component interface {
            Name() string
            Register(app *App) error
            Shutdown(ctx context.Context) error
        }
        ```
    -   [ ] 确保所有核心功能模块（如 router, ai, store）都实现此接口。

-   [ ] **任务 1.4: 彻底根除全局状态**
    -   [ ] 全面审查代码库，定位并移除所有包级别的可变全局变量（特别是 `config` 和 `log` 包）。
    -   [ ] 改造为严格的构造函数依赖注入（Constructor-based DI），配置和依赖在组件初始化时传入。

### 第二阶段：路由与中间件重塑 (Phase 2: Routing & Middleware) - 明确契约

**目标**: 打造一个直观、可预测且与标准库生态完全兼容的 HTTP 层。

-   [ ] **任务 2.1: 实现显式路由 `router.Router`**
    -   [ ] 设计一个新的 `router` 组件，强制使用 `router.Add(method, path, handler)` 进行路由注册。
    -   [ ] 移除所有复杂的链式 API 和隐式的路由分组，追求路由表的最大可读性。
    -   [ ] 保留但简化反向路由功能。

-   [ ] **任务 2.2: 强制推行标准中间件签名**
    -   [ ] 统一所有中间件签名为 `func(http.Handler) http.Handler`。
    -   [ ] 提供一个辅助函数 `middleware.Chain(...)` 用于清晰地组合中间件链。
    -   [ ] 移除任何通过 `context` 注入业务逻辑的“魔法”中间件。

-   [ ] **任务 2.3: 精简 `contract.Context`**
    -   [ ] 设计一个全新的 `plumego.Context`，它仅作为 `http.ResponseWriter` 和 `*http.Request` 的轻量级包装。
    -   [ ] 核心功能仅保留：请求绑定 (`contract.BindJSON`)、响应写入 (`contract.WriteJSON`) 和结构化错误 (`contract.WriteError`)。
    -   [ ] 移除所有服务定位器或业务辅助函数。

### 第三阶段：AI 网关与工具链 (Phase 3: AI Gateway & Tooling) - 塑造核心优势

**目标**: 构建一个强大、易于扩展的 AI 能力层，使其成为 Plumego 的标志性功能。

-   [ ] **任务 3.1: 固化 `ai.Provider` 抽象接口**
    -   [ ] 为聊天、嵌入等功能定义稳定、简洁的接口。
    -   [ ] 重构现有的 `openai`, `claude` 等实现，使其符合新接口。
    -   [ ] AI Provider 的配置（如 API Key）应在注册组件时注入。

-   [ ] **任务 3.2: 建立声明式的 `ai.Tool` (Function Calling) 框架**
    -   [ ] 设计一个允许开发者通过简单的 Go 结构体（带 `json` 标签）来定义工具的框架。
    -   [ ] 框架应自动处理工具与 LLM 的交互、参数解析和函数调度。

-   [ ] **任务 3.3: 提供原生的流式响应与 SSE 支持**
    -   [ ] 创建一个高级辅助工具（如 `sse.StreamHandler`），极大简化从 LLM 到客户端的流式响应实现。

### 第四阶段：生产力与可观测性 (Phase 4: Productivity & Observability) - 对齐生产环境

**目标**: 将 Plumego 从一个工具包转变为一个“开箱即用”的生产级框架。

-   [ ] **任务 4.1: 构建统一的 `observability` 组件**
    -   [ ] 将结构化日志 (`log/slog`)、Prometheus 指标和 OpenTelemetry 链路追踪打包成一个可选组件。
    -   [ ] 实现通过一行代码 `app.Register(observability.New())` 即可启用。

-   [ ] **任务 4.2: 整合 `security` 中间件组件**
    -   [ ] 将 CORS、安全头、CSRF 防护、滥用检测等标准安全措施整合成一个 `security` 组件。
    -   [ ] 以标准中间件的形式提供，方便按需取用。

-   [ ] **任务 4.3: 建立核心路径的性能基准**
    -   [ ] 为路由匹配、中间件调用、JSON 序列化等关键路径编写性能测试（Benchmark）。
    -   [ ] 将性能基准测试纳入 CI/CD 的质量门禁。

### 第五阶段：文档、示例与 v1.0 发布 (Phase 5: Documentation, Examples & Release)

**目标**: 完善开发者体验，清理历史债务，正式发布一个稳定、可靠的 v1.0 版本。

-   [ ] **任务 5.1: 撰写全新的 `AGENTS.md`**
    -   [ ] 基于 v1.0 的新架构，重写 Agent 操作指南，明确模块边界、设计哲学和开发戒律。
    -   [ ] 这将是指导未来 AI Agent 迭代框架的“宪法”。

-   [ ] **任务 5.2: 更新 `CANONICAL_STYLE_GUIDE.md`**
    -   [ ] 将所有在重构过程中确立的新 API 范式和代码风格固化到文档中。

-   [ ] **任务 5.3: 创建全新的 `examples/v1` 示例库**
    -   [ ] 提供多个端到端的示例项目，覆盖 Web 服务、AI 应用、CRUD API 等典型场景。

-   [ ] **任务 5.4: 定义最终的质量门禁脚本 (`scripts/quality-gates.sh`)**
    -   [ ] 创建一个独立的 shell 脚本，统一执行 `gofmt`, `go vet`, `go test -race`, 和 `go test -bench`。

-   [ ] **任务 5.5: 标记并发布 `v1.0.0`**
    -   [ ] 在所有任务完成且质量门禁通过后，更新 `README.md` 和 `CHANGELOG.md`。
    -   [ ] 创建 `v1.0.0` Git 标签并发布 Release。
