# 可观测性契约

本文档定义 Plumego 的默认可观测性约定。
不依赖特定 tracing SDK，但统一上下文 key、header 和日志字段命名。

## Request ID / Trace ID
- 推荐请求头：`X-Request-ID`（不区分大小写）。
- 若存在，应写入上下文。
- 若不存在，中间件应生成并注入：
  - 响应头 `X-Request-ID`
  - Context key `contract.TraceIDKey{}`
  - `middleware.RequestID()` 提供该行为，无需 tracing SDK。

## Trace Context
- 优先使用 `contract.TraceContext`。
- 使用 `contract.TraceContextFromContext(ctx)` 读取。
- 若无 trace context，则回退到 `contract.TraceIDFromContext(ctx)`。

## Context 字段规范
- 只读标识：`trace_id`、`span_id`、`request_id`
- 请求元数据：method、path、status、duration、bytes
- 路由元信息（可用时）：`route`（pattern）和 `route_name`
- 安全字段不可记录 secrets 或原始 token。

## 日志字段命名建议
最小排障字段：
- `trace_id`
- `method`
- `path`
- `route`
- `status`
- `duration_ms`
- `bytes`

可选字段：
- `span_id`
- `request_id`
- `client_ip`
- `user_agent`
- `route_name`
- `error`
- `error_code`
- `error_type`
- `error_category`

## 错误映射
- 若使用 `contract.APIError`，建议映射：
  - `error_code` ← `APIError.Code`
  - `error_category` ← `APIError.Category`
  - `error_type` ← `APIError.Details["type"]`（若存在）
