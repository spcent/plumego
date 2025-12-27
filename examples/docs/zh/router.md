# Router 模块

**router** 提供基于前缀树的高性能匹配，支持路由分组、命名参数与中间件继承。

## 职责
- 通过 `Get`、`Post`、`Put`、`Patch`、`Delete`、`Options`、`Head` 以及对应的 `*Ctx` 变体注册 Handler。
- 使用 `Group(prefix, ...middleware)` 复用路径前缀和中间件，保持边界清晰。
- 在 `App.Boot()` 阶段冻结路由表，阻止晚注册导致的不一致。

## 使用策略
- 优先采用 RESTful 路径，如 `GET /users/:id`、`POST /users`，便于文档与监控聚合。
- 命名参数用 `:name`，全局通配符用 `*filepath`（适合静态资源与多级文档）。
- 通过 `frontend.RegisterFromDir` 或嵌入式资源挂载前端到独立分组，隔离缓存与鉴权策略。

## 中间件继承
执行顺序为 **全局 → 分组 → 路由专属**。分组中追加的中间件会继承给子路由，适合区分公开与管理入口的安全策略。

## 调试
- 开启调试模式可在启动时输出 Method/Path 列表。
- 在开发环境添加遍历 `router.Routes()` 的路由（如 `/debug/routes`）便于审计最终路由树。
- `Boot()` 之后注册路由触发 panic，属于预期行为，务必调整初始化顺序而非吞掉异常。

## 示例
```go
api := app.Router().Group("/api")
api.Use(middleware.NewConcurrencyLimiter(100))
api.Get("/users/:id", userHandler)
api.Post("/users", createUserHandler)

assets := app.Router().Group("/docs")
assets.Get("/*filepath", docsHandler)
```

## 代码位置
- `router/router.go`：路由结构、节点匹配与分组逻辑。
- `frontend/register.go`：静态目录或嵌入式前端的挂载助手。
- `examples/reference/main.go`：API、指标、健康检查与前端路由的实战接线。
