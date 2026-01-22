# Router 模块

**router** 提供基于前缀树的匹配，支持分组、路径参数和中间件继承，并保持与标准 `http.Handler` 兼容。

## Handler 形态
- 标准 Handler：`Get`、`Post`、`Put`、`Patch`、`Delete`、`Options`、`Head`、`Any` 接受 `http.Handler`。
- 函数辅助：`GetFunc`/`PostFunc` 等接受 `http.HandlerFunc`。
- 上下文 Handler：`GetCtx`/`PostCtx` 等接受 `contract.CtxHandlerFunc`，自动解析路径参数和请求信息。

## 路径模式
- 静态：`/docs/index.html`
- 参数：`/users/:id`、`/teams/:teamID/members/:id`
- 通配：`/*filepath`（捕获剩余路径，用于静态资源或文档）

## 分组与中间件顺序
中间件执行顺序为 **全局 → 分组 → 路由专属**。

```go
api := app.Router().Group("/api", middleware.SimpleAuth("token"))
api.Use(middleware.Timeout(2 * time.Second))

api.GetCtx("/users/:id", func(ctx *contract.Ctx) {
    id := ctx.Param("id")
    ctx.Text(http.StatusOK, "user="+id)
})

// 嵌套分组继承前缀与中间件。
v1 := api.Group("/v1")
v1.Get("/ping", func(w http.ResponseWriter, _ *http.Request) { w.Write([]byte("pong")) })
```

所有注册应在 `app.Boot()` 前完成；启动时路由会被冻结以避免遗漏。

## 静态前端与通配
为静态资源单独划分分组，方便隔离缓存策略或鉴权。

```go
assets := app.Router().Group("/docs")
assets.Get("/*filepath", func(w http.ResponseWriter, r *http.Request) {
    http.ServeFileFS(w, r, os.DirFS("./examples/docs"), path.Clean(r.URL.Path))
})
```

若使用嵌入资源，可利用 `frontend` 辅助：

```go
// 将嵌入的 SPA 或文档站点挂载到 "/"。
_ = frontend.RegisterFS(
    app.Router(),
    http.FS(staticFS),
    frontend.WithPrefix("/"),
    frontend.WithCacheControl("public, max-age=31536000"),
    frontend.WithIndexCacheControl("no-cache"),
    frontend.WithFallback(true),
)
```

## 调试技巧
- 打开 `core.WithDebug()` 可在启动时打印所有 method/path。
- 本地诊断冲突时，可在开发路由中遍历 `router.Routes()` 输出结果。
- 启动期的延迟注册 panic 属于预期行为，请调整初始化顺序而不是忽略。

## 代码位置
- `router/router.go`：前缀树匹配、分组与辅助函数。
- `frontend/frontend.go`：挂载磁盘目录或嵌入前端包的辅助方法。
- `examples/reference/main.go`：API、指标、健康检查、文档和前端路由的实际接线。
