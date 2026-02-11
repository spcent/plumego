# Plumego 模块文档规划方案

> **版本**: v1.0 | **状态**: 规划中 | **更新时间**: 2026-02-11

## 📋 目录

- [概述](#概述)
- [文档结构](#文档结构)
- [模块优先级](#模块优先级)
- [详细规划](#详细规划)
- [实施时间线](#实施时间线)
- [文档规范](#文档规范)

---

## 概述

本文档是plumego框架的**模块文档补全规划方案**。目标是为每个核心模块提供完整、清晰、实用的文档，帮助开发者快速理解和使用框架。

### 现状分析

#### ✅ 已有文档
- `docs/ai-gateway/` - AI网关相关（7个文件，Phase 1-4计划）
- `docs/cli/` - CLI工具（9个文件）
- `docs/tenant/` - 多租户（4个文件）
- `docs/store/` - 存储（2个文件）
- `docs/pubsub/` - 发布订阅（2个文件）
- `docs/mq/` - 消息队列（1个文件）
- `docs/api/router.md` - 路由基础文档
- `docs/getting-started.md` - 快速开始

#### ❌ 缺失文档（核心模块）
- `core/` - **应用核心**（生命周期、DI、组件系统）
- `middleware/` - **中间件系统**（19个子包）
- `security/` - **安全特性**（5个子包：jwt, password, abuse, headers, input）
- `contract/` - **契约层**（Context, Error, Response, Protocol）
- `router/` - **路由系统**（仅有简单文档）
- `scheduler/` - **任务调度**（cron, delayed, retry）
- `frontend/` - **前端服务**（静态文件、嵌入式资源）
- `health/` - **健康检查**（liveness, readiness）
- `log/` - **日志系统**（结构化日志）
- `metrics/` - **指标系统**（Prometheus, OpenTelemetry）
- `config/` - **配置管理**（env, .env）
- `validator/` - **验证系统**（请求验证）
- `net/` - **网络工具**（7个子包）
- `utils/` - **工具类**（5个子包）

---

## 文档结构

建议采用分层模块文档结构：

```
docs/
├── getting-started.md                    # ✅ 已存在
├── MODULE_DOCUMENTATION_PLAN.md          # 📝 本文档（规划）
├── CHANGELOG.md                          # ✅ 已存在
│
├── modules/                              # 🆕 模块文档目录（新增）
│   ├── README.md                         # 模块文档索引
│   │
│   ├── core/                             # 核心模块
│   │   ├── README.md                     # 核心概览
│   │   ├── application.md                # App创建与生命周期
│   │   ├── components.md                 # 组件系统
│   │   ├── dependency-injection.md       # 依赖注入容器
│   │   ├── options.md                    # 函数式配置选项
│   │   └── examples.md                   # 示例代码
│   │
│   ├── router/                           # 路由系统
│   │   ├── README.md                     # 路由概览
│   │   ├── basic-routing.md              # 基础路由
│   │   ├── route-groups.md               # 路由分组
│   │   ├── path-parameters.md            # 路径参数
│   │   ├── reverse-routing.md            # 反向路由
│   │   └── examples.md                   # 示例代码
│   │
│   ├── middleware/                       # 中间件系统
│   │   ├── README.md                     # 中间件概览
│   │   ├── chain.md                      # 中间件链
│   │   ├── registry.md                   # 中间件注册表
│   │   ├── auth.md                       # 认证中间件
│   │   ├── bind.md                       # 请求绑定
│   │   ├── cache.md                      # 缓存中间件
│   │   ├── circuitbreaker.md             # 熔断器
│   │   ├── coalesce.md                   # 请求合并
│   │   ├── compression.md                # 响应压缩
│   │   ├── cors.md                       # CORS处理
│   │   ├── debug.md                      # 调试工具
│   │   ├── limits.md                     # 请求限制
│   │   ├── observability.md              # 可观测性
│   │   ├── protocol.md                   # 协议适配器
│   │   ├── proxy.md                      # HTTP反向代理
│   │   ├── ratelimit.md                  # 速率限制
│   │   ├── recovery.md                   # 恐慌恢复
│   │   ├── security.md                   # 安全头
│   │   ├── tenant.md                     # 租户路由
│   │   ├── timeout.md                    # 请求超时
│   │   ├── transform.md                  # 响应转换
│   │   ├── versioning.md                 # API版本控制
│   │   └── custom-middleware.md          # 自定义中间件开发
│   │
│   ├── contract/                         # 契约层
│   │   ├── README.md                     # 契约概览
│   │   ├── context.md                    # 请求上下文
│   │   ├── errors.md                     # 错误处理
│   │   ├── response.md                   # 响应助手
│   │   └── protocol/                     # 协议适配器
│   │       ├── http.md                   # HTTP协议
│   │       ├── grpc.md                   # gRPC适配
│   │       └── graphql.md                # GraphQL适配
│   │
│   ├── security/                         # 安全特性
│   │   ├── README.md                     # 安全概览
│   │   ├── jwt.md                        # JWT令牌管理
│   │   ├── password.md                   # 密码哈希与验证
│   │   ├── abuse.md                      # 滥用防护（速率限制）
│   │   ├── headers.md                    # 安全头策略（CSP, HSTS等）
│   │   ├── input.md                      # 输入验证（邮箱、URL、电话）
│   │   └── best-practices.md             # 安全最佳实践
│   │
│   ├── tenant/                           # ✅ 已有部分文档（需整合）
│   │   ├── README.md                     # 多租户概览
│   │   ├── configuration.md              # 租户配置管理
│   │   ├── quota.md                      # 配额管理
│   │   ├── policy.md                     # 策略评估
│   │   ├── ratelimit.md                  # 租户级速率限制
│   │   ├── database-isolation.md         # 数据库隔离
│   │   └── examples.md                   # 多租户SaaS示例
│   │
│   ├── ai/                               # ✅ 已有部分文档（需整合）
│   │   ├── README.md                     # AI网关概览
│   │   ├── provider.md                   # LLM提供商抽象
│   │   ├── session.md                    # 会话管理
│   │   ├── sse.md                        # 服务器发送事件（SSE）
│   │   ├── streaming.md                  # 流式响应
│   │   ├── tokenizer.md                  # Token计数与管理
│   │   ├── tool.md                       # 函数调用框架
│   │   ├── semantic-cache.md             # 语义缓存
│   │   ├── circuit-breaker.md            # AI调用熔断器
│   │   ├── orchestration.md              # AI工作流编排
│   │   ├── prompt.md                     # 提示词管理
│   │   ├── marketplace.md                # 模型市场
│   │   ├── multimodal.md                 # 多模态支持
│   │   ├── distributed.md                # 分布式AI特性
│   │   └── examples.md                   # AI网关示例
│   │
│   ├── scheduler/                        # 任务调度
│   │   ├── README.md                     # 调度器概览
│   │   ├── cron.md                       # Cron定时任务
│   │   ├── delayed-tasks.md              # 延迟任务
│   │   ├── retry-policies.md             # 重试策略
│   │   ├── persistence.md                # 任务持久化
│   │   └── examples.md                   # 示例代码
│   │
│   ├── store/                            # ✅ 已有部分文档（需整合）
│   │   ├── README.md                     # 存储概览
│   │   ├── cache/                        # 缓存
│   │   │   ├── README.md                 # 缓存接口
│   │   │   ├── memory.md                 # 内存缓存
│   │   │   └── redis.md                  # Redis缓存
│   │   ├── db/                           # 数据库
│   │   │   ├── README.md                 # 数据库包装器
│   │   │   └── tenant-isolation.md       # 租户隔离
│   │   ├── kv/                           # 键值存储
│   │   │   ├── README.md                 # KV存储概览
│   │   │   ├── wal.md                    # 预写日志（WAL）
│   │   │   └── lru-eviction.md           # LRU淘汰
│   │   ├── file/                         # 文件存储
│   │   │   └── README.md                 # 文件存储后端
│   │   └── idempotency/                  # 幂等性
│   │       └── README.md                 # 幂等请求处理
│   │
│   ├── net/                              # 网络工具
│   │   ├── README.md                     # 网络工具概览
│   │   ├── discovery/                    # 服务发现
│   │   │   ├── README.md                 # 服务发现概览
│   │   │   ├── static.md                 # 静态配置
│   │   │   └── consul.md                 # Consul集成
│   │   ├── http/                         # HTTP客户端
│   │   │   └── README.md                 # HTTP辅助工具
│   │   ├── ipc/                          # 进程间通信
│   │   │   └── README.md                 # Unix/Windows IPC
│   │   ├── mq/                           # 消息队列
│   │   │   └── README.md                 # ✅ 已有文档
│   │   ├── webhookin/                    # Webhook接收
│   │   │   ├── README.md                 # 入站webhook概览
│   │   │   ├── github.md                 # GitHub webhook
│   │   │   └── stripe.md                 # Stripe webhook
│   │   ├── webhookout/                   # Webhook发送
│   │   │   └── README.md                 # 出站webhook交付
│   │   └── websocket/                    # WebSocket
│   │       ├── README.md                 # WebSocket Hub
│   │       ├── authentication.md         # JWT认证
│   │       └── broadcasting.md           # 广播机制
│   │
│   ├── pubsub/                           # ✅ 已有文档
│   │   └── README.md                     # 进程内发布订阅
│   │
│   ├── frontend/                         # 前端服务
│   │   ├── README.md                     # 前端服务概览
│   │   ├── static-files.md               # 静态文件服务
│   │   └── embedded-assets.md            # 嵌入式资源
│   │
│   ├── health/                           # 健康检查
│   │   ├── README.md                     # 健康检查概览
│   │   ├── liveness.md                   # 存活探针
│   │   └── readiness.md                  # 就绪探针
│   │
│   ├── log/                              # 日志系统
│   │   ├── README.md                     # 日志概览
│   │   └── structured-logging.md         # 结构化日志
│   │
│   ├── metrics/                          # 指标系统
│   │   ├── README.md                     # 指标概览
│   │   ├── prometheus.md                 # Prometheus适配器
│   │   └── opentelemetry.md              # OpenTelemetry适配器
│   │
│   ├── config/                           # 配置管理
│   │   ├── README.md                     # 配置概览
│   │   ├── environment.md                # 环境变量
│   │   └── dotenv.md                     # .env文件解析
│   │
│   ├── validator/                        # 验证系统
│   │   ├── README.md                     # 验证概览
│   │   └── request-validation.md         # 请求验证
│   │
│   └── utils/                            # 工具类
│       ├── README.md                     # 工具概览
│       ├── httpx.md                      # HTTP工具
│       ├── jsonx.md                      # JSON工具
│       ├── pool.md                       # 对象池
│       ├── semver.md                     # 语义版本
│       └── stringsx.md                   # 字符串操作
│
├── guides/                               # 🆕 使用指南（新增）
│   ├── architecture.md                   # 架构设计
│   ├── error-handling.md                 # 错误处理模式
│   ├── testing.md                        # 测试模式
│   ├── deployment.md                     # 部署指南
│   ├── performance.md                    # 性能优化
│   └── migration.md                      # 迁移指南
│
├── api/                                  # ✅ 已存在（需扩充）
│   └── router.md                         # 路由API文档
│
├── cli/                                  # ✅ 已存在
├── ai-gateway/                           # ✅ 已存在
├── tenant/                               # ✅ 已存在
├── store/                                # ✅ 已存在
├── pubsub/                               # ✅ 已存在
├── mq/                                   # ✅ 已存在
├── migrations/                           # ✅ 已存在
├── lowcode/                              # ✅ 已存在
└── other/                                # ✅ 已存在
```

---

## 模块优先级

根据模块的核心程度和使用频率，划分优先级：

### P0 - 最高优先级（核心基础模块）

必须首先完成，这些是框架的基石：

| 模块 | 子文档数 | 预估行数 | 状态 | 理由 |
|------|---------|---------|------|------|
| `core/` | 5 | 600-800 | ❌ 缺失 | 应用生命周期、DI、组件系统 |
| `router/` | 5 | 500-700 | ⚠️ 不完整 | HTTP路由核心功能 |
| `contract/` | 4 | 400-600 | ❌ 缺失 | Context、Error、Response |
| `middleware/` | 22 | 1800-2200 | ❌ 缺失 | 请求处理管道（19个子包） |
| `config/` | 3 | 200-300 | ❌ 缺失 | 配置管理基础 |

**小计**: 5个模块，39个文档文件，**3500-4600行**

### P1 - 高优先级（安全与扩展）

安全关键和常用功能：

| 模块 | 子文档数 | 预估行数 | 状态 | 理由 |
|------|---------|---------|------|------|
| `security/` | 7 | 600-800 | ❌ 缺失 | JWT、密码、滥用防护（安全关键） |
| `scheduler/` | 5 | 400-600 | ❌ 缺失 | 后台任务调度（常用） |
| `health/` | 3 | 200-300 | ❌ 缺失 | 健康检查（生产必备） |
| `log/` | 2 | 150-250 | ❌ 缺失 | 日志系统（可观测性） |
| `metrics/` | 3 | 250-350 | ❌ 缺失 | 指标收集（可观测性） |
| `validator/` | 2 | 150-250 | ❌ 缺失 | 请求验证 |

**小计**: 6个模块，22个文档文件，**1750-2550行**

### P2 - 中优先级（高级特性）

已有部分文档或高级功能：

| 模块 | 子文档数 | 预估行数 | 状态 | 理由 |
|------|---------|---------|------|------|
| `tenant/` | 7 | 600-800 | ⚠️ 已有4个文档 | 多租户SaaS（需整合） |
| `ai/` | 15 | 1200-1600 | ⚠️ 已有7个文档 | AI网关（需整合） |
| `store/` | 9 | 700-900 | ⚠️ 已有2个文档 | 数据持久化（需补充） |
| `net/` | 10 | 800-1000 | ⚠️ 部分有文档 | 网络工具（webhook, websocket, mq） |
| `frontend/` | 3 | 200-300 | ❌ 缺失 | 静态文件服务 |

**小计**: 5个模块，44个文档文件，**3500-4600行**

### P3 - 低优先级（工具与指南）

辅助性文档：

| 模块 | 子文档数 | 预估行数 | 状态 | 理由 |
|------|---------|---------|------|------|
| `utils/` | 6 | 300-500 | ❌ 缺失 | 工具类（辅助性） |
| `guides/` | 6 | 600-800 | ❌ 缺失 | 最佳实践和指南 |
| `pubsub/` | 1 | 100-150 | ✅ 已有 | 已完成 |
| `cli/` | - | - | ✅ 已有 | 已完成 |
| `mq/` | - | - | ✅ 已有 | 已完成 |

**小计**: 5个模块，12个文档文件，**1000-1450行**

### 总计

| 优先级 | 模块数 | 文档数 | 预估行数 | 优先完成 |
|--------|-------|--------|---------|---------|
| P0 | 5 | 39 | 3500-4600 | ✅ Week 1-2 |
| P1 | 6 | 22 | 1750-2550 | ✅ Week 3 |
| P2 | 5 | 44 | 3500-4600 | ⚠️ Week 4-5 |
| P3 | 5 | 12 | 1000-1450 | ⏸ Week 6 |
| **总计** | **21** | **117** | **9750-13200** | **6周** |

---

## 详细规划

### P0 模块详细规划

#### 1. `docs/modules/core/` - 核心模块

**优先级**: P0 - 最高
**依赖**: 无
**预估工作量**: 600-800行，5个文档

##### 文档清单

| 文件名 | 内容 | 行数 | 关键章节 |
|--------|------|------|---------|
| `README.md` | 核心概览，模块职责 | 80-120 | 什么是core、核心概念、快速链接 |
| `application.md` | App创建与生命周期 | 150-200 | New(), Boot(), Shutdown(), 生命周期钩子 |
| `components.md` | 组件系统 | 150-200 | Component接口、注册、启动顺序、依赖解析 |
| `dependency-injection.md` | DI容器 | 120-150 | 注册服务、解析依赖、作用域管理 |
| `options.md` | 函数式配置选项 | 100-130 | WithAddr, WithDebug, WithTLS等选项详解 |

##### 示例代码需求
- [ ] 最小化应用示例
- [ ] 组件注册示例
- [ ] 自定义组件实现
- [ ] DI容器使用示例
- [ ] 完整配置选项示例

---

#### 2. `docs/modules/router/` - 路由系统

**优先级**: P0 - 最高
**依赖**: core
**预估工作量**: 500-700行，5个文档

##### 文档清单

| 文件名 | 内容 | 行数 | 关键章节 |
|--------|------|------|---------|
| `README.md` | 路由概览 | 80-100 | Trie路由器、路由注册方法、性能特点 |
| `basic-routing.md` | 基础路由 | 120-150 | GET/POST/PUT/DELETE、静态路由、处理器签名 |
| `route-groups.md` | 路由分组 | 100-130 | Group()、嵌套分组、分组中间件 |
| `path-parameters.md` | 路径参数 | 100-120 | :id语法、参数提取、通配符 |
| `reverse-routing.md` | 反向路由 | 100-150 | 路由命名、URL生成 |

##### 示例代码需求
- [ ] RESTful API路由示例
- [ ] 路由分组示例
- [ ] 参数提取示例
- [ ] 反向路由示例

---

#### 3. `docs/modules/contract/` - 契约层

**优先级**: P0 - 最高
**依赖**: router
**预估工作量**: 400-600行，5个文档（含protocol子目录）

##### 文档清单

| 文件名 | 内容 | 行数 | 关键章节 |
|--------|------|------|---------|
| `README.md` | 契约概览 | 60-80 | 什么是contract、Context vs http.Request |
| `context.md` | 请求上下文 | 150-200 | Param(), Query(), Bind(), JSON(), Stream() |
| `errors.md` | 错误处理 | 120-150 | Error类型、分类、NewError(), WriteError() |
| `response.md` | 响应助手 | 70-100 | JSON(), XML(), Stream(), Redirect() |
| `protocol/http.md` | HTTP协议 | 50-70 | HTTP适配器 |

##### 示例代码需求
- [ ] Context使用示例
- [ ] 结构化错误示例
- [ ] JSON响应示例
- [ ] 流式响应示例

---

#### 4. `docs/modules/middleware/` - 中间件系统

**优先级**: P0 - 最高（最复杂）
**依赖**: router, contract
**预估工作量**: 1800-2200行，22个文档

##### 文档清单

| 文件名 | 内容 | 行数 | 关键章节 |
|--------|------|------|---------|
| `README.md` | 中间件概览 | 120-150 | 中间件模式、执行顺序、19个内置中间件列表 |
| `chain.md` | 中间件链 | 80-100 | Chain类型、Use()、Apply() |
| `registry.md` | 中间件注册表 | 80-100 | Registry、全局vs局部 |
| `auth.md` | 认证中间件 | 100-120 | Bearer token、Basic auth、自定义认证 |
| `bind.md` | 请求绑定 | 80-100 | JSON/XML/Form绑定、验证 |
| `cache.md` | 缓存中间件 | 100-120 | HTTP缓存、ETag、Cache-Control |
| `circuitbreaker.md` | 熔断器 | 100-120 | 熔断状态、配置、降级策略 |
| `coalesce.md` | 请求合并 | 80-100 | 相同请求去重 |
| `compression.md` | 响应压缩 | 80-100 | Gzip、Brotli |
| `cors.md` | CORS处理 | 100-120 | 允许源、方法、头 |
| `debug.md` | 调试工具 | 60-80 | 请求dump、响应记录 |
| `limits.md` | 请求限制 | 80-100 | Body大小、并发数 |
| `observability.md` | 可观测性 | 100-120 | 链路追踪、指标收集 |
| `protocol.md` | 协议适配器 | 60-80 | gRPC、GraphQL适配 |
| `proxy.md` | HTTP反向代理 | 120-150 | 负载均衡、重写规则 |
| `ratelimit.md` | 速率限制 | 100-120 | Token bucket、滑动窗口 |
| `recovery.md` | 恐慌恢复 | 80-100 | panic捕获、错误响应 |
| `security.md` | 安全头 | 100-120 | CSP、HSTS、X-Frame-Options |
| `tenant.md` | 租户路由 | 100-120 | 租户识别、隔离 |
| `timeout.md` | 请求超时 | 80-100 | Context超时、超时响应 |
| `transform.md` | 响应转换 | 80-100 | 响应修改、格式转换 |
| `versioning.md` | API版本控制 | 80-100 | Header版本、URL版本 |
| `custom-middleware.md` | 自定义中间件开发 | 120-150 | 签名、最佳实践、测试 |

##### 示例代码需求
- [ ] 中间件链示例
- [ ] 每个中间件的配置示例
- [ ] 自定义中间件开发示例
- [ ] 中间件组合模式

---

#### 5. `docs/modules/config/` - 配置管理

**优先级**: P0 - 最高
**依赖**: 无
**预估工作量**: 200-300行，3个文档

##### 文档清单

| 文件名 | 内容 | 行数 | 关键章节 |
|--------|------|------|---------|
| `README.md` | 配置概览 | 60-80 | 配置来源、优先级 |
| `environment.md` | 环境变量 | 80-120 | LoadEnv()、环境变量列表 |
| `dotenv.md` | .env文件解析 | 60-100 | .env语法、加载顺序 |

---

### P1 模块详细规划

#### 6. `docs/modules/security/` - 安全特性

**优先级**: P1 - 高
**依赖**: middleware
**预估工作量**: 600-800行，7个文档

##### 文档清单

| 文件名 | 内容 | 行数 | 关键章节 |
|--------|------|------|---------|
| `README.md` | 安全概览 | 80-100 | 安全模块职责、关键原则 |
| `jwt.md` | JWT令牌管理 | 120-150 | 签发、验证、刷新、密钥轮换 |
| `password.md` | 密码哈希与验证 | 80-100 | Bcrypt、强度验证 |
| `abuse.md` | 滥用防护 | 100-120 | 速率限制、IP黑名单 |
| `headers.md` | 安全头策略 | 120-150 | CSP、HSTS、XSS保护、Frame Options |
| `input.md` | 输入验证 | 80-100 | 邮箱、URL、电话验证 |
| `best-practices.md` | 安全最佳实践 | 120-150 | OWASP Top 10、安全检查清单 |

---

#### 7-11. 其他P1模块（略，详见上方优先级表格）

---

### P2 模块详细规划

#### 12. `docs/modules/tenant/` - 多租户

**优先级**: P2 - 中
**依赖**: security, config
**预估工作量**: 600-800行，7个文档
**现有文档**: 4个（需整合到新结构）

##### 文档清单

| 文件名 | 内容 | 行数 | 关键章节 | 现有文档 |
|--------|------|------|---------|---------|
| `README.md` | 多租户概览 | 80-100 | 什么是多租户、使用场景 | - |
| `configuration.md` | 租户配置管理 | 100-120 | ConfigManager、内存/DB存储 | ⚠️ TENANT_QUICKSTART.md |
| `quota.md` | 配额管理 | 100-120 | QuotaManager、配额强制执行 | ⚠️ TENANT_QUICKSTART.md |
| `policy.md` | 策略评估 | 100-120 | PolicyEvaluator、策略决策 | - |
| `ratelimit.md` | 租户级速率限制 | 80-100 | 每租户限流 | - |
| `database-isolation.md` | 数据库隔离 | 100-120 | TenantDB、自动过滤 | ⚠️ TENANT_QUICKSTART.md |
| `examples.md` | 多租户SaaS示例 | 140-160 | 完整示例应用 | ⚠️ TENANT_PRODUCTION_PLAN.md |

**整合策略**:
- 将现有 `docs/tenant/TENANT_QUICKSTART.md` 的内容拆分到 configuration.md, quota.md, database-isolation.md
- 将 `TENANT_PRODUCTION_PLAN.md` 的生产实践整合到 examples.md
- 保留 `TENANT_IMPLEMENTATION_CHECKLIST.md` 和 `TENANT_PACKAGE_STATUS.md` 作为内部文档

---

#### 13. `docs/modules/ai/` - AI网关

**优先级**: P2 - 中
**依赖**: middleware, security
**预估工作量**: 1200-1600行，15个文档
**现有文档**: 7个Phase文档（需整合）

##### 文档清单

| 文件名 | 内容 | 行数 | 关键章节 | 现有文档 |
|--------|------|------|---------|---------|
| `README.md` | AI网关概览 | 100-120 | 什么是AI网关、能力地图 | ⚠️ Phase 1-4 |
| `provider.md` | LLM提供商抽象 | 120-150 | Provider接口、Claude/OpenAI实现 | ⚠️ Phase 1 |
| `session.md` | 会话管理 | 100-120 | SessionManager、上下文窗口 | ⚠️ Phase 1 |
| `sse.md` | 服务器发送事件 | 100-120 | SSE协议、流式响应 | ⚠️ Phase 1 |
| `streaming.md` | 流式响应 | 80-100 | 流式处理、背压控制 | ⚠️ Phase 2 |
| `tokenizer.md` | Token计数 | 80-100 | Token计数、配额管理 | ⚠️ Phase 2 |
| `tool.md` | 函数调用框架 | 100-120 | Tool定义、执行 | ⚠️ Phase 3 |
| `semantic-cache.md` | 语义缓存 | 100-120 | 向量嵌入、相似度匹配 | ⚠️ Phase 3 |
| `circuit-breaker.md` | AI调用熔断器 | 80-100 | 熔断策略、降级 | ⚠️ Phase 2 |
| `orchestration.md` | AI工作流编排 | 120-150 | 工作流定义、执行 | ⚠️ Phase 4 |
| `prompt.md` | 提示词管理 | 80-100 | 提示模板、版本控制 | ⚠️ Phase 3 |
| `marketplace.md` | 模型市场 | 80-100 | 模型发现、计费 | ⚠️ Phase 4 |
| `multimodal.md` | 多模态支持 | 80-100 | 图像、音频处理 | ⚠️ Phase 4 |
| `distributed.md` | 分布式AI特性 | 80-100 | 负载均衡、跨区域 | ⚠️ Phase 4 |
| `examples.md` | AI网关示例 | 150-180 | 完整示例应用 | ⚠️ examples/ai-agent-gateway |

**整合策略**:
- 将Phase 1-4文档的技术细节提取到各个主题文档中
- 保留Phase文档作为历史设计记录（移到 docs/ai-gateway/archive/）

---

#### 14-16. 其他P2模块（store/, net/, frontend/）（略）

---

### P3 模块详细规划（略）

---

## 实施时间线

### Week 1: P0 核心基础（core, router, contract）

| 天数 | 任务 | 交付物 | 行数 |
|------|------|--------|------|
| Day 1-2 | `docs/modules/core/` | 5个文档 | 600-800 |
| Day 3-4 | `docs/modules/router/` | 5个文档 | 500-700 |
| Day 5 | `docs/modules/contract/` | 5个文档 | 400-600 |

**Week 1 小计**: 15个文档，1500-2100行

---

### Week 2: P0 核心扩展（middleware, config）

| 天数 | 任务 | 交付物 | 行数 |
|------|------|--------|------|
| Day 1-4 | `docs/modules/middleware/` | 22个文档（重点） | 1800-2200 |
| Day 5 | `docs/modules/config/` | 3个文档 | 200-300 |

**Week 2 小计**: 25个文档，2000-2500行

---

### Week 3: P1 安全与扩展

| 天数 | 任务 | 交付物 | 行数 |
|------|------|--------|------|
| Day 1-2 | `docs/modules/security/` | 7个文档 | 600-800 |
| Day 3 | `docs/modules/scheduler/` + `health/` | 8个文档 | 600-900 |
| Day 4 | `docs/modules/log/` + `metrics/` | 5个文档 | 400-600 |
| Day 5 | `docs/modules/validator/` | 2个文档 | 150-250 |

**Week 3 小计**: 22个文档，1750-2550行

---

### Week 4-5: P2 高级特性

| 时间 | 任务 | 交付物 | 行数 |
|------|------|--------|------|
| Week 4 Day 1-3 | `docs/modules/tenant/` 整合 | 7个文档 | 600-800 |
| Week 4 Day 4-5 | `docs/modules/ai/` 整合（Part 1） | 8个文档 | 650-850 |
| Week 5 Day 1-2 | `docs/modules/ai/` 整合（Part 2） | 7个文档 | 550-750 |
| Week 5 Day 3-4 | `docs/modules/store/` 整合 | 9个文档 | 700-900 |
| Week 5 Day 5 | `docs/modules/net/` + `frontend/` | 13个文档 | 1000-1300 |

**Week 4-5 小计**: 44个文档，3500-4600行

---

### Week 6: P3 工具与收尾

| 天数 | 任务 | 交付物 | 行数 |
|------|------|--------|------|
| Day 1 | `docs/modules/utils/` | 6个文档 | 300-500 |
| Day 2-4 | `docs/guides/` | 6个文档 | 600-800 |
| Day 5 | `docs/modules/README.md` 索引 + 整体审查 | 1个文档 + 审查 | 100-150 |

**Week 6 小计**: 13个文档，1000-1450行

---

### 总计

| 阶段 | 周数 | 文档数 | 行数 |
|------|------|--------|------|
| Week 1-2 | 2周 | 40个文档 | 3500-4600 |
| Week 3 | 1周 | 22个文档 | 1750-2550 |
| Week 4-5 | 2周 | 44个文档 | 3500-4600 |
| Week 6 | 1周 | 13个文档 | 1000-1450 |
| **总计** | **6周** | **119个文档** | **9750-13200行** |

---

## 文档规范

### 文档结构

每个模块文档（README.md）应遵循以下结构：

```markdown
# 模块名称

> **包路径**: `github.com/spcent/plumego/模块名` | **稳定性**: 高/中/低

## 概述

（1-2段，说明模块的职责和使用场景）

## 快速开始

```go
// 最小化示例代码
```

## 核心概念

（关键接口、类型、设计模式）

## API参考

（主要函数、方法文档）

## 示例

（完整的可运行示例）

## 最佳实践

（使用建议、常见陷阱）

## 相关文档

（链接到相关模块文档）
```

### 代码示例规范

1. **可编译**: 所有示例必须是可编译的
2. **注释**: 关键部分添加注释说明
3. **错误处理**: 展示正确的错误处理模式
4. **完整性**: 包含必要的import语句

### 术语一致性

| 中文 | 英文 | 说明 |
|------|------|------|
| 应用 | Application | core.App实例 |
| 路由器 | Router | 路由系统 |
| 中间件 | Middleware | 请求处理管道 |
| 组件 | Component | 实现Component接口的模块 |
| 处理器 | Handler | HTTP请求处理函数 |
| 上下文 | Context | 请求上下文（contract.Context） |
| 租户 | Tenant | 多租户系统中的租户 |
| 会话 | Session | AI对话会话 |

---

## 度量指标

### 文档质量KPI

- [ ] 覆盖率：117个模块文档 / 119个目标文档 = **98%+**
- [ ] 完整性：每个文档包含：概述、快速开始、API参考、示例、最佳实践
- [ ] 可用性：每个示例代码可编译通过
- [ ] 一致性：术语、结构、风格保持一致

### 进度追踪

可使用以下命令检查进度：

```bash
# 统计已完成文档数
find docs/modules -name "*.md" | wc -l

# 统计文档总行数
find docs/modules -name "*.md" -exec wc -l {} + | tail -1

# 检查缺失的README.md
find docs/modules -type d -exec sh -c '[ ! -f "$1/README.md" ] && echo "Missing: $1/README.md"' _ {} \;
```

---

## 下一步行动

### 立即执行（本次会话）

1. [x] 创建 `docs/MODULE_DOCUMENTATION_PLAN.md`（本文档）
2. [ ] 创建 `docs/modules/` 目录结构
3. [ ] 开始编写 P0 模块文档（core, router, contract）

### 后续任务

4. [ ] Week 1: 完成 P0 核心基础文档（core, router, contract）
5. [ ] Week 2: 完成 P0 扩展文档（middleware, config）
6. [ ] Week 3: 完成 P1 文档（security, scheduler, health, log, metrics, validator）
7. [ ] Week 4-5: 完成 P2 文档（tenant, ai, store, net, frontend）
8. [ ] Week 6: 完成 P3 文档（utils, guides）
9. [ ] 审查与优化：确保文档质量和一致性
10. [ ] 中文文档同步（如需要）

---

## 附录

### 参考资源

- [CLAUDE.md](../CLAUDE.md) - AI助手指南
- [README.md](../README.md) - 项目概述
- [examples/reference/](../examples/reference/) - 参考应用
- [examples/](../examples/) - 示例集合

### 文档生成工具

建议开发文档生成工具（可选）：
- 从Go代码注释生成API文档
- 自动提取示例代码
- 生成交叉引用索引

---

**文档维护者**: Claude Code (AI Assistant)
**最后更新**: 2026-02-11
**下次审查**: 完成P0文档后（预计Week 2结束）
