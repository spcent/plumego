# Router 模块

`router` 包提供基于前缀树的匹配能力，支持路由分组、路径参数、通配路由与逆向路由，同时保持标准 `http.Handler` 兼容。

## Handler 形态
使用标准处理器：

- `Get`、`Post`、`Put`、`Patch`、`Delete`、`Options`、`Head`、`Any`
- 签名：`http.Handler`（或 `http.HandlerFunc`）

路径参数通过 `contract.Param(r, "id")` 读取。

## 路径模式
- 静态：`/docs/index.html`
- 参数：`/users/:id`、`/teams/:teamID/members/:id`
- 通配：`/*filepath`

## 分组与中间件顺序
中间件顺序显式：`全局 router.Use(...) -> group.Use(...) -> 路由处理器`。

```go
r := app.Router()
r.Use(observability.RequestID())

api := r.Group("/api")
api.Use(auth.SimpleAuth(os.Getenv("AUTH_TOKEN")))
api.Use(timeout.Timeout(2 * time.Second))

api.Get("/users/:id", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
    id, _ := contract.Param(req, "id")
    _, _ = w.Write([]byte("user=" + id))
}))
```

所有路由需在 `app.Boot()` 前注册；启动时会冻结路由表。

## 逆向路由
通过路由名安全生成 URL：

```go
if err := r.AddRouteWithOptions(router.GET, "/users/:id", http.HandlerFunc(showUser),
    router.WithRouteName("users.show"),
); err != nil {
    log.Fatal(err)
}

u := r.URL("users.show", "id", "42")
fmt.Println(u) // /users/42
```

## Method Not Allowed 行为
开启方法不匹配返回 405：

```go
app := core.New(core.WithMethodNotAllowed(true))
```

## 代码位置
- `router/registration.go`：路由/分组注册与校验。
- `router/dispatch.go`：匹配与分发。
- `router/metadata.go`：路由元信息与逆向路由。
- `router/reverse_routing_group_test.go`：分组与逆向路由边界测试。
