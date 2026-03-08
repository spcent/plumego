# Router 包重构计划

**状态**: 草案
**目标分支**: `claude/router-refactor-plan-S9ncp`
**依据**: `docs/CANONICAL_STYLE_GUIDE.md`
**范围**: `router/` 包全面重构，不考虑向后兼容性

---

## 一、诊断：当前问题清单

### 1.1 API 表面积爆炸（违反 §2.2、§8.1、§21）

当前 `router_api.go` 为每个 HTTP 方法暴露了**三套注册家族**：

```go
// 家族 A —— 标准 http.Handler
r.Get(path, handler http.Handler) error

// 家族 B —— contract.CtxHandlerFunc（非标准）
r.GetCtx(path, handler contract.CtxHandlerFunc) error

// 家族 C —— http.HandlerFunc 的"Func"别名
r.GetFunc(path, h http.HandlerFunc) error
```

加上通用变体：`AddRoute`、`Handle`、`HandleFunc`、`AddRouteWithOptions`、`HandleWithOptions` —— 共 **5 种**路由注册方式，全部是同一件事。

样式指南 §8.1 明确规定："只有**一种**可以是规范的"，§21 列为禁止模式："在同一个示例中混用 `Get`、`GetCtx`、`GetHandler` 风格"。

---

### 1.2 注册方法返回 `error`（设计问题）

```go
if err := r.Get("/users", handler); err != nil {
    log.Fatal(err)
}
```

`error` 只会在两种情况出现：路由器已冻结、路由重复注册。两种都是**编程错误**，应在启动时立即暴露，而不是要求每个注册调用都检查错误。这让所有注册代码变得冗余，背离了 §7.2 "一条路由，一行声明"的原则。

---

### 1.3 `RouteRegistrar` 接口引入隐式副作用（违反 §6.2、§21）

```go
type RouteRegistrar interface {
    Register(r *Router)
}
r.Register(&UserRoutes{})
```

这与 §7.2 "路由注册应保持 grep 友好" 和 §6.3 "避免隐式副作用驱动的启动行为"直接冲突。调用者无法在不跳转到 `UserRoutes.Register` 实现的情况下了解注册了哪些路由。

---

### 1.4 `resource.go` 包含已弃用的再导出（违反 §4.2）

`resource.go` 从 `rest` 包以类型别名形式再导出了超过 **20 个**类型（`Response`、`ErrorResponse`、`JSONWriter`、`ResourceController` 等），全部标记为 `Deprecated`。
§4.2 明确："`router` 不允许包含资源控制器"。由于不考虑兼容性，这些内容应全部删除。

`router_api.go` 中的 `Resource(path, controller)` 方法同样属于 REST 概念，不属于路由关注点。

---

### 1.5 验证器（Validator）违反包边界（违反 §4.2）

`validator.go` 当前包含：

- `ParamValidator` 接口、`RouteValidation` 结构体（路由关注点，**可以保留**）
- `RegexValidator`、`RangeValidator`、`LengthValidator`、`CompositeValidator`、`EnumValidator` 等验证器类型
- **16 个预定义验证器**：`UUIDValidator`、`EmailValidator`、`PhoneValidator`、`IPv4Validator`、`IPv6Validator`、`DateValidator`、`DateTimeValidator`、`SlugValidator`、`AlphanumericValidator` 等

`EmailValidator`、`PhoneValidator`、`IPv6Validator` 明显不是路由关注点。§4.2 明确列举："路由包中不允许包含验证器"。

---

### 1.6 `MiddlewareManager` 不必要地导出内部细节

`MiddlewareManager` 当前作为导出类型，暴露了 `GetMiddlewares()`、`MergeMiddlewares()`、`Version()` 等方法。外部调用者只需要 `Router.Use()`；中间件管理器本身是实现细节，不应成为公共 API 的一部分。

---

### 1.7 路由选项（RouteOption）API 混入文档元数据

```go
WithRouteName(name string) RouteOption      // 功能性：用于反向 URL 生成 ✓
WithRouteTags(tags ...string) RouteOption   // 文档元数据 ✗
WithRouteDescription(desc string) RouteOption // 文档元数据 ✗
WithRouteDeprecated(bool) RouteOption      // 文档元数据 ✗
```

路由注册代码不应承载文档/API 生成关注点。如果需要文档支持，应通过独立的文档层（如 OpenAPI 工具）处理，而不是嵌入路由注册选项。

---

### 1.8 Static 文件 API 有 4 个变体（违反 §2.2）

```go
r.Static(prefix, dir string)
r.StaticWithConfig(config StaticConfig)
r.StaticFS(prefix string, fs http.FileSystem)
r.StaticSPA(prefix, dir, indexFile string)
```

§2.2："每个常见任务应该有且只有一种推荐模式"。`StaticSPA` 的功能完全是 `StaticWithConfig{SPAFallback: true}` 的子集。

---

### 1.9 文件内部类型组织混乱

`router.go` 中混杂了：公共 `Router` 结构体、`MiddlewareManager`（应为内部细节）、`segment`（内部 trie 类型）、`node`（内部 trie 类型）、`routerState`（内部状态）。内部实现细节应与公共 API 明确分离。

---

## 二、重构原则

依据 `CANONICAL_STYLE_GUIDE.md` 第 4.2 节，`router` 包的职责边界：

**允许的职责：**
- 路由匹配（trie）
- 路由参数（`:param`、`*wildcard`）
- 路由分组（共享前缀 + 中间件）
- 路由树 / 查找
- 路由注册内部逻辑
- 静态文件挂载（保留）
- 参数验证**接口**（`ParamValidator`、`RouteValidation`）
- 命名路由（反向 URL 生成）

**不允许的职责：**
- REST 资源控制器
- 参数验证的具体实现（email、phone、UUID 等）
- JSON 写入器
- 业务响应封装
- 持久化抽象

---

## 三、分阶段重构计划

### Phase 1：删除非规范注册家族

**目标**：建立"唯一规范注册 API"

#### 1.1 删除所有 `*Ctx` 变体

```
删除: GetCtx, PostCtx, PutCtx, DeleteCtx, PatchCtx, AnyCtx
删除: addCtxRoute 内部方法
```

这些方法包装了 `contract.CtxHandlerFunc`，属于非规范处理器形状。如果需要 Ctx 风格处理器，由 `contract` 包提供适配器，由用户显式调用：

```go
// 替代方式：用户在注册时显式适配
r.Get("/users", contract.AdaptCtxHandler(myCtxHandler, logger))
```

#### 1.2 删除所有 `*Func` 变体

```
删除: GetFunc, PostFunc, PutFunc, DeleteFunc, PatchFunc, AnyFunc
```

`http.HandlerFunc` 本身就实现了 `http.Handler`，无需单独的 Func 变体：

```go
// 删除前
r.GetFunc("/users", func(w http.ResponseWriter, r *http.Request) { ... })

// 删除后（http.HandlerFunc 直接满足 http.Handler）
r.Get("/users", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { ... }))

// 或者更简洁（因为 func 字面量可以直接传给接受 http.Handler 的方法）
r.Get("/users", http.HandlerFunc(myHandler))
```

#### 1.3 删除 `Handle` 和 `HandleFunc` 同义方法

```
删除: Handle(method, path, http.Handler)
删除: HandleFunc(method, path, http.HandlerFunc)
删除: HandleWithOptions(method, path, http.Handler, ...RouteOption)
```

保留 `AddRoute(method, path, handler)` 作为底层注册入口（适合在需要动态指定 method 的场景下使用）。

**重构后的规范注册 API**：

```go
// 规范 HTTP 方法快捷方式
r.Get(path string, handler http.Handler)
r.Post(path string, handler http.Handler)
r.Put(path string, handler http.Handler)
r.Delete(path string, handler http.Handler)
r.Patch(path string, handler http.Handler)
r.Options(path string, handler http.Handler)
r.Head(path string, handler http.Handler)
r.Any(path string, handler http.Handler)

// 通用方法（用于动态 method 或带元数据注册）
r.AddRoute(method, path string, handler http.Handler)
r.AddRouteWithName(method, path, name string, handler http.Handler)
```

---

### Phase 2：改变注册方法签名 —— 去除 `error` 返回

**目标**：让路由注册代码干净可读

**理由**：重复注册和冻结后注册都是程序错误（programming errors），应在启动时立即 panic，而不是让调用者处理错误。这与 `net/http` 的 `http.Handle`（重复注册直接 panic）一致。

#### 2.1 修改所有注册方法的签名

```go
// 重构前
func (r *Router) Get(path string, handler Handler) error
func (r *Router) Post(path string, handler Handler) error
// ...

// 重构后
func (r *Router) Get(path string, handler http.Handler)
func (r *Router) Post(path string, handler http.Handler)
// ...
```

#### 2.2 修改 `AddRoute` 内部实现

```go
// 重构前
func (r *Router) AddRoute(method, path string, handler Handler) error {
    // ...
    return contract.WrapError(...)
}

// 重构后
func (r *Router) AddRoute(method, path string, handler http.Handler) {
    if err := r.addRoute(method, path, handler); err != nil {
        panic(fmt.Sprintf("router: %v", err))
    }
}

// 内部实现保留 error 返回用于逻辑复用
func (r *Router) addRoute(method, path string, handler http.Handler) error {
    // ...（原有逻辑不变）
}
```

#### 2.3 注册代码变为真正可读

```go
// 重构后 —— 无错误处理噪音
func registerRoutes(app *core.App) {
    app.Get("/healthz", healthHandler)
    app.Post("/users", createUserHandler)
    app.Get("/users/:id", getUserHandler)
}
```

---

### Phase 3：删除 `RouteRegistrar` 接口

**目标**：消除隐式路由注册，保持 grep 友好性

```
删除: RouteRegistrar 接口定义
删除: Router.Register(registrars ...RouteRegistrar) 方法
删除: routerState.registrars 字段
```

**替代模式**（规范方式，显式传参）：

```go
// 重构前（隐式注册）
r.Register(&UserRoutes{})

// 重构后（显式注册，grep 友好）
func registerUserRoutes(r *router.Router, svc user.Service) {
    h := handlers.UserHandler{Service: svc}
    r.Get("/users", h.List)
    r.Post("/users", h.Create)
    r.Get("/users/:id", h.Show)
    r.Put("/users/:id", h.Update)
    r.Delete("/users/:id", h.Delete)
}
```

---

### Phase 4：删除 `resource.go` 及 `Resource()` 方法

**目标**：清除 REST 资源控制器概念，符合包边界

```
删除: router/resource.go 整个文件
删除: Router.Resource(path, controller) 方法（router_api.go）
```

这些内容属于 `rest` 包，不属于 `router` 包。

**影响**：调用 `r.Resource(...)` 的代码需迁移到 `rest` 包提供的独立辅助函数。

---

### Phase 5：将预定义验证器迁移到新的 `routeparam` 包

**目标**：保持 `router` 包的职责纯粹

#### 5.1 在 `router` 包中保留（路由关注点）

```go
// 保留：参数验证接口
type ParamValidator interface {
    Validate(name, value string) error
}

// 保留：路由验证规则容器
type RouteValidation struct {
    Params map[string]ParamValidator
}

// 保留：RouteValidation 的方法
func NewRouteValidation() *RouteValidation
func (rv *RouteValidation) AddParam(name string, v ParamValidator) *RouteValidation
func (rv *RouteValidation) Validate(params map[string]string) error

// 保留：WithValidation、WithValidationRule RouterOption
// 保留：Router.AddValidation 方法
```

#### 5.2 迁移到新的 `routeparam` 包

创建 `routeparam/validators.go`，包含：

```go
package routeparam

import "github.com/spcent/plumego/router"

// 迁移具体验证器类型
type RegexValidator struct { ... }
type RangeValidator struct { ... }
type LengthValidator struct { ... }
type CompositeValidator struct { ... }
type EnumValidator struct { ... }
type NotEmptyValidator struct { ... }
type CustomValidator struct { ... }

// 迁移预定义验证器
var (
    UUID         = &RegexValidator{ ... }
    Email        = &RegexValidator{ ... }
    Phone        = &RegexValidator{ ... }
    PositiveInt  = &RegexValidator{ ... }
    Slug         = &RegexValidator{ ... }
    Alphanumeric = &RegexValidator{ ... }
    // ...
)

// 构造函数
func Regex(pattern string) (router.ParamValidator, error)
func Range(min, max int64) router.ParamValidator
func Length(min, max int) router.ParamValidator
func Enum(values ...string) router.ParamValidator
func Custom(fn func(name, value string) error) router.ParamValidator
```

用法示例：

```go
import (
    "github.com/spcent/plumego/router"
    "github.com/spcent/plumego/routeparam"
)

r.AddValidation("GET", "/users/:id", router.NewRouteValidation().
    AddParam("id", routeparam.UUID))
```

#### 5.3 从 `validator.go` 删除

```
删除: RegexValidator, RangeValidator, LengthValidator
删除: CompositeValidator, EnumValidator, NotEmptyValidator, CustomValidator
删除: 所有 16 个预定义验证器变量
删除: NewRegexValidator, NewRangeValidator, NewLengthValidator 等构造函数
删除: NewCompositeValidator, NewCustomValidator, NewCustomValidatorWithMessage
```

---

### Phase 6：收敛路由选项，删除文档元数据

**目标**：路由选项只承载功能性路由关注点

#### 6.1 删除文档元数据选项

```
删除: WithRouteTags(tags ...string) RouteOption
删除: WithRouteDescription(desc string) RouteOption
删除: WithRouteDeprecated(bool) RouteOption
```

#### 6.2 删除 `RouteMeta` 中的文档字段

```go
// 重构前
type RouteMeta struct {
    Name        string
    Tags        []string   // 删除
    Description string     // 删除
    Deprecated  bool       // 删除
}

// 重构后
type RouteMeta struct {
    Name string
}
```

#### 6.3 简化路由命名 API

提供更直接的命名方式，而不是通过 `AddRouteWithOptions`：

```go
// 重构后：可选的命名在注册时完成
r.AddRouteWithName("GET", "/users/:id", "user.show", getUserHandler)

// 或保留 WithRouteName 作为选项（但仅此一个选项）
r.AddRoute("GET", "/users/:id", getUserHandler, router.WithRouteName("user.show"))
```

**注意**：`WithRouteName` 保留，因为命名路由支持反向 URL 生成（`r.URL("user.show", "id", "123")`），这是合法的路由关注点。

---

### Phase 7：简化 Static 文件 API

**目标**：两种方式（目录 / FileSystem），去除 SPA 专用变体

#### 7.1 删除 `StaticSPA`

```
删除: StaticSPA(prefix, dir, indexFile string) 方法
```

`StaticSPA` 的功能完全可以通过 `StaticWithConfig` 表达：

```go
// 重构前
r.StaticSPA("/", "./dist", "index.html")

// 重构后（等价，显式）
r.StaticWithConfig(router.StaticConfig{
    Prefix:      "/",
    Root:        "./dist",
    IndexFile:   "index.html",
    SPAFallback: true,
})
```

#### 7.2 保留 3 个 Static 变体

```go
// 保留：目录快捷方式
r.Static(prefix, dir string)

// 保留：FileSystem（embed.FS 等）
r.StaticFS(prefix string, fs http.FileSystem)

// 保留：全配置（覆盖所有场景包括 SPA）
r.StaticWithConfig(config StaticConfig)
```

---

### Phase 8：将 `MiddlewareManager` 设为内部类型

**目标**：外部用户只需要 `Router.Use()`，不应直接操作 MiddlewareManager

#### 8.1 将 `MiddlewareManager` 重命名为 `middlewareManager`（unexported）

```go
// 重构前
type MiddlewareManager struct { ... }
func NewMiddlewareManager() *MiddlewareManager { ... }
func (mm *MiddlewareManager) AddMiddleware(m middleware.Middleware) { ... }
func (mm *MiddlewareManager) GetMiddlewares() []middleware.Middleware { ... }
func (mm *MiddlewareManager) MergeMiddlewares(other *MiddlewareManager) []middleware.Middleware { ... }
func (mm *MiddlewareManager) Version() uint64 { ... }

// 重构后（全部 unexported）
type middlewareManager struct { ... }
func newMiddlewareManager() *middlewareManager { ... }
// 方法均为 unexported（内部使用）
```

`Router.Use()` 是唯一对外暴露的中间件 API，保持不变。

---

### Phase 9：内部文件组织重构

**目标**：将内部实现细节与公共 API 明确分离

#### 9.1 文件职责重新划分

| 文件 | 职责 |
|------|------|
| `router.go` | `Router` 公共结构体 + `NewRouter` + 生命周期方法（`Use`、`Freeze`、`SetLogger` 等） |
| `trie.go` | `node`、`segment`、`routerState` 内部类型 + trie 操作（`insertChild`、`findChild` 等）|
| `registration.go` | `AddRoute`、`Group`、`GroupFunc`、`compilePathSegments` |
| `dispatch.go` | `ServeHTTP`、`matchRoute`、`applyMiddlewareAndServe`（保持现有结构）|
| `api.go` | `Get`、`Post`、`Put`、`Delete`、`Patch`、`Any`、`Options`、`Head` 快捷方式 |
| `metadata.go` | `SetRouteMeta`、`URL`、`URLMust`、`Routes`、`Print`、`NamedRoutes` |
| `static.go` | `Static`、`StaticFS`、`StaticWithConfig`、内部文件处理函数 |
| `validation.go` | `ParamValidator`、`RouteValidation`、`WithValidation`、`AddValidation` |
| `cache.go` | `RouteCache`、`CacheStats`（保持现有结构）|
| `pool.go` | 对象池（保持现有结构）|
| `matcher.go` | `RouteMatcher`（保持现有结构）|

#### 9.2 当前文件对应关系

```
router.go          → 拆分 → router.go + trie.go
router_api.go      → api.go（删除 Ctx/Func/Resource 变体后）
router_registration.go → registration.go
router_dispatch.go → dispatch.go（基本保持）
router_metadata.go → metadata.go（基本保持）
router_trie.go     → 合并到 trie.go
types.go           → 保持（删除 RouteMeta 文档字段）
resource.go        → 完全删除
validator.go       → validation.go（保留接口，删除实现）
static.go          → 保持（删除 StaticSPA）
cache.go           → 保持
pool.go            → 保持
matcher.go         → 保持
```

---

### Phase 10：测试同步

**目标**：测试代码同步到规范风格

#### 10.1 删除涉及已删除 API 的测试

```
删除或重写: 所有测试 GetCtx/PostCtx 等的用例
删除或重写: 所有测试 GetFunc/PostFunc 等的用例
删除或重写: 所有测试 RouteRegistrar/Register 的用例
删除或重写: 所有测试 Resource() 的用例（resource_test.go）
删除或重写: 所有测试 StaticSPA 的用例
```

#### 10.2 测试规范风格

所有测试统一采用 `httptest` + 标准断言模式（§15）：

```go
func TestGetUser(t *testing.T) {
    r := router.NewRouter()
    r.Get("/users/:id", getUserHandler)

    req := httptest.NewRequest(http.MethodGet, "/users/123", nil)
    rec := httptest.NewRecorder()
    r.ServeHTTP(rec, req)

    if rec.Code != http.StatusOK {
        t.Fatalf("expected 200, got %d", rec.Code)
    }
}
```

#### 10.3 更新 benchmark 测试

`router_bench_test.go` 和 `performance_test.go` 需移除对已删除 API 的引用。

---

## 四、完整变更列表（按文件）

### 删除的文件

| 文件 | 原因 |
|------|------|
| `router/resource.go` | 全部是 `rest` 包的弃用再导出，违反包边界 |

### 删除的导出符号

| 符号 | 文件 | 原因 |
|------|------|------|
| `GetCtx` / `PostCtx` / `PutCtx` / `DeleteCtx` / `PatchCtx` / `AnyCtx` | `router_api.go` | 非规范处理器形状 |
| `GetFunc` / `PostFunc` / `PutFunc` / `DeleteFunc` / `PatchFunc` / `AnyFunc` | `router_api.go` | `http.HandlerFunc` 已满足 `http.Handler` |
| `Handle` | `router_registration.go` | `AddRoute` 的同义词 |
| `HandleFunc` | `router_registration.go` | `AddRoute` 的同义词 |
| `HandleWithOptions` | `router_registration.go` | `AddRouteWithOptions` 的同义词 |
| `Resource(path, controller)` | `router_api.go` | 属于 `rest` 包 |
| `RouteRegistrar` 接口 | `router.go` | 引入隐式路由注册 |
| `Register(registrars ...RouteRegistrar)` | `router.go` | 引入隐式路由注册 |
| `MiddlewareManager` | `router.go` | 设为内部类型（unexported）|
| `NewMiddlewareManager` | `router.go` | 随 MiddlewareManager 一同 unexported |
| `WithRouteTags` | `router.go` | 文档元数据，不属于路由关注点 |
| `WithRouteDescription` | `router.go` | 文档元数据，不属于路由关注点 |
| `WithRouteDeprecated` | `router.go` | 文档元数据，不属于路由关注点 |
| `RouteMeta.Tags` 字段 | `types.go` | 文档元数据 |
| `RouteMeta.Description` 字段 | `types.go` | 文档元数据 |
| `RouteMeta.Deprecated` 字段 | `types.go` | 文档元数据 |
| `StaticSPA` | `static.go` | `StaticWithConfig` 的子集 |
| `RegexValidator` / `RangeValidator` / `LengthValidator` | `validator.go` | 迁移到 `routeparam` 包 |
| `CompositeValidator` / `EnumValidator` / `NotEmptyValidator` / `CustomValidator` | `validator.go` | 迁移到 `routeparam` 包 |
| 所有 16 个预定义验证器变量 | `validator.go` | 迁移到 `routeparam` 包 |
| `NewCompositeValidator` / `NewCustomValidator` / `NewCustomValidatorWithMessage` 等 | `validator.go` | 迁移到 `routeparam` 包 |

### 修改的签名

| 符号 | 修改前 | 修改后 |
|------|--------|--------|
| `Get` / `Post` / `Put` / `Delete` / `Patch` / `Any` / `Options` / `Head` | 返回 `error` | 无返回值（panic on error）|
| `AddRoute` | 返回 `error`，`Handler` 参数 | 无返回值，`http.Handler` 参数 |
| `AddRouteWithOptions` | 返回 `error` | 无返回值 |
| `RouteMeta` | 含 `Tags`、`Description`、`Deprecated` | 仅含 `Name` |

### 新增文件

| 文件 | 内容 |
|------|------|
| `routeparam/validators.go` | 迁移自 `validator.go` 的具体验证器类型和预定义变量 |

### 保持不变的核心

| 组件 | 说明 |
|------|------|
| Trie 匹配算法 | 核心路由逻辑保持不变 |
| LRU 路由缓存 | `cache.go` 保持不变 |
| 对象池 | `pool.go` 保持不变 |
| `ServeHTTP` 分发逻辑 | `dispatch.go` 保持不变（移除 Ctx 分支）|
| `Group` / `GroupFunc` | 分组机制保持不变 |
| 路由元数据（名称）| `metadata.go` 保持不变（删除文档字段）|
| `ParamValidator` 接口 | `validation.go` 保持不变 |
| `RouteValidation` | `validation.go` 保持不变 |
| `WithValidation` / `WithValidationRule` | `validation.go` 保持不变 |
| `Static` / `StaticFS` / `StaticWithConfig` | `static.go` 保持（删除 SPA 变体）|
| `Freeze` / `SetLogger` / `SetMethodNotAllowed` | `router.go` 保持不变 |
| `CacheStats` / `ClearCache` / `CacheSize` | `router.go` 保持不变 |
| `Routes` / `URL` / `URLMust` / `NamedRoutes` / `Print` | `metadata.go` 保持不变 |

---

## 五、重构后的规范使用示例

### 5.1 应用启动

```go
package main

import (
    "log"
    "net/http"

    "github.com/spcent/plumego/core"
)

func main() {
    app := core.New()

    app.Use(RequestID())
    app.Use(Recovery())
    app.Use(RequestLogger())

    registerRoutes(app)

    if err := http.ListenAndServe(":8080", app); err != nil {
        log.Fatal(err)
    }
}
```

### 5.2 路由注册（规范风格，无错误处理噪音）

```go
func registerRoutes(app *core.App, userSvc user.Service) {
    h := handlers.UserHandler{Service: userSvc}

    app.Get("/healthz", healthHandler)
    app.Get("/users", h.List)
    app.Post("/users", h.Create)
    app.Get("/users/:id", h.Show)
    app.Put("/users/:id", h.Update)
    app.Delete("/users/:id", h.Delete)
}
```

### 5.3 路由分组

```go
api := app.Group("/api/v1")
api.Use(AuthMiddleware())

api.Get("/users", h.List)
api.Post("/users", h.Create)
```

### 5.4 命名路由

```go
app.AddRouteWithName("GET", "/users/:id", "user.show", h.Show)

// 反向 URL 生成
url := r.URL("user.show", "id", "123") // → "/users/123"
```

### 5.5 路由参数验证

```go
import (
    "github.com/spcent/plumego/router"
    "github.com/spcent/plumego/routeparam"
)

r.AddValidation("GET", "/users/:id", router.NewRouteValidation().
    AddParam("id", routeparam.UUID))
```

### 5.6 静态文件

```go
// 目录
r.Static("/static", "./public")

// SPA（显式配置）
r.StaticWithConfig(router.StaticConfig{
    Prefix:      "/",
    Root:        "./dist",
    IndexFile:   "index.html",
    SPAFallback: true,
})

// embed.FS
r.StaticFS("/assets", http.FS(embeddedFiles))
```

---

## 六、实施检查清单

### 执行前提

- [ ] 当前所有测试通过：`go test -timeout 20s ./...`
- [ ] 当前 vet 清洁：`go vet ./...`

### Phase 1 完成标志

- [ ] `router_api.go` 中不存在任何 `Ctx` 或 `Func` 后缀方法
- [ ] `Handle` 和 `HandleFunc` 已从 `router_registration.go` 删除
- [ ] `resource.go` 文件已删除
- [ ] `router_api.go` 中 `Resource()` 方法已删除
- [ ] `resource_test.go` 已删除或重写

### Phase 2 完成标志

- [ ] `Get`、`Post` 等方法签名无 `error` 返回
- [ ] `AddRoute` 签名无 `error` 返回，内部 panic
- [ ] `core/routing.go` 相应更新
- [ ] 所有相关测试通过

### Phase 3 完成标志

- [ ] `RouteRegistrar` 接口已删除
- [ ] `Router.Register()` 方法已删除
- [ ] `routerState.registrars` 字段已删除

### Phase 4 完成标志

- [ ] `MiddlewareManager` 已 unexported 为 `middlewareManager`
- [ ] `NewMiddlewareManager` 已 unexported 为 `newMiddlewareManager`
- [ ] 外部调用者只通过 `router.Use()` 操作中间件

### Phase 5 完成标志

- [ ] `routeparam` 包已创建，包含所有具体验证器
- [ ] `validator.go` 中只保留 `ParamValidator` 接口和 `RouteValidation`
- [ ] `routeparam` 包有对应测试

### Phase 6 完成标志

- [ ] `WithRouteTags`、`WithRouteDescription`、`WithRouteDeprecated` 已删除
- [ ] `RouteMeta` 只含 `Name` 字段
- [ ] `RouteInfo.Meta` 相应简化

### Phase 7 完成标志

- [ ] `StaticSPA` 已删除
- [ ] `static_test.go` 中 SPA 测试改用 `StaticWithConfig`

### Phase 8 完成标志

- [ ] `go test -timeout 20s ./...` 全部通过
- [ ] `go vet ./...` 无告警
- [ ] `go test -race ./...` 无竞态

---

## 七、质量门控

每个 Phase 完成后必须执行：

```bash
go test -timeout 20s ./...
go vet ./...
gofmt -w .
```

全部 Phase 完成后额外执行：

```bash
go test -race ./...
```

---

## 八、不在本次范围内的内容

以下内容经评估后**不在本次重构范围**：

1. **`contract` 包变更** —— `contract.CtxHandlerFunc` 和 `contract.AdaptCtxHandler` 保持不变，只是不再由 `router` 包直接使用
2. **`core` 包大规模改动** —— `core/routing.go` 只需同步更新签名（去除 error 返回）
3. **`rest` 包** —— 不受影响，`resource.go` 的删除只影响 `router` 包的再导出层
4. **trie 匹配算法优化** —— 当前算法正确且经过充分测试，不在本次范围
5. **缓存策略变更** —— 当前 LRU 策略合理，不在本次范围
6. **新增路由功能** —— 本次只做收敛，不增加新能力
