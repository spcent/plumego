# Request Context (`contract.Ctx`)

`Ctx` 是对 `http.ResponseWriter` 和 `*http.Request` 的轻量封装，保持 `net/http` 兼容。

## 字段（核心）

- `W`, `R`
- `Params`, `Query`, `Headers`
- `ClientIP`, `TraceID`, `Deadline`
- `Logger`

## 路由与共享数据

- `ctx.Param(key) (string, bool)`
- `ctx.MustParam(key) (string, error)`
- `ctx.Set(key, val)` / `ctx.Get(key)` / `ctx.MustGet(key)`

## 响应

- `ctx.JSON(status, data)`
- `ctx.Text(status, text)`
- `ctx.Bytes(status, data)`
- `ctx.Response(status, data, meta)`
- `ctx.ErrorJSON(status, code, message, details)`
- `ctx.Redirect(status, location)`
- `ctx.SafeRedirect(status, location)`
- `ctx.File(path)`
- `ctx.SetCookie(...)` / `ctx.Cookie(name)`

## 绑定

- `ctx.BindJSON(dst)`
- `ctx.BindJSONWithOptions(dst, opts)`
- `ctx.BindAndValidateJSON(dst)`
- `ctx.BindAndValidateJSONWithOptions(dst, opts)`
- `ctx.BindQuery(dst)`
- `ctx.FormFile(name)` / `ctx.SaveUploadedFile(file, dst)`

## 流式输出

- `ctx.RespondWithSSE()` / `ctx.WriteSSE(event)`
- `ctx.StreamJSON(...)`, `ctx.StreamText(...)`, `ctx.StreamBinary(...)`
- channel / generator / chunked / retry 变体

## 生命周期

- `ctx.Abort()` / `ctx.AbortWithStatus(code)` / `ctx.IsAborted()`
- `ctx.Close()`
- `ctx.GetRequestDuration()` / `ctx.GetBodySize()` / `ctx.IsCompressed()`
