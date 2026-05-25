# Plumego 架构审查报告（中文）

> 审查日期：2026-05-25
> 审查范围：全仓库稳定根包、x/* 扩展包、文档与规格文件
> 审查目标：识别冗余、不一致、不清晰和过时的问题

---

## 一、module.yaml 公开 API 列表缺失（文档与代码不一致）

### 1.1 `core/module.yaml` — 缺少 `Group`、`Run`、`RouteGroup`

`public_entrypoints` 列出了 `(*App).Get/Post/Put/Delete/Patch/Any`，但以下实际存在的公开符号未收录：

| 缺失符号 | 文件位置 |
|---|---|
| `(*App).Group` | `core/routing.go:91` |
| `(*App).Run` | `core/lifecycle.go:18` |
| `RouteGroup`（类型及其全部方法） | `core/routing.go:98` |

`RouteGroup` 是 `Group()` 的返回类型，暴露了 `.Get/.Post/.Put/.Delete/.Patch/.AddRoute/.Group` 共 7 个方法。机器可读规格无法识别这一公开表面，AI 代理在发现路由分组功能时会产生盲区。

### 1.2 `log/module.yaml` — 遗漏 5 个使用必需的类型

`module.yaml` 仅列出 `StructuredLogger` 和 `NewLogger`，但调用 `NewLogger` 时必须用到的类型均未收录：

```
Level        — 外部包直接使用（如 glog.go 中的枚举值）
Fields       — 在 middleware/recovery、x/gateway/ipc 等处直接引用
LoggerFormat — 在 x/observability/ops 中使用 LoggerFormatDiscard
LoggerConfig — 在 x/observability/ops、x/observability/devtools 中使用
```

### 1.3 `contract/module.yaml` — `ErrHandlerNil`、`ErrResponseWriterNil` 藏在通配符下

```yaml
- Err* transport sentinel errors   # 仅有通配符，无具体枚举
```

`ErrHandlerNil` 和 `ErrResponseWriterNil` 并非"transport sentinel"，而是**构造错误（construction error）**。`core/routing.go:23` 直接引用 `contract.ErrHandlerNil`，但规格文件无法通过精确名称匹配找到它，语义分类也有误。

---

## 二、Resilience 层多重实现并存 — 最严重的结构问题

### 2.1 Circuit Breaker 双重实现

| 实现 | 路径 | 使用方 |
|---|---|---|
| 通用 CB | `x/resilience/circuitbreaker/` | `x/gateway` |
| AI 专用 CB | `x/ai/circuitbreaker/` | `x/ai/resilience` |

两个实现的 API 形状不同：

```go
// x/resilience：Call(fn)，Counts 含 FailureRate()/SuccessRate()
circuitbreaker.New(config Config)

// x/ai：Execute(fn)，提供两个工厂函数
NewCircuitBreakerWithConfig(config Config)
NewCircuitBreaker(name, maxFailures, resetTimeout)  // 3 参数生成器
```

`x/ai/module.yaml` 本身已承认这一问题：

```yaml
- resilience  # Wrapper API may be consolidated with circuitbreaker/ratelimit
```

已知问题，但以 `experimental` 标签搁置，未推进整合。

### 2.2 Rate Limiter 四处分散

```
security/abuse/limiter.go     — 稳定层，为 HTTP 中间件适配器提供基础
x/resilience/ratelimit/       — 通用令牌桶，x/messaging、x/tenant 等共享
x/ai/ratelimit/               — AI 专用令牌桶，仅供 x/ai/resilience 使用
x/tenant/core/ratelimit.go    — 租户策略专用
```

`x/ai/ratelimit` 和 `x/resilience/ratelimit` 各自独立实现了相同的令牌桶算法。
`x/ai/resilience.ResilientProvider` 使用 `x/ai/ratelimit` 而非 `x/resilience/ratelimit`，
导致 `x/resilience` 的设计意图对 `x/ai` 形同虚设。

---

## 三、`x/ai/metrics` — 与稳定 `metrics` 包平行的独立接口

稳定 `metrics` 包导出：`AggregateCollector`、`Recorder`、`HTTPObserver`（无通用 `Collector` 接口）。

`x/ai/metrics` 自行定义了独立的接口体系：

```go
// x/ai/metrics 自定义
type Collector interface {
    Counter(name string, value float64, tags ...Tag)
    Gauge(name string, value float64, tags ...Tag)
    Histogram(name string, value float64, tags ...Tag)
    Timing(name string, duration time.Duration, tags ...Tag)
}
type Tag struct { Key, Value string }
// 还有 MemoryCollector、PrometheusExporter 等完整实现
```

`x/ai/metrics` 与稳定 `metrics` 之间没有适配器，也没有官方说明其分离原因，形成了一个孤立的指标子栈。

---

## 四、AGENTS.md 规则违反 — panic-only 构造函数

AGENTS.md 明确规定：*"No new panic-only constructors for fallible behavior"*

以下函数违反此规定：

| 函数 | 文件 |
|---|---|
| `NewManager(opts ...ManagerOption) *Manager` | `x/ai/provider/manager.go:26,55` |
| `NewStreamManager() *StreamManager` | `x/ai/streaming/streaming.go:95,107` |
| `NewResilientProvider(config Config) *ResilientProvider` | `x/ai/resilience/provider.go:42,45` |
| `x/ai/semanticcache` 中的构造函数 | `x/ai/semanticcache/provider.go:74` |
| `x/ai/metrics` 中的构造函数 | `x/ai/metrics/metrics.go:62` |

`NewResilientProvider` 虽与 `NewResilientProviderE`（返回 error）成对出现，但 panic 版本仍作为公开 API 保留，其余函数则无 E-variant。

---

## 五、`x/rpc/gateway` — 重复定义 `contract.ErrHandlerNil`

```go
// x/rpc/gateway/transcoder.go:11
var ErrHandlerNil = errors.New("rpc gateway handler is nil")
```

`contract.ErrHandlerNil` 已存在（`contract/context_core.go:69`），且被 `core/routing.go:23` 引用。
`x/rpc` 创建了本地副本，违反了 AGENTS.md 的"每一层保持唯一的错误构造路径"原则。

---

## 六、`store` 子包配置结构命名不一致

| 子包 | 类型名 | 工厂函数 |
|---|---|---|
| `store/cache` | `Config` | `DefaultConfig()` ✓ |
| `store/db` | `Config` | `DefaultConfig(driver, dsn)` ✓ |
| `store/kv` | `Options` | 无（`setDefaults()` 为私有） |
| `store/file` | `PutOptions` | 不适用（请求参数） |

`store/kv` 独用 `Options` 命名，且没有公开工厂函数。外部调用方无法通过 `DefaultOptions()` 获取默认值，只能手动填写 zero-value 字段。

---

## 七、`middleware` 子包构造函数模式不一致

`middleware/module.yaml` agent_hints 中已声明规则：
> "stateless packages use `Middleware(config T)`; stateful packages use `New(config T) *Type`"

实际情况：

| 包 | 实际构造函数 | 是否符合规则 |
|---|---|---|
| `compression` | `Middleware(cfg Config)` | ✓ |
| `timeout` | `Middleware(cfg Config)` | ✓ |
| `recovery` | `Middleware(config Config) (mw, error)` | ✓ |
| `accesslog` | `Middleware(config Config) (mw, error)` | ✓ |
| `cors` | `Middleware(opts CORSOptions)` | △ 类型名非 `Config` |
| `cors` | `StrictDefaultOptions(...)` | △ 非标准工厂命名 |
| `ratelimit` | `NewAbuseGuard(config AbuseGuardConfig)` | △ 有状态但配置类型名为 `AbuseGuardConfig` |
| `auth` | `Authenticate(...)`、`Authorize(...)` | △ 不遵循 `Middleware()` 模式 |

`cors` 和 `auth` 使用领域专化命名或许出于设计意图，但与规则文档的冲突会给 AI 代理造成混淆。

---

## 八、`log` 包内部类型名 — glog 品牌污染

```go
// log/glog.go:37
type gLogger struct { ... }   // 私有类型名引用了外部库 glog

// log/logger.go:48 — doc comment 中的错误引用
//   logger.Info("server started", glog.Fields{"addr": ":8080"})
//                                  ↑ 应为 log.Fields，不是 glog.Fields
```

`gLogger` 是私有类型，不构成 API 污染，但 doc comment 中出现的 `glog.Fields` 是**错误的包名**，会误导阅读文档的用户。正确写法是 `log.Fields`。

---

## 九、`middleware/module.yaml` — selection_guide 与 landing_zone 中 rate limiting 路径混乱

```yaml
selection_guide:
  rate_limiting: x/resilience/ratelimit/   # 推荐路径（原始令牌桶）

landing_zones:
  rate_limit: ratelimit/                   # 稳定中间件子包（HTTP 适配器）
```

两个条目名称相近但指向不同层次：`middleware/ratelimit` 是 HTTP 适配器，`x/resilience/ratelimit` 是原始令牌桶原语。文件中未说明两者的角色差异，容易造成导航混淆。

---

## 十、`core/module.yaml` forbidden_imports 冗余条目

```yaml
forbidden_imports:
  - x/**          # 已覆盖所有 x/ 子路径
  - x/ai/**       # 冗余 — 已被 x/** 包含
  - tenant/**     # 不存在的顶级包（实际为 x/tenant）
  - net/**        # 不存在的顶级包
  - pubsub/**     # 不存在的顶级包
  - rest/**       # 不存在的顶级包
  - validator/**  # 不存在的顶级包
  - utils/**      # 不存在的顶级包
```

`tenant`、`net`、`pubsub`、`rest`、`validator`、`utils` 是 `specs/repo.yaml` 中的 `discouraged_roots`（禁止新建的根），并非当前存在的包。将它们列入 `forbidden_imports` 既混淆了"禁止导入现有包"与"禁止创建新根"的语义，也让 `x/ai/**` 的单独列出显得多余。

---

## 十一、`(*App).Run()` — 存在但未推荐、未记录、未使用

`core/lifecycle.go:18` 存在公开方法 `Run()`，执行 `Prepare()` + `ListenAndServe`。
但是：

- `core/module.yaml` `public_entrypoints` 中**未列出**
- `docs/CANONICAL_STYLE_GUIDE.md` 第 4 节的标准引导示例使用手动 `Prepare()` + `Server()` 模式
- 全部 14 个参考应用均使用 `app.Start(ctx)` 包装了手动模式，**没有一个调用 `Run()`**

结果：`Run()` 是公开 API，但处于"官方幽灵方法"状态——既不被推荐，也未被文档化，更未被使用。

---

---

## 十二、17 个扩展包 module.yaml 缺少 `public_entrypoints`

以下扩展包的 `module.yaml` 完全没有 `public_entrypoints` 字段，但其中多数实际导出了具体符号：

| 包 | 是否有代码 | 说明 |
|---|---|---|
| `x/data/migrate` | ✓ | 有 `Runner`、`Migrator` 等公开类型 |
| `x/data/pgx` | ✓ | 有 `DB`、`Querier` 等公开类型 |
| `x/data/sqlx` | ✓ | 有 `DB`、`Querier` 等公开类型 |
| `x/frontend` | ✓ | 有完整资源服务逻辑 |
| `x/gateway/discovery` | ✓ | 有服务发现接口 |
| `x/gateway/ipc` | ✓ | 有 IPC 传输原语 |
| `x/messaging/mq` | ✓ | 有 `InProcBroker` 等 |
| `x/messaging/pubsub` | ✓ | 有 pub/sub 原语 |
| `x/messaging/scheduler` | ✓ | 有调度器原语 |
| `x/observability/devtools` | ✓ | 有完整调试工具链 |
| `x/observability` | ✓ | 有 OTel、Prometheus 导出器 |
| `x/observability/ops` | ✓ | 有运行时诊断路由 |
| `x/openapi` | ✓ | 有 `Generate`、`MarshalJSON` 等 |
| `x/rpc/client` | ✓ | 有 gRPC 客户端接口 |
| `x/rpc/gateway` | ✓ | 有转码器接口 |
| `x/rpc/server` | ✓ | 有 gRPC 服务器接口 |
| `x/validate` | ✓ | 有 `Bind`、`BindJSON`、`Validator` 等 |

`public-entrypoints-sync` 检查工具只验证 module.yaml 中已列出的符号是否在代码中存在，不会反向检测代码中的符号是否已被收录。因此这 17 个包的公开 API 对机器完全不可见。

---

## 十三、`public_entrypoints` 记录规范不一致（三种写法混用）

不同 module.yaml 使用了三种截然不同的记录方式：

| 记录方式 | 代表包 | 示例 |
|---|---|---|
| **具体符号名** | `contract`、`core`、`health` | `NewErrorBuilder`、`(*App).Group` |
| **子包名** | `security`、`store`、`x/ai`、`x/gateway` | `authn`、`jwt`、`cache`、`proxy` |
| **混合** | `x/rest` | `ResourceSpec`（符号）+ `pagination`（子包名）|

`x/websocket` 与 `x/gateway` 同为 beta 扩展包，但一个用具体符号，另一个用子包名，规范完全相反。这导致机器解析歧义：`pagination` 在 `x/rest` 中是子包，在其他地方可能是类型名。

---

## 十四、多个扩展包 `forbidden_imports` 列出不存在的路径 `core/components/**`

16 个扩展包的 `module.yaml` 都包含：
```yaml
forbidden_imports:
  - core/components/**
```

但 `core/components/` 目录从未存在过。这个条目可能是历史遗留或模板填充错误，对边界强制执行没有实际作用，但会误导阅读者认为 `core` 曾经有过 `components` 子包。

---

## 十五、`x/data/module.yaml` public_entrypoints 列出概念标签而非实际包名

```yaml
public_entrypoints:
  - topology     # 不存在的包名
  - routing      # 不存在的包名
  - composition  # 不存在的包名
  - kvengine     # ✓ 存在
  - idempotency  # ✓ 存在
```

`x/data` 实际子包：`cache`、`file`、`idempotency`、`kvengine`、`migrate`、`pgx`、`rw`、`sharding`、`sqlx`。

`topology`、`routing`、`composition` 是架构设计中的概念描述词，不是可导入的包路径。实际存在的 `cache`、`file`、`migrate`、`pgx`、`rw`、`sharding`、`sqlx` 均未在 `public_entrypoints` 中出现。

---

## 十六、`x/data/pgx` 和 `x/data/sqlx` 命名与实际不符

两个包的名称暗示它们分别是 `github.com/jackc/pgx` 和 `github.com/jmoiron/sqlx` 的适配层，但实际上：

- `x/data/pgx` 仅封装了 `stdlib` + `store/db`，没有引入 `pgx` 库，是"PostgreSQL 风格接口"而非"pgx 库适配器"
- `x/data/sqlx` 使用 `database/sql`（标准库），没有引入 `sqlx` 库，是"database/sql 风格接口"

包注释中有说明（"without importing a concrete driver"），但包名本身具有强烈的误导性，会让开发者误以为需要先安装 pgx/sqlx 依赖。

---

## 十七、`x/messaging/mq` 暴露永远失败的配置字段

`Config` 中存在两个布尔字段，设置后会立即在 `Validate()` 时报错：

```go
// x/messaging/mq/config.go:80-91
EnableMQTT bool   // 文档：not implemented，设为 true 时 Validate() 返回错误
EnableAMQP bool   // 文档：not implemented，设为 true 时 Validate() 返回错误
```

这两个字段已在 `deprecation-inventory.yaml` 中登记为 `x-mq-protocol-bridge-placeholders`（`status: keep`），但并未声明移除计划。将"永远失败"的配置字段保留在公开 API 中，比不提供这两个字段更糟糕——它们会被代码自动补全工具暴露给用户，却从不起作用。

---

## 十八、`x/websocket` 兼容别名无移除计划

`deprecation-inventory.yaml` 中的 `x-websocket-merged-main-compatibility-aliases` 条目（`status: keep`）保留了两个公开别名：

```go
// x/websocket/auth.go:144
func NewHS256TokenAuth(secret []byte) (*SimpleHS256TokenAuth, error)  // 别名

// x/websocket/auth.go:149
func (s *SimpleHS256TokenAuth) AuthenticateToken(token string) (map[string]any, error)  // 别名
```

AGENTS.md 明确规定：*"Deprecated symbols must be removed in the same PR that replaces their last caller. Do not leave dead wrappers behind."*

这两个别名有"兼容"理由（`merged-main`），但 `status: keep` 没有写明移除里程碑，与规则中"同 PR 清理"的要求相悖。

---

## 十九、`reference/workerfleet` 有 9 处 `ErrNotImplemented` 但未标记为 WIP

`reference/workerfleet/internal/app/service.go` 中的 9 个业务方法全部返回 `ErrNotImplemented`，涵盖 `RegisterWorker`、`HeartbeatWorker`、`WorkerList`、`WorkerDetail`、`TaskDetail`、`CaseTimeline` 等核心功能。

`deprecation-inventory.yaml` 中的 `reference-workerfleet-unimplemented-registration` 条目标注为 `status: keep`，说明这是刻意的占位设计。但该参考应用的 AGENTS.md 将其定位为"richer production-style reference for worker monitoring"，与大量未实现方法的现状存在落差。

---

## 问题优先级汇总

| 优先级 | 问题 | 涉及路径 |
|---|---|---|
| **高** | x/ai 独自实现 Circuit Breaker（已有 migration path 但未执行） | `x/ai/circuitbreaker/` vs `x/resilience/circuitbreaker/` |
| **高** | x/ai 独立指标接口（与稳定 metrics 完全断开） | `x/ai/metrics/` |
| **高** | panic-only 构造函数规则违反（5 处） | `x/ai/provider`, `x/ai/streaming`, `x/ai/resilience` 等 |
| **高** | 17 个扩展包 module.yaml 缺少 public_entrypoints | `x/observability`、`x/openapi`、`x/validate` 等 |
| **中** | x/data/module.yaml public_entrypoints 列出概念词而非包名 | `x/data/module.yaml` |
| **中** | module.yaml 缺失：`Group`、`Run`、`RouteGroup` | `core/module.yaml` |
| **中** | module.yaml 缺失：`Level`、`Fields`、`LoggerConfig` | `log/module.yaml` |
| **中** | x/rpc/gateway 重复定义 `contract.ErrHandlerNil` | `x/rpc/gateway/transcoder.go:11` |
| **中** | store 子包 Config/Options 命名不一致 | `store/kv` vs `store/cache`、`store/db` |
| **中** | public_entrypoints 三种记录规范混用 | 跨全部 module.yaml |
| **中** | 16 个扩展包 forbidden_imports 列出 `core/components/**`（不存在） | 所有 beta/experimental x/* 包 |
| **低** | x/data/pgx 和 x/data/sqlx 命名与实际实现不符 | `x/data/pgx/`、`x/data/sqlx/` |
| **低** | x/messaging/mq 暴露永远失败的 EnableMQTT/EnableAMQP 字段 | `x/messaging/mq/config.go` |
| **低** | x/websocket 兼容别名无移除计划 | `x/websocket/auth.go` |
| **低** | logger.go doc comment 中 `glog.Fields` 错误引用 | `log/logger.go:48` |
| **低** | middleware 构造函数模式部分不一致 | `cors`、`auth` |
| **低** | `(*App).Run()` 公开但未推荐、未记录、未使用 | `core/lifecycle.go:18` |
| **低** | core forbidden_imports 列出不存在的顶级根路径 | `core/module.yaml` |
| **低** | selection_guide 与 landing_zone 中 rate_limiting 导航混乱 | `middleware/module.yaml` |
