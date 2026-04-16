# Card 0011

Priority: P2
State: active
Primary Module: x/resilience
Owned Files:
  - x/resilience/module.yaml
  - security/abuse/limiter.go

Depends On: —

Goal:
代码库中存在 3 个相互独立的 rate limiter 实现，没有任何代码复用：

- `security/abuse/limiter.go`：完整 token bucket 实现，支持 per-IP、per-key、全局限流
- `x/pubsub/ratelimit.go`：为 pubsub 专写的 `RateLimitedPubSub` 包装器，内含独立 token bucket
- `x/tenant/core/ratelimit.go`：`InMemoryRateLimitManager` 含第三套 token bucket 逻辑

与此同时，`x/resilience` 的 `module.yaml` 明确写道："Put reusable resilience primitives here"，
但 x/resilience 目前只有 circuitbreaker，无任何包导入它（零使用）。

本卡目标：将可复用的 rate limiter 原语迁入 `x/resilience/ratelimit`，让上述三个场景
通过共享原语实现，消除重复逻辑。

Scope:
- 在 `x/resilience/ratelimit/` 新建通用 token bucket rate limiter：
  - 支持按 key 分桶（per-IP、per-tenant、per-topic 通用）
  - `Allow(key string) bool` + `Wait(ctx, key) error` 两种使用模式
  - 可配置 QPS、Burst、TTL（idle key 清理）
- 将 `x/pubsub/ratelimit.go` 的 token bucket 逻辑替换为调用 x/resilience/ratelimit
  （保持 RateLimitedPubSub 的公共 API 不变）
- 将 `x/tenant/core/ratelimit.go` 的 InMemoryRateLimitManager 内部实现
  替换为调用 x/resilience/ratelimit（保持接口不变）
- `security/abuse/limiter.go` 视复杂度决定：若其 per-key 窗口逻辑可直接用通用原语
  替代则迁移；若业务逻辑差异较大则暂保留，仅做内部复用注释
- 更新 x/resilience 的 module.yaml，将 `ratelimit` 加入 `public_entrypoints`

Non-goals:
- 不改变 security/abuse、x/pubsub、x/tenant 的任何对外 API
- 不将 security/abuse 的策略逻辑（ban list、suspicious pattern）移出
- 不引入任何外部依赖（仍使用 sync/atomic 等标准库）
- 不修改 middleware/ratelimit（该包使用 security/abuse，不直接实现 bucket）

Files:
  - x/resilience/ratelimit/ratelimit.go（新建）
  - x/resilience/ratelimit/ratelimit_test.go（新建）
  - x/resilience/module.yaml（更新 public_entrypoints）
  - x/pubsub/ratelimit.go（内部替换）
  - x/tenant/core/ratelimit.go（内部替换）
  - security/abuse/limiter.go（可选部分复用）

Tests:
  - go test -race ./x/resilience/...
  - go test ./x/pubsub/...
  - go test ./x/tenant/...

Docs Sync:
  - docs/modules/x-resilience/README.md（补充 ratelimit 说明）

Done Definition:
- `x/resilience/ratelimit` 包存在，有测试，`go test -race` 通过
- x/pubsub 和 x/tenant 的 token bucket 逻辑不再各自实现，改用 x/resilience/ratelimit
- `grep -rn "golang.org/x/time/rate" .` 结果统一（只在 x/resilience 内部或完全消失）
- 所有被修改包的测试通过

Outcome:
