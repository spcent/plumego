# Card 0008

Priority: P2
State: active
Primary Module: middleware
Owned Files:
  - middleware/error_registry.go

Depends On: —

Goal:
`middleware/error_registry.go` 中定义的 `WriteTransportError` 和三个错误码常量
（CodeServerBusy、CodeServerQueueTimeout、CodeUpstreamFailed）位于父包 `middleware`，
导致所有子包（concurrencylimit、ratelimit、recovery、timeout、bodylimit、coalesce）
都必须导入父包，形成"子包反向依赖父包"的不对称耦合。
目前有 15+ 处调用，模式统一但依赖链不干净。

将 `WriteTransportError` 和相关错误码迁移到 `middleware/internal/transport` 子包，
使子包改为导入该 internal 包，父包也从该 internal 包 re-export（保持公共 API 不变）。

Scope:
- 新建 `middleware/internal/transport/transport.go`，迁入：
  - `WriteTransportError` 函数
  - `CodeServerBusy`、`CodeServerQueueTimeout`、`CodeUpstreamFailed` 常量
- `middleware/error_registry.go` 改为从 `middleware/internal/transport` re-export，
  保持对外 API 不变（现有调用 `mw.WriteTransportError`、`mw.CodeServerBusy` 等无需修改）
- 各子包改为直接从 `middleware/internal/transport` 导入（可选，若保持 re-export 则无需改动子包）
- 重点是消除子包对父包的隐式依赖，不引入新的对外破坏

Non-goals:
- 不改变 `WriteTransportError` 的签名或行为
- 不修改 x/gateway、x/rest 中使用 `mw.WriteTransportError` 的调用方
- 不统一 middleware 与 handler 层的错误写入路径（那是另一个独立讨论）
- 不迁移与错误码无关的 middleware 功能

Files:
  - middleware/error_registry.go（保留为 re-export shim 或删除，视决策）
  - middleware/internal/transport/transport.go（新建）
  - middleware/concurrencylimit/concurrency_limit.go（更新 import）
  - middleware/ratelimit/abuse_guard.go（更新 import）
  - middleware/recovery/recover.go（更新 import）
  - middleware/timeout/timeout.go（更新 import）
  - middleware/bodylimit/body_limit.go（更新 import）
  - middleware/coalesce/coalesce.go（更新 import）

Tests:
  - go build ./middleware/...
  - go test ./middleware/...

Docs Sync: —

Done Definition:
- `middleware/internal/transport/transport.go` 存在，含函数与常量
- `middleware/error_registry.go` 内容仅为 re-export 或删除（不含实现逻辑）
- 所有子包编译通过，无循环依赖
- `go test ./middleware/...` 通过

Outcome:
