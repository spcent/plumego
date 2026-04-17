# Card 0016

Priority: P2
State: active
Primary Module: x/gateway
Owned Files:
  - x/gateway/transform/transform.go
  - x/websocket/security.go

Depends On: 0007（internal/nethttp 包名修复后可放心引用）

Goal:
两个扩展包中积累了小型但具体的技术债，归为一张卡集中处理：

**问题 1：x/gateway/transform 私有 responseRecorder 与已有工具重复**

`x/gateway/transform/transform.go:144-169` 定义了私有 `responseRecorder` 结构体，
实现 `Header()`、`WriteHeader()`、`Write()` 三个方法以捕获响应体。

而同一包 `x/gateway` 的 `cache/http_cache.go:227` 已经在使用
`nethttp.NewResponseRecorder(w)`（即 `internal/nethttp.ResponseRecorder`）做同样的事。

同包内两条路径并存，且 `internal/nethttp.ResponseRecorder` 功能更完整（含写透、
EnsureNoSniff 安全头等）。transform.go 应直接复用已有工具。

**问题 2：x/websocket SecurityMetrics 保留已废弃字段**

`x/websocket/security.go` 的 `SecurityMetrics` 结构体含 3 个字段，注释明确写明
"no longer tracked here"：
- `InvalidWebSocketKeys`（推荐用 `Hub.Metrics().InvalidWSKeys`）
- `BroadcastQueueFull`（推荐用 `Hub.Metrics().BroadcastDropped`）
- `RejectedConnections`（推荐用 `Hub.Metrics().SecurityRejections`）

这 3 个字段永远为零值，存在会误导使用者以为它们有效。按照风格指南
"Deprecated symbols must be removed in the same PR"，应当删除。

Scope:
- **transform.go**：删除私有 `responseRecorder` 类型及其 3 个方法，
  改为导入 `internal/nethttp` 并使用 `nethttp.NewResponseRecorder(w)`；
  验证捕获语义一致（body、status code、header）
- **security.go**：删除 `InvalidWebSocketKeys`、`BroadcastQueueFull`、
  `RejectedConnections` 三个字段；
  执行前：`grep -rn "InvalidWebSocketKeys\|BroadcastQueueFull\|RejectedConnections" . --include="*.go"`
  确认无外部赋值或读取（仅 security.go 内定义则可安全删除）

Non-goals:
- 不修改 transform 包的对外 API（transform.go 导出函数签名不变）
- 不改变 SecurityMetrics 其余字段（InvalidJWTSecrets 等仍有效）
- 不修改 Hub.Metrics() 返回类型

Files:
  - x/gateway/transform/transform.go（删除 responseRecorder，改用 nethttp）
  - x/websocket/security.go（删除 3 个废弃字段）
  - x/websocket/security_test.go（若有字段断言则同步更新）

Tests:
  - go build ./x/gateway/...
  - go test ./x/gateway/...
  - go build ./x/websocket/...
  - go test ./x/websocket/...

Docs Sync: —

Done Definition:
- `grep -n "responseRecorder" x/gateway/transform/transform.go` 结果为空
- `grep -n "InvalidWebSocketKeys\|BroadcastQueueFull\|RejectedConnections" x/websocket/security.go` 结果为空
- 所有被修改包编译通过，测试通过

Outcome:
