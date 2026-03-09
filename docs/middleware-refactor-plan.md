# Middleware 包重构计划

> 基础：`docs/CANONICAL_STYLE_GUIDE.md`
> 原则：不考虑兼容，按最佳实践执行
> 目标：使 `middleware` 包成为纯 transport 层 cross-cutting，单一类型，无隐式流动

---

## 一、问题全景

经逐文件分析，当前 `middleware` 包存在以下违规类别：

| 违规类型 | 文件 |
|---|---|
| OOP 包装取代 plain function | `auth/auth.go` |
| 非 canonical 显式标注但未移除 | `bind/bind.go` |
| 纯 re-export 模糊包边界 | `cache/adapter.go`, `proxy/adapter.go`, `circuitbreaker/adapter.go` |
| 禁止的 response envelope wrapping | `transform/transform.go` (`WrapJSONResponse`) |
| 域策略决策混入 transport 层 | `tenant/policy.go`, `tenant/quota.go` |
| 接口重复（向后兼容残留） | `observability/logging.go` (`MetricsCollector` / `UnifiedMetricsCollector`) |
| 每请求编译正则 | `versioning/version.go` |
| context key 类型错误 | `versioning/version.go` |
| 返回类型未声明为 `middleware.Middleware` | `versioning/version.go`, `circuitbreaker/adapter.go` |
| stdlib 已提供的工具函数 | `ratelimit/limiter.go` (`min`, `max`) |
| 并发限流逻辑与 `limits` 包重叠 | `ratelimit/limiter.go` vs `limits/limit.go` |
| `responseRecorder` 跨包重复 | `observability/logging.go`, `transform/transform.go` |
| 弃用 HTTP/2 Push 接口 | `observability/logging.go` |
| Registry 文档描述执行顺序有误 | `registry.go` |

---

## 二、详细问题与重构方案

### P1 — `middleware/auth/auth.go`：OOP 包装违反 plain function 原则

**当前状态**

```go
type AuthMiddleware interface {
    Authenticate(next http.Handler) http.Handler
}

type SimpleAuthMiddleware struct {
    authToken string
    realm     string
}

func (am *SimpleAuthMiddleware) Authenticate(next http.Handler) http.Handler { ... }

// 兼容包装
func Auth(next http.HandlerFunc) http.HandlerFunc { ... }
func FromAuthMiddleware(am AuthMiddleware) middleware.Middleware { ... }
```

**问题**

- `AuthMiddleware` 接口是对 `func(http.Handler) http.Handler` 的 OOP 包装，增加了无意义的间接层
- `SimpleAuthMiddleware` 构造后调用 `.Authenticate()` 的使用方式打破了 middleware 统一的 `func(http.Handler) http.Handler` 签名
- `Auth(next http.HandlerFunc)` 是空 token 配置的兼容函数，fail-closed 语义隐蔽
- `FromAuthMiddleware` 适配器层完全多余

**目标**

```go
// SimpleAuth 返回验证 Bearer token 或 X-Token header 的 Middleware。
// token 为空时 fail-closed（所有请求均被拒绝）。
func SimpleAuth(token string) middleware.Middleware {
    token = strings.TrimSpace(token)
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            if token == "" || subtle.ConstantTimeCompare([]byte(extractToken(r)), []byte(token)) != 1 {
                writeUnauthorized(w, r, "Protected Area")
                return
            }
            next.ServeHTTP(w, r)
        })
    }
}
```

**删除**：`AuthMiddleware` 接口、`SimpleAuthMiddleware` 结构体、`Auth` 函数、`FromAuthMiddleware` 函数

`auth/contract.go` 中的 `Authenticate`、`Authorize`、`AuthorizeFunc`、`SessionCheck` 均已是 plain function 模式，**保持不变**。

---

### P2 — `middleware/bind/bind.go`：删除非 canonical 包

**当前状态**

文件本身已注明：
```
// BindJSON is provided as an optional compatibility helper...
// It must not be used in canonical examples or generated scaffolding.
```

**问题**：非 canonical 代码留在核心 `middleware` 路径下，会被新用户误用。

**目标**：整包删除。调用方按 §7 规范在 handler 内显式解码：

```go
var req MyRequest
if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
    contract.WriteError(w, r, contract.NewBadRequestError("invalid_json", "invalid request body"))
    return
}
```

---

### P3 — `middleware/cache/adapter.go`、`middleware/proxy/adapter.go`、`middleware/circuitbreaker/adapter.go`：删除错位 re-export

**当前状态**

三个文件均为纯 type alias + var 别名，将 `net/gateway`、`net/gateway/cache`、`security/resilience/circuitbreaker` 的内容 re-export 到 `middleware` 命名空间。

**问题**

- §2 明确：`middleware` 只允许 "logging, recovery, timeout, request-id, CORS, auth adapters, rate-limiting, tracing, metrics"
- Cache、Proxy、CircuitBreaker 是 gateway 能力，不属于 transport cross-cutting
- `circuitbreaker/adapter.go` 额外问题：`Middleware()` 和 `MiddlewareWithErrorHandler()` 返回 `func(http.Handler) http.Handler` 而非 `middleware.Middleware`

**目标**：删除三个 adapter 包。用户直接从 source 包导入：

```go
import gatewaycache "github.com/spcent/plumego/net/gateway/cache"
import gateway "github.com/spcent/plumego/net/gateway"
import cb "github.com/spcent/plumego/security/resilience/circuitbreaker"
```

---

### P4 — `middleware/transform/transform.go`：删除 `WrapJSONResponse`，提取共享 recorder

#### P4-a：删除 `WrapJSONResponse`

**当前状态**

```go
func WrapJSONResponse(wrapperKey string) ResponseTransformer {
    return func(r *http.Response) error {
        wrapped := map[string]any{wrapperKey: data}
        // ...
    }
}
```

**问题**：§18 明确禁止："Introducing new response helper families" 和 response envelope proliferation（`{ "success": true, "data": ... }`）。

**目标**：删除 `WrapJSONResponse`。

#### P4-b：提取共享 `responseRecorder`

`observability/logging.go` 和 `transform/transform.go` 各自定义了行为相近的 `responseRecorder`，但实现不同（logging 的版本实现了 `Hijack`/`Flush`）。

**目标**：提取到 `middleware/internal/recorder/recorder.go`（internal 包，外部不可见），两处均引用同一实现。

```
middleware/
  internal/
    recorder/
      recorder.go   ← 统一 responseRecorder，实现 Hijack/Flush
```

---

### P5 — `middleware/versioning/version.go`：多处类型与性能违规

#### P5-a：返回类型声明

**当前**：`Middleware(config Config) func(http.Handler) http.Handler`

**目标**：`Middleware(config Config) middleware.Middleware`

`CustomExtractor` 同理。

#### P5-b：context key 类型

**当前**：
```go
type contextKey string
const versionContextKey contextKey = "api_version"
```

**问题**：`string` 底层类型可能与其他包的 key 碰撞（虽然类型不同，但反射时不够明确）。Go 惯例是使用不导出 struct 类型。

**目标**：
```go
type versionKey struct{}
// ctx.Value(versionKey{})
```

#### P5-c：每请求编译正则

**当前**：`extractFromAccept` 在每次请求时调用 `regexp.MustCompile`：

```go
func extractFromAccept(r *http.Request, vendorPrefix string) int {
    pattern := regexp.QuoteMeta(vendorPrefix) + `\.v(\d+)(?:\+|;|$)`
    re := regexp.MustCompile(pattern)   // ← 每次请求都编译
    // ...
}
```

**目标**：在 `Middleware()` 构造时编译一次，闭包捕获：

```go
func Middleware(config Config) middleware.Middleware {
    cfg := config.withDefaults()
    var acceptRE *regexp.Regexp
    if cfg.Strategy == StrategyAcceptHeader {
        acceptRE = regexp.MustCompile(
            regexp.QuoteMeta(cfg.VendorPrefix) + `\.v(\d+)(?:\+|;|$)`,
        )
    }
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            version := extractVersion(r, cfg, acceptRE)
            // ...
        })
    }
}
```

#### P5-d：`Config.WithDefaults()` 返回 `*Config`

**当前**：`func (c *Config) WithDefaults() *Config` —— 在 `*Config` 上定义方法，返回拷贝的指针，语义混乱。

**目标**：改为包内私有 `func applyDefaults(c Config) Config`，外部不暴露。

---

### P6 — `middleware/tenant/policy.go`、`tenant/quota.go`：域策略决策不属于 transport 层

**当前状态**

`TenantPolicy` 调用 `tenant.PolicyEvaluator.Evaluate()` 决定请求是否被允许。
`TenantQuota` 调用 `tenant.QuotaManager.Allow()` 决定配额是否超限。

**问题**

§2 明确：`middleware` 不允许 "business DTO assembly, hidden request binding, domain-policy branching"。
`PolicyEvaluator` 和 `QuotaManager` 是域级别的业务决策接口，不是 transport 关注点。

**比较**：`TenantResolver`（解析并验证 tenant ID 放入 context）和 `TenantRateLimit`（基于 tenant 的速率限制）是合法的 transport 关注点，**保留**。

**目标**：将 `policy.go` 和 `quota.go` 移出 `middleware/tenant/`，移入 `tenant/` 扩展包（§2 已将 `tenant` 列为 extension package）：

```
tenant/
  middleware/
    policy.go     ← TenantPolicy
    quota.go      ← TenantQuota
```

`middleware/tenant/` 只保留：
- `resolver.go` — `TenantResolver`
- `ratelimit.go` — `TenantRateLimit`
- `helpers.go` — 共享常量/工具

---

### P7 — `middleware/observability/logging.go`：接口重复与弃用实现

#### P7-a：统一 MetricsCollector 接口

**当前**：

```go
// 旧接口
type MetricsCollector interface {
    Observe(ctx context.Context, metrics RequestMetrics)
}

// "统一"接口（文档说是新的，但两者并存）
type UnifiedMetricsCollector interface {
    ObserveHTTP(ctx context.Context, method, path string, status, bytes int, duration time.Duration)
}
```

`Logging()` 参数类型为旧 `MetricsCollector`，内部用 type switch 判断是否实现了 `UnifiedMetricsCollector`。

**问题**：一个函数两套接口是典型的"multiple equally valid" anti-pattern（§18 禁止）。`RequestMetrics` 结构体与 `ObserveHTTP` 参数列表信息等价但不统一。

**目标**：统一为一个接口，`Logging()` 参数使用该接口：

```go
type MetricsCollector interface {
    ObserveHTTP(ctx context.Context, method, path string, status, bytes int, duration time.Duration)
}

func Logging(logger log.StructuredLogger, metrics MetricsCollector, tracer Tracer) middleware.Middleware
```

`RequestMetrics` 只在 `TraceSpan.End()` 参数中保留（tracing 扩展点需要），不对外暴露为 metrics 传递载体。

#### P7-b：移除弃用的 `http.Pusher` 实现

**当前**：`responseRecorder` 实现了 `Push(target string, opts *http.PushOptions) error`。

Go 1.21 起 `http.Pusher` 已弃用，Go 1.24（本项目工具链）中 HTTP/2 server push 在标准库中已不推荐使用。

**目标**：删除 `Push` 方法。

---

### P8 — `middleware/ratelimit/limiter.go`：stdlib 函数重复，与 `limits` 包功能重叠

#### P8-a：删除自定义 `min`/`max`

**当前**：

```go
func min(a, b int64) int64 { ... }
func max(a, b int64) int64 { ... }
```

**问题**：Go 1.21 引入了内置 `min`/`max`，本项目使用 Go 1.24，自定义版本产生遮蔽（shadowing）。

**目标**：删除这两个函数，直接使用内置 `min`/`max`。

#### P8-b：`RateLimiter` 与 `limits.ConcurrencyLimit` 功能重叠评估

`limits.ConcurrencyLimit` 提供了简洁的并发限制（semaphore + queue + timeout）。`ratelimit.RateLimiter` 在此基础上增加了：

- 动态调整（pressure-based auto-scaling）
- 后台 goroutine（`adjustmentLoop`）
- 详细 metrics 回调

动态调整逻辑引入了持续运行的 goroutine，需要 `Stop()` 清理，这与 middleware 的无状态期望不符。

**目标**：

1. 将动态调整（`adjustmentLoop`、`analyzePressure`、`updateMaxConcurrent`）拆出为独立的 `ConcurrencyController` 类型，放到 `ops/` 或 `ratelimit/` 包的非 middleware 层
2. `RateLimiter.Middleware()` 退化为与 `limits.ConcurrencyLimit` 等价的简洁实现，或直接删除 `ratelimit.RateLimiterConfig` 路径，统一到 `limits.ConcurrencyLimit` + 外部 metrics hook

---

### P9 — `middleware/registry.go`：文档描述执行顺序有误

**当前文档**：

```
// Middleware execution order:
//   - Prepend() adds middleware to the beginning (executes first)
//   - Use() adds middleware to the end (executes last)
//   - When applied, middlewares execute in reverse order (last added runs first)
```

**问题**：`Chain.Apply` 实现是：

```go
for i := len(c.middlewares) - 1; i >= 0; i-- {
    h = c.middlewares[i](h)
}
```

从末尾向前包装，使得 `middlewares[0]` 成为最外层，即**注册顺序 = 执行顺序**（第一个注册的最先执行）。文档说的 "reverse order" 与实现相反。

**目标**：修正文档，删除 "When applied, middlewares execute in reverse order" 描述，改为：

```
// Execution order matches registration order: the first registered
// middleware executes outermost (runs first before the handler).
```

---

## 三、执行优先级

### 阶段一：删除（影响最小，收益最大）

| 任务 | 文件 | 理由 |
|---|---|---|
| 删除 `bind` 包 | `middleware/bind/` | 已自标注非 canonical |
| 删除 `WrapJSONResponse` | `middleware/transform/transform.go` | §18 明确禁止 |
| 删除 `cache` adapter | `middleware/cache/adapter.go` | 错位 re-export |
| 删除 `proxy` adapter | `middleware/proxy/adapter.go` | 错位 re-export |
| 删除 `circuitbreaker` adapter | `middleware/circuitbreaker/adapter.go` | 错位 re-export |
| 删除 `Auth`/`FromAuthMiddleware`/`AuthMiddleware`/`SimpleAuthMiddleware` | `middleware/auth/auth.go` | OOP 包装 |
| 删除 `min`/`max` helpers | `middleware/ratelimit/limiter.go` | stdlib 已提供 |
| 删除 `http.Pusher` 实现 | `middleware/observability/logging.go` | 弃用 API |

### 阶段二：移动（包边界修正）

| 任务 | 从 | 到 |
|---|---|---|
| 移动 `TenantPolicy` | `middleware/tenant/policy.go` | `tenant/middleware/policy.go` |
| 移动 `TenantQuota` | `middleware/tenant/quota.go` | `tenant/middleware/quota.go` |

### 阶段三：重构（类型与实现修正）

| 任务 | 文件 |
|---|---|
| 统一 `MetricsCollector` 接口 | `middleware/observability/logging.go` |
| `SimpleAuth(token string) middleware.Middleware` | `middleware/auth/auth.go` |
| 提取共享 `responseRecorder` | `middleware/internal/recorder/recorder.go` |
| 修正 `Middleware()`/`CustomExtractor()` 返回类型 | `middleware/versioning/version.go` |
| 修正 context key 类型 | `middleware/versioning/version.go` |
| 预编译正则 | `middleware/versioning/version.go` |
| 内化 `Config.WithDefaults()` | `middleware/versioning/version.go` |
| 拆分 `RateLimiter` 动态调整逻辑 | `middleware/ratelimit/limiter.go` |

### 阶段四：文档修正

| 任务 | 文件 |
|---|---|
| 修正执行顺序描述 | `middleware/registry.go` |

---

## 四、重构后目标结构

```
middleware/
  middleware.go              ← Chain, Apply（不变）
  registry.go                ← Registry（文档修正）
  error_registry.go          ← 错误码 + WriteTransportError（不变）
  conformance_test.go        ← 已有一致性测试
  layer_boundary_test.go     ← 已有层级边界测试

  auth/
    auth.go                  ← SimpleAuth(token) middleware.Middleware
    contract.go              ← Authenticate/Authorize/SessionCheck（不变）

  coalesce/coalesce.go       ← 不变
  compression/gzip.go        ← 不变
  conformance/               ← 不变
  cors/cors.go               ← 不变
  debug/debug_errors.go      ← 不变
  limits/limit.go            ← 不变

  observability/
    logging.go               ← 单一 MetricsCollector，无 http.Pusher
    request_id.go            ← 不变

  protocol/middleware.go     ← 不变
  ratelimit/
    abuse_guard.go           ← 不变
    token_bucket.go          ← 不变
    limiter.go               ← 删除 min/max，动态调整拆出

  recovery/recover.go        ← 不变
  security/headers.go        ← 不变

  tenant/
    resolver.go              ← TenantResolver（不变）
    ratelimit.go             ← TenantRateLimit（不变）
    helpers.go               ← 不变
    ← policy.go 已移走
    ← quota.go 已移走

  timeout/timeout.go         ← 不变

  transform/
    transform.go             ← 删除 WrapJSONResponse，使用共享 recorder

  versioning/version.go      ← 类型修正，预编译正则，私有 context key

  internal/
    recorder/
      recorder.go            ← 共享 responseRecorder（Hijack/Flush，无 Push）

  ← bind/ 已删除
  ← cache/ 已删除
  ← circuitbreaker/ 已删除
  ← proxy/ 已删除
```

---

## 五、质量门

每个阶段完成后必须通过：

```bash
go test -timeout 20s ./...
go vet ./...
gofmt -w .
go test -race ./...
```

conformance 测试（`conformance_test.go`、`layer_boundary_test.go`）必须全部通过，不得降级。

---

## 六、非目标

以下不在本次重构范围内：

- `contract/` 包（错误类型、写入辅助，已是 canonical）
- `core/` 包（bootstrap，已符合 §4）
- `router/` 包
- `ai/`、`ops/`、`net/` 等扩展包内部实现
- 新增功能、新增 middleware 种类
