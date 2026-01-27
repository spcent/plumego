# Router 稳定语义契约

本文档定义 Plumego Router 的稳定语义，用于冻结行为，便于用户与智能体安全依赖。

## 范围
- 适用于 `router.Router` 的匹配与分发语义。
- 保持与 `net/http` 的兼容。
- 不定义业务逻辑或中间件实现细节（仅定义顺序）。

## 路由匹配顺序
- 段优先级：静态段 > 参数段（`:id`）> 通配段（`*path`）。
- 自左向右逐段匹配。
- 通配段捕获剩余路径并终止继续匹配。

## 方法分发规则
- 优先选择请求方法对应的路由树。
- 若该方法树不存在，回退到 `ANY` 树。
- 若方法树未命中，会再次尝试 `ANY` 树。
- 仍未命中则返回 `404 Not Found`。
- `405 Method Not Allowed` 默认关闭，可通过 `router.WithMethodNotAllowed(true)` 启用并返回 `Allow`。

## 路径归一化
- 使用 `req.URL.Path` 原始值（不做 URL 解码）。
- 根路径固定为 `/`。
- 非根路径会裁剪前后 `/`，因此 `/a` 与 `/a/` 等价。
- 内部重复 `/` 不会被归一化（例如 `/a//b` 按原样匹配，通常无法命中 `/a/:id`）。

## 参数提取
- 参数值来自 `req.URL.Path`（Router 不做额外解码）。
- `net/http` 会对 `URL.Path` 做百分号解码（例如 `%20` 会变成空格）。
- 参数 key 按路由定义顺序生成。
- 若出现重复参数 key，后者会覆盖前者。
- 通配参数包含 `/`（例如 `/files/*path` → `path = "a/b/c.txt"`）。
- `%2F` 会被 `net/http` 解码为 `/`，因此单段参数不会命中；通配参数会捕获剩余路径（`a/b`）。
- 参数注入到请求上下文：
  - `contract.ParamsContextKey`（params map）
  - `contract.RequestContextKey`（RequestContext 中含 Params）
- Router 参数会覆盖请求上下文中已有的同名参数。
- 命中路由后，Router 还会写入：
  - `RequestContext.RoutePattern`（命中的路由模式）
  - `RequestContext.RouteName`（如果设置）

## 缓存行为
- Router 使用“规范化后的 path”作为缓存 key（非根路径会去掉多余斜杠与末尾斜杠）。
- 缓存 key 会包含请求 method 与 host。
- 当 middleware 注册发生变化时，缓存会失效（middleware version 增长）。
- 缓存仅是性能优化，不改变匹配语义。

## 中间件顺序（Router 级）
- 全局中间件（`router.Use`）先执行。
- 组级中间件从外到内执行，且按注册顺序执行。
- 最后执行路由 handler。

## 路由冲突规则（注册期）
- 同一 method+path 重复注册会报错。
- 同一路径位置不同参数名视为冲突。
- 同一父节点仅允许一个通配段。
