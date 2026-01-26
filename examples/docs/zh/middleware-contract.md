# Middleware 稳定语义契约

本文档定义 Plumego 中间件的稳定语义，用于冻结行为，便于用户与智能体安全依赖。

## 范围
- 适用于 `middleware.Middleware`、`middleware.Apply` 与 `middleware.Chain`。
- 保持 `net/http` 兼容（`http.Handler` / `http.HandlerFunc`）。

## 执行顺序
- 中间件执行顺序与注册顺序一致。
- 示例：`Apply(h, A, B)` 的执行顺序为 `A -> B -> h`。
- `Chain.Use(A).Use(B)` 的执行顺序为 `A -> B -> h`。

## 短路行为
- 中间件可以不调用 `next.ServeHTTP` 以短路请求。
- 短路会阻止后续中间件和 handler 执行。
- 必须显式返回（例如认证失败、限流等）。

## Context 传播
- 必须保留 `context.Context`，允许附加请求级数据。
- 推荐使用 `contract` 中定义的标准 key（如 `ContextWithPrincipal`）。
- 除非文档声明，禁止使用全局可变状态。

## 响应行为
- 中间件调用 `next` 后不应再次写 header/body，除非明确负责响应。
- 避免重复写入或冲突状态码。

## 兼容性
- 中间件必须接收并返回 `http.Handler`。
- 不应假设特定类型的 `ResponseWriter`。
