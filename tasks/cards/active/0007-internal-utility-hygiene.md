# Card 0007

Priority: P1
State: active
Primary Module: internal
Owned Files:
  - internal/httputil/html.go
  - internal/httputil/http_response.go
  - internal/httpx/client_ip.go
  - x/gateway/websocket.go
  - internal/validator/validator.go

Depends On: —

Goal:
修复 internal 工具包中三类相互独立但同属"内部卫生"范畴的问题：
1. **包名不一致**：`internal/httputil` 目录下所有文件声明 `package utils`，与目录名不符，
   且 `utils` 是反模式包名。
2. **ClientIP 逻辑重复**：`internal/httpx.ClientIP` 已实现完整的 IP 提取逻辑，
   但 `x/gateway/websocket.go:241-262` 自定义了私有 `getClientIP`，逻辑相同却未复用。
3. **internal/httpx 缺少测试**：`internal/httpx/client_ip.go` 无任何测试文件，
   而该函数处理安全敏感的 header 解析逻辑。
4. **internal/validator/validator.go 过大**：单文件 2133 行，包含规则定义、结构体验证、
   HTTP binding、工具函数等不相关逻辑，难以导航和局部修改。

Scope:
- 将 `internal/httputil` 包名从 `package utils` 改为 `package httputil`，
  同步更新唯一调用方 `internal/nethttp/response_recorder.go`
- 将 `x/gateway/websocket.go:getClientIP` 替换为调用 `internal/httpx.ClientIP`
- 为 `internal/httpx/client_ip.go` 新增测试文件，覆盖：
  X-Forwarded-For（含多 IP 链）、X-Real-IP、RemoteAddr、nil request 四个分支
- 将 `internal/validator/validator.go` 按职责拆分为：
  - `validator.go`（核心类型：Rule、Validator、ValidationError、FieldErrors，~300 行）
  - `rules.go`（所有内置 Rule 实现：Required、Min、Max、Email 等）
  - `http.go`（BindJSON、ValidateRequest 等 HTTP binding 函数）
  拆分不得改变任何公共 API，仅是文件级重组

Non-goals:
- 不修改 internal/httputil 的功能逻辑
- 不改变 internal/validator 的公共接口
- 不将 internal/httpx 提升为公共包
- 不处理 internal/config/validator.go 的调用（调用的是本地 `validator` 变量，非包导入）

Files:
  - internal/httputil/html.go（改包名）
  - internal/httputil/http_response.go（改包名）
  - internal/httputil/html_test.go（改包名）
  - internal/httputil/http_response_test.go（改包名）
  - internal/nethttp/response_recorder.go（更新 import 别名或引用）
  - internal/httpx/client_ip.go
  - internal/httpx/client_ip_test.go（新建）
  - x/gateway/websocket.go（替换 getClientIP，引入 httpx 依赖）
  - internal/validator/validator.go（缩减为核心类型）
  - internal/validator/rules.go（新建，移入规则实现）
  - internal/validator/http.go（新建，移入 HTTP binding）

Tests:
  - go test ./internal/httputil/...
  - go test ./internal/httpx/...
  - go test ./internal/validator/...
  - go test ./x/gateway/...
  - go build ./internal/nethttp/...

Docs Sync: —

Done Definition:
- `internal/httputil` 内所有文件声明 `package httputil`，`go build` 通过
- `x/gateway/websocket.go` 中不再含私有 `getClientIP`，改用 `httpx.ClientIP`
- `internal/httpx/client_ip_test.go` 存在且 `go test ./internal/httpx/...` 通过
- `internal/validator/validator.go` ≤ 400 行，其余逻辑分布在新文件中
- 所有原有 validator 测试继续通过，无需修改

Outcome:
