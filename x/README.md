# Plumego Extensions (`x/*`)

`x/*` 是 Plumego 的能力扩展层，承载 beta 和 experimental 功能。稳定根（`core`, `router`, `contract`, ...）保持零运行时依赖，扩展层按需引入外部依赖。

> 扩展生命周期遵循 [Extension Maturity Model](../docs/concepts/extension-maturity.md)。

## Maturity Matrix

| 扩展族 | 子包 | 成熟度 | 推荐入口 | 可生产使用 |
|---|---|---|---|---|
| **ai** | provider, session, streaming, tool | ⭐ Experimental | `x/ai/provider` | ⚠️ 评估中 |
| **data** | cache, kvengine, pgx, sharding | ⭐ Experimental | `docs/modules/x/data/README.md` | ⚠️ 评估中 |
| **fileapi** | - | ⭐ Experimental | `x/fileapi` | ⚠️ 评估中 |
| **frontend** | - | 🔶 Beta | `x/frontend` | ✅ 可用 |
| **gateway** | discovery, ipc, transform | 🔶 Beta | `x/gateway` | ✅ 可用 |
| **messaging** | pubsub, scheduler, webhook | 🔶 Beta | `x/messaging/pubsub` | ✅ 可用 |
| **observability** | ops, exporter | 🔶 Beta | `x/observability/ops` | ✅ 可用 |
| **openapi** | - | ⭐ Experimental | `x/openapi` | ⚠️ 评估中 |
| **resilience** | circuitbreaker, retry | 🔶 Beta | `x/resilience/circuitbreaker` | ✅ 可用 |
| **rest** | - | 🔶 Beta | `x/rest` | ✅ 可用 |
| **rpc** | server | ⭐ Experimental | `x/rpc/server` | ⚠️ 评估中 |
| **tenant** | core, policy, quota, resolve, transport | 🔶 Beta | `x/tenant/core` | ✅ 可用 |
| **validate** | - | 🔶 Beta | `x/validate` | ✅ 可用 |
| **websocket** | - | 🔶 Beta | `x/websocket` | ✅ 可用 |

> ⭐ = Experimental · 🔶 = Beta · ✅ = GA-track
>
> 权威来源: `specs/extension-maturity.yaml` 和 `specs/extension-beta-evidence.yaml`

## 导入约束

- 稳定根（`core`, `router`, `contract`, ...）**禁止**导入 `x/*`
- `x/*` 之间导入需遵循 `specs/dependency-rules.yaml` 的 `cross-extension` 约束
- `reference/*` 可自由选择所需扩展

## 快速开始

参考应用展示每个扩展的推荐集成方式：

| 场景 | 参考应用 |
|---|---|
| LLM/AI 集成 | `reference/with-ai` |
| 多租户 API | `reference/with-tenant` |
| WebSocket 实时通信 | `reference/with-websocket` |
| gRPC / RPC | `reference/with-rpc` |
| 网关/代理 | `reference/with-gateway` |
| 消息队列 | `reference/with-messaging` |
| 可观测性 | `reference/with-observability` |
| REST CRUD | `reference/with-rest` |
