# Response Helpers

## 标准成功响应

`contract.WriteResponse` / `ctx.Response` 输出结构：

```json
{
  "data": {...},
  "meta": {...},
  "trace_id": "..."
}
```

## 主要方法

- `contract.WriteJSON(w, status, payload)`
- `contract.WriteResponse(w, r, status, data, meta)`
- `ctx.JSON(status, data)`
- `ctx.Text(status, text)`
- `ctx.Bytes(status, data)`
- `ctx.File(path)`
- `ctx.Redirect(status, location)` / `ctx.SafeRedirect(status, location)`

## 流式响应

- `ctx.RespondWithSSE()` + `sseWriter.Write(event)`
- `ctx.StreamJSON(...)`
- `ctx.StreamText(...)`
- `ctx.StreamBinary(reader, chunkSize)`

`chunkSize <= 0` 会返回 `ErrInvalidChunkSize`。
