# Error Handling

`contract` 使用 `APIError` 作为统一错误模型。

## APIError

```go
type APIError struct {
    Status   int            `json:"-"`
    Code     string         `json:"code"`
    Message  string         `json:"message"`
    Category ErrorCategory  `json:"category"`
    TraceID  string         `json:"trace_id,omitempty"`
    Details  map[string]any `json:"details,omitempty"`
}
```

HTTP 输出 envelope：

```json
{"error": {"code":"...","message":"...","category":"..."}}
```

## 写出错误

- `contract.WriteError(w, r, apiErr)`
- `contract.WriteBindError(w, r, err)`

## 构造错误

- Builder：`contract.NewErrorBuilder().Status(...).Code(...).Message(...).Build()`
- 快捷函数：
  - `NewValidationError`
  - `NewNotFoundError`
  - `NewUnauthorizedError`
  - `NewForbiddenError`
  - `NewTimeoutError`
  - `NewInternalError`
  - `NewRateLimitError`

## 分类

- `CategoryClient`
- `CategoryServer`
- `CategoryBusiness`
- `CategoryTimeout`
- `CategoryValidation`
- `CategoryAuthentication`
- `CategoryRateLimit`

## 错误工具

- `WrapError`, `WrapErrorf`
- `IsRetryable`
- `GetErrorDetails`, `FormatError`
- `PanicToError`
- `ErrorHandler`（集中处理日志 + response）
