# Contract Module

> Package: `github.com/spcent/plumego/contract`
> Stability: High

`contract` 定义 plumego 在 HTTP 处理链中的统一契约，重点包含：

- `Ctx` 请求上下文与响应辅助
- 结构化错误 `APIError`
- 请求绑定与校验辅助
- 流式输出（NDJSON / text / SSE）
- 协议适配契约（`contract/protocol`）

## 主要类型

- `type Ctx struct { ... }`
- `type CtxHandlerFunc func(*Ctx)`
- `type APIError struct { Code, Message, Category, TraceID, Details }`
- `type Response struct { Data, Meta, TraceID }`

## 常用入口

- `contract.NewCtx(w, r, params)`
- `contract.AdaptCtxHandler(handler, logger)`
- `contract.WriteError(w, r, apiErr)`
- `contract.WriteResponse(w, r, status, data, meta)`
- `contract.WriteBindError(w, r, err)`

## 与根包关系

`plumego.Context` 是 `contract.Ctx` 的别名。当前 v1 推荐保持标准 `net/http` handler 形态，并通过 `contract.AdaptCtxHandler(...)` 显式适配 `Ctx` 处理器。

## 相关文档

- [context.md](context.md)
- [errors.md](errors.md)
- [response.md](response.md)
- [request.md](request.md)
- [protocol/README.md](protocol/README.md)
