# Card 0012

Priority: P2
State: active
Primary Module: x/pubsub
Owned Files:
  - x/pubsub/pubsub.go
  - x/pubsub/ack.go

Depends On: —

Goal:
`x/pubsub.InProcBroker` 的公共 API 在 context 处理上严重不一致：

**发布端（Publish 侧）**：
- `Publish(topic, msg)` — 无 context
- `PublishWithContext(ctx, topic, msg)` — 有 context（同一操作的两个变体同时存在）
- `PublishAsync(topic, msg)` — 无 context
- `PublishBatch(topic, msgs)` — 无 context

**订阅端（Subscribe 侧）**：
- `Subscribe(topic, opts)` — 无 context
- `SubscribeAckable(topic, subOpts, ackOpts)` — 无 context
- `SubscribeOrdered(ctx, topic, opts)` — 有 context

API 使用者无法形成统一心智模型：订阅时何时需要 ctx？发布时应用哪个变体？
`Publish` + `PublishWithContext` 并存是明显的 API 表面噪音。

Scope:
- **发布端统一**：将 `Publish(topic, msg)` 改为首选 `Publish(ctx, topic, msg)`，
  内部 `PublishWithContext` 变为私有或删除（保持向后兼容的过渡期内保留为 deprecated 包装）；
  `PublishAsync` 和 `PublishBatch` 同步补充 ctx 参数
- **订阅端补充 ctx**：`Subscribe(topic, opts)` 和 `SubscribeAckable(topic, subOpts, ackOpts)`
  补充 ctx 参数，用于控制订阅生命周期（ctx cancel 触发订阅关闭，与 SubscribeOrdered 一致）
- 更新所有内部调用和测试
- 若改动影响面过大（调用方超过 20 处），可分两步：本卡先完成 Subscribe/SubscribeAckable
  补 ctx，发布端单独一卡

Non-goals:
- 不改变消息投递语义（at-most-once / at-least-once）
- 不修改 Message、SubOptions、AckOptions 等数据结构
- 不影响 RateLimitedPubSub / DistributedPubSub 等包装类型的行为（跟随主类型调整签名即可）

Files:
  - x/pubsub/pubsub.go（Publish 系列签名）
  - x/pubsub/ack.go（SubscribeAckable 补 ctx）
  - x/pubsub/ratelimit.go（跟随调整）
  - x/pubsub/distributed.go（跟随调整）
  - x/pubsub/*_test.go（更新调用方）

Tests:
  - go test -race ./x/pubsub/...

Docs Sync: —

Done Definition:
- `Subscribe` 和 `SubscribeAckable` 均接受 `ctx context.Context` 作为第一参数
- `Publish` 与 `PublishWithContext` 二选一（推荐 Publish 带 ctx，旧变体标 Deprecated 或删除）
- `go test -race ./x/pubsub/...` 通过
- `grep -n "PublishWithContext" x/pubsub/*.go` 结果为空或仅剩 deprecated wrapper

Outcome:
