# Plumego 模块文档索引

> **版本**: v1.0.0-rc.1 | **Go**: 1.24+ | **状态**: 文档建设中

欢迎来到 Plumego 模块文档中心。本目录包含框架所有核心模块的详细文档。

---

## 📚 快速导航

### 核心模块 (P0 - 必读)

| 模块 | 路径 | 说明 | 状态 |
|------|------|------|------|
| **核心系统** | [core/](core/) | 应用生命周期、DI容器、组件系统 | 📝 计划中 |
| **路由系统** | [router/](router/) | HTTP路由、路径参数、路由分组 | 📝 计划中 |
| **契约层** | [contract/](contract/) | 请求上下文、错误处理、响应助手 | 📝 计划中 |
| **中间件系统** | [middleware/](middleware/) | 19个内置中间件、中间件链 | 📝 计划中 |
| **配置管理** | [config/](config/) | 环境变量、.env文件解析 | 📝 计划中 |

### 安全与扩展 (P1 - 推荐)

| 模块 | 路径 | 说明 | 状态 |
|------|------|------|------|
| **安全特性** | [security/](security/) | JWT、密码哈希、滥用防护、安全头 | 📝 计划中 |
| **任务调度** | [scheduler/](scheduler/) | Cron定时任务、延迟任务、重试策略 | 📝 计划中 |
| **健康检查** | [health/](health/) | 存活探针、就绪探针 | 📝 计划中 |
| **日志系统** | [log/](log/) | 结构化日志 | 📝 计划中 |
| **指标系统** | [metrics/](metrics/) | Prometheus、OpenTelemetry | 📝 计划中 |
| **请求验证** | [validator/](validator/) | 请求验证 | 📝 计划中 |

### 高级特性 (P2 - 进阶)

| 模块 | 路径 | 说明 | 状态 |
|------|------|------|------|
| **多租户** | [tenant/](tenant/) | 租户隔离、配额管理、策略评估 | 📝 计划中 |
| **AI网关** | [ai/](ai/) | LLM提供商、会话管理、SSE流式 | 📝 计划中 |
| **数据存储** | [store/](store/) | KV存储、缓存、数据库、文件存储 | 📝 计划中 |
| **网络工具** | [net/](net/) | WebSocket、Webhook、服务发现、MQ | 📝 计划中 |
| **前端服务** | [frontend/](frontend/) | 静态文件、嵌入式资源 | 📝 计划中 |

### 工具与指南 (P3 - 参考)

| 模块 | 路径 | 说明 | 状态 |
|------|------|------|------|
| **工具类** | [utils/](utils/) | HTTP、JSON、对象池、字符串等工具 | 📝 计划中 |
| **使用指南** | [../guides/](../guides/) | 架构、测试、部署、性能优化 | 📝 计划中 |

---

## 🚀 快速开始

如果你是第一次使用 Plumego，推荐按以下顺序阅读：

1. **[快速开始](../getting-started.md)** - 5分钟上手教程
2. **[核心系统](core/)** - 理解应用生命周期
3. **[路由系统](router/)** - 定义HTTP路由
4. **[契约层](contract/)** - 处理请求和响应
5. **[中间件系统](middleware/)** - 扩展请求处理管道
6. **[安全特性](security/)** - 保护你的应用

---

## 📖 按模块分类

### 核心基础设施

<details>
<summary><b>core/</b> - 核心系统</summary>

- 应用创建与配置
- 生命周期管理 (Boot, Shutdown)
- 组件系统 (Component接口)
- 依赖注入容器
- 函数式配置选项

**关键文档**:
- [应用生命周期](core/application.md)
- [组件系统](core/components.md)
- [依赖注入](core/dependency-injection.md)
</details>

<details>
<summary><b>router/</b> - 路由系统</summary>

- Trie路由器
- 路径参数 (`:id`, `*path`)
- 路由分组
- 反向路由
- 中间件绑定

**关键文档**:
- [基础路由](router/basic-routing.md)
- [路由分组](router/route-groups.md)
- [路径参数](router/path-parameters.md)
</details>

<details>
<summary><b>contract/</b> - 契约层</summary>

- 请求上下文 (`*Context`)
- 结构化错误处理
- 响应助手 (JSON, XML, Stream)
- 协议适配器 (HTTP, gRPC, GraphQL)

**关键文档**:
- [请求上下文](contract/context.md)
- [错误处理](contract/errors.md)
- [响应助手](contract/response.md)
</details>

<details>
<summary><b>middleware/</b> - 中间件系统</summary>

**19个内置中间件**:
- 认证 (auth)
- 请求绑定 (bind)
- 缓存 (cache)
- 熔断器 (circuitbreaker)
- 请求合并 (coalesce)
- 响应压缩 (compression)
- CORS (cors)
- 调试 (debug)
- 请求限制 (limits)
- 可观测性 (observability)
- 协议适配 (protocol)
- 反向代理 (proxy)
- 速率限制 (ratelimit)
- 恐慌恢复 (recovery)
- 安全头 (security)
- 租户路由 (tenant)
- 请求超时 (timeout)
- 响应转换 (transform)
- API版本控制 (versioning)

**关键文档**:
- [中间件概览](middleware/README.md)
- [中间件链](middleware/chain.md)
- [自定义中间件](middleware/custom-middleware.md)
</details>

### 安全与认证

<details>
<summary><b>security/</b> - 安全特性</summary>

- JWT令牌管理 (签发、验证、刷新、密钥轮换)
- 密码哈希与验证 (Bcrypt、强度验证)
- 滥用防护 (速率限制、IP黑名单)
- 安全头策略 (CSP, HSTS, X-Frame-Options)
- 输入验证 (邮箱、URL、电话)

**关键文档**:
- [JWT令牌管理](security/jwt.md)
- [密码安全](security/password.md)
- [安全最佳实践](security/best-practices.md)
</details>

### 多租户与SaaS

<details>
<summary><b>tenant/</b> - 多租户系统</summary>

- 租户配置管理
- 配额管理与强制执行
- 策略评估
- 租户级速率限制
- 数据库隔离 (TenantDB)

**关键文档**:
- [配置管理](tenant/configuration.md)
- [配额管理](tenant/quota.md)
- [数据库隔离](tenant/database-isolation.md)

**参考**: [examples/multi-tenant-saas/](../../examples/multi-tenant-saas/)
</details>

### AI与智能

<details>
<summary><b>ai/</b> - AI网关</summary>

- LLM提供商抽象 (Claude, OpenAI)
- 会话管理
- SSE流式响应
- Token计数与管理
- 函数调用框架
- 语义缓存
- AI调用熔断器
- 工作流编排

**关键文档**:
- [提供商抽象](ai/provider.md)
- [会话管理](ai/session.md)
- [SSE流式](ai/sse.md)
- [函数调用](ai/tool.md)

**参考**: [examples/ai-agent-gateway/](../../examples/ai-agent-gateway/)
</details>

### 数据持久化

<details>
<summary><b>store/</b> - 数据存储</summary>

- **缓存**: 内存缓存、Redis缓存
- **数据库**: database/sql包装器、租户隔离
- **KV存储**: WAL、LRU淘汰
- **文件存储**: 本地、云存储后端
- **幂等性**: 幂等请求处理

**关键文档**:
- [缓存系统](store/cache/)
- [数据库包装器](store/db/)
- [KV存储](store/kv/)
</details>

### 网络与通信

<details>
<summary><b>net/</b> - 网络工具</summary>

- **服务发现**: 静态配置、Consul
- **HTTP客户端**: 辅助工具
- **IPC**: Unix/Windows进程间通信
- **消息队列**: 内存消息队列
- **Webhook接收**: GitHub、Stripe
- **Webhook发送**: 出站webhook交付
- **WebSocket**: Hub、JWT认证、广播

**关键文档**:
- [服务发现](net/discovery/)
- [Webhook接收](net/webhookin/)
- [WebSocket Hub](net/websocket/)
</details>

### 后台任务

<details>
<summary><b>scheduler/</b> - 任务调度</summary>

- Cron定时任务
- 延迟任务
- 重试策略 (线性、指数退避)
- 任务持久化
- 并发控制

**关键文档**:
- [Cron任务](scheduler/cron.md)
- [延迟任务](scheduler/delayed-tasks.md)
- [重试策略](scheduler/retry-policies.md)

**参考**: [examples/scheduler-app/](../../examples/scheduler-app/)
</details>

### 可观测性

<details>
<summary><b>健康检查</b> - health/</summary>

- 存活探针 (Liveness)
- 就绪探针 (Readiness)
- 自定义健康检查

**关键文档**: [health/README.md](health/README.md)
</details>

<details>
<summary><b>日志系统</b> - log/</summary>

- 结构化日志
- 日志级别
- 上下文日志

**关键文档**: [log/README.md](log/README.md)
</details>

<details>
<summary><b>指标系统</b> - metrics/</summary>

- Prometheus适配器
- OpenTelemetry适配器
- 自定义指标

**关键文档**: [metrics/README.md](metrics/README.md)
</details>

### 配置与验证

<details>
<summary><b>config/</b> - 配置管理</summary>

- 环境变量加载
- .env文件解析
- 配置验证

**关键文档**: [config/README.md](config/README.md)
</details>

<details>
<summary><b>validator/</b> - 请求验证</summary>

- 请求参数验证
- 结构体验证

**关键文档**: [validator/README.md](validator/README.md)
</details>

### 前端与静态资源

<details>
<summary><b>frontend/</b> - 前端服务</summary>

- 静态文件服务
- 嵌入式资源 (embed.FS)
- SPA路由支持

**关键文档**: [frontend/README.md](frontend/README.md)
</details>

### 工具类

<details>
<summary><b>utils/</b> - 工具类</summary>

- HTTP工具 (httpx)
- JSON工具 (jsonx)
- 对象池 (pool)
- 语义版本 (semver)
- 字符串操作 (stringsx)

**关键文档**: [utils/README.md](utils/README.md)
</details>

---

## 🎯 按使用场景导航

### 我想构建...

| 场景 | 推荐模块 | 参考示例 |
|------|---------|---------|
| **REST API** | core + router + contract + middleware | [examples/crud-api/](../../examples/crud-api/) |
| **多租户SaaS** | core + tenant + security + store | [examples/multi-tenant-saas/](../../examples/multi-tenant-saas/) |
| **AI聊天应用** | core + ai + websocket | [examples/ai-agent-gateway/](../../examples/ai-agent-gateway/) |
| **API网关** | core + middleware/proxy + net/discovery | [examples/api-gateway/](../../examples/api-gateway/) |
| **后台任务系统** | core + scheduler + store/kv | [examples/scheduler-app/](../../examples/scheduler-app/) |
| **Webhook处理** | core + net/webhookin + security | [examples/reference/](../../examples/reference/) |
| **实时通信** | core + net/websocket + security/jwt | [examples/websocket/](../../examples/websocket/) |

---

## 📊 文档建设进度

| 优先级 | 模块数 | 文档数 | 状态 | 预计完成 |
|--------|-------|--------|------|---------|
| P0 - 核心基础 | 5 | 40 | 📝 计划中 | Week 1-2 |
| P1 - 安全扩展 | 6 | 22 | 📝 计划中 | Week 3 |
| P2 - 高级特性 | 5 | 44 | 📝 计划中 | Week 4-5 |
| P3 - 工具指南 | 5 | 13 | 📝 计划中 | Week 6 |

**完整规划**: [../MODULE_DOCUMENTATION_PLAN.md](../MODULE_DOCUMENTATION_PLAN.md)

---

## 🔗 相关资源

- **项目主页**: [README.md](../../README.md)
- **快速开始**: [getting-started.md](../getting-started.md)
- **AI助手指南**: [CLAUDE.md](../../CLAUDE.md)
- **变更日志**: [CHANGELOG.md](../CHANGELOG.md)
- **示例代码**: [examples/](../../examples/)

---

## 🤝 贡献文档

发现文档问题或想要改进？欢迎贡献！

1. 遵循 [文档规范](../MODULE_DOCUMENTATION_PLAN.md#文档规范)
2. 确保代码示例可编译
3. 保持术语一致性
4. 提交 PR 到对应模块目录

---

**维护者**: Plumego Team
**最后更新**: 2026-02-11
**文档版本**: v1.0 (建设中)
