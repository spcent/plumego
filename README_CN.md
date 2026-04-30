# Plumego — 仅基于golang标准库的 Web 工具包

[![Go 版本](https://img.shields.io/badge/Go-1.24%2B-00ADD8?style=flat&logo=go)](https://go.dev/)
[![状态](https://img.shields.io/badge/status-pre--v1-orange)](https://github.com/spcent/plumego/releases)
[![许可证](https://img.shields.io/badge/license-MIT-green)](LICENSE)

Plumego 是一个小型 Go HTTP 工具包，完全基于标准库实现，同时覆盖路由、中间件、优雅关闭、安全辅助、传输适配器以及可选的 `x/*` 能力包。它设计为嵌入到你自己的 `main` 包中，而不是作为一个独立的框架二进制文件运行。

## 仓库演进方向

目标仓库结构已经收敛为：

- 稳定根级包：`core`、`router`、`contract`、`middleware`、`security`、`store`、`health`、`log`、`metrics`
- 扩展能力包：`x/*`
- 架构权威文档：`docs/architecture/*`
- 机器可读规则：`specs/*`
- 仓库原生执行面：`tasks/*`

仓库控制面拆分如下：

- `docs/`：面向人的说明、架构、模块 primer 与路线图
- `specs/`：机器可读规则、ownership、依赖策略与 change recipe
- `tasks/`：可执行工作卡与 agent 面向的任务队列

不要把 `specs/` 挪进 `docs/`。在 Plumego 中，`specs/` 是一等仓库控制面，而不是附属说明文档。

后续架构规划与重构请优先参考：

- `docs/architecture/AGENT_FIRST_REPO_BLUEPRINT.md`
- `docs/CANONICAL_STYLE_GUIDE.md`
- `specs/repo.yaml`
- `specs/task-routing.yaml`
- `specs/extension-taxonomy.yaml`
- `specs/package-hotspots.yaml`
- `specs/dependency-rules.yaml`
- `specs/checks.yaml`
- `specs/change-recipes/*`
- `<模块>/module.yaml`

当前优先级与剩余扩展工作见 `docs/ROADMAP.md`。

机器强制执行的仓库护栏位于 `internal/checks/*`，并直接在 CI 中执行。

新的应用结构工作应遵循唯一 canonical 路径：

- 先看 `reference/standard-service`，以它作为目录结构和 wiring 的标准
- `reference/standard-service` 有意只依赖稳定根级包；`x/*` 示例都不属于 canonical 路径

## 当前支持矩阵

该矩阵描述的是当前仓库在正式 v1 tag 之前的状态。不同层级的兼容性承诺并不相同。

| 范围 | 状态 | 兼容性承诺 | 模块 |
| --- | --- | --- | --- |
| 稳定根级库 | Stable-root candidate | 这些公开包预期在 v1 hardening 后构成长期稳定 API 面 | `core`、`router`、`contract`、`middleware`、`security`、`store`、`health`、`log`、`metrics` |
| canonical 参考应用 | 支持的参考实现 | 保持与 canonical bootstrap 和稳定根级用法一致，但不作为扩展能力目录 | `reference/standard-service` |
| CLI | v1 hardening scope | 作为命令行工具受支持，而不是 Go import 面；命令行为和生成产物必须与 canonical 文档保持一致 | `cmd/plumego` |
| 面向应用的扩展族 | Experimental | 纳入仓库质量门禁和发布范围，但 API / 配置兼容性尚未冻结 | `x/ai`、`x/data`、`x/fileapi`、`x/frontend`、`x/gateway`、`x/messaging`、`x/observability`、`x/resilience`、`x/rest`、`x/tenant`、`x/websocket` |
| 从属扩展原语 | Experimental | 保持维护和测试，但发现入口应先从所属能力族开始，兼容性尚未冻结 | `x/cache`、`x/devtools`、`x/discovery`、`x/ipc`、`x/mq`、`x/ops`、`x/pubsub`、`x/scheduler`、`x/webhook` |

## 亮点
- **路由器支持分组和参数**：基于 Trie 的匹配器，支持 `/:param` 段、路由冻结，以及每路由/分组的中件栈。
- **中间件链**：日志、恢复、gzip、CORS、超时（默认缓冲上限 10 MiB）、限流、并发限制、请求体大小限制、安全头，以及认证辅助工具，全部包装标准 `http.Handler`。
- **安全辅助**：JWT + 密码工具、安全头策略、输入安全校验与基础防滥用组件，便于进行安全基线加固。
- **集成扩展**：提供 `database/sql`、Redis 缓存，以及扩展层的服务发现与消息能力。优先从 `x/data`、`x/gateway` 与 `x/messaging` 入手；只有在需要直接操作对应原语时才进入 `x/cache`、`x/discovery` 或 `x/mq`。
- **幂等工具**：`store/idempotency` 保留稳定幂等记录与接口；KV/SQL 持久化 provider 位于 `x/data/idempotency`。
- **结构化日志钩子**：接入自定义日志器，并通过中间件钩子收集指标/链路追踪。
- **优雅生命周期**：环境变量加载、连接排水、就绪标志，以及可选的 TLS/HTTP2 配置，带有合理默认值。
- **可选服务**：WebSocket、Webhook、前端托管、网关、消息等能力都位于 `x/*`，并且有意不进入 canonical 应用路径。
- **任务调度**：通过 `x/scheduler` 包提供进程内 cron、延迟任务与可重试任务。

新代码应在应用自己的 wiring 包中显式注册路由、中间件和后台任务。Plumego 已经移除了 `core` 中的兼容组件层。

## 快速开始

建议先读 [`docs/getting-started.md`](./docs/getting-started.md)，再打开 [`reference/standard-service`](./reference/standard-service) 查看 canonical 应用结构。

推荐的 onboarding 顺序：

1. [`docs/getting-started.md`](./docs/getting-started.md) 先跑通最小可运行示例
2. [`reference/standard-service`](./reference/standard-service) 再看 canonical 应用结构与路由 wiring
3. [`docs/README.md`](./docs/README.md) 再进入面向人的文档入口
4. 当 reference 路径不再够用时，再进入 `specs/*` 和 `tasks/*`

最小可运行示例：

```go
package main

import (
    "context"
    "errors"
    "log"
    "net/http"

    "github.com/spcent/plumego/contract"
    "github.com/spcent/plumego/core"
    plog "github.com/spcent/plumego/log"
    "github.com/spcent/plumego/middleware/requestid"
    "github.com/spcent/plumego/middleware/recovery"
)

func main() {
    ctx := context.Background()
    cfg := core.DefaultConfig()
    cfg.Addr = ":8080"
    app := core.New(cfg, core.AppDependencies{Logger: plog.NewLogger()})

    if err := app.Use(
        requestid.Middleware(),
        recovery.Recovery(app.Logger()),
    ); err != nil {
        log.Fatalf("register middleware: %v", err)
    }

    if err := app.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
        if err := contract.WriteResponse(w, r, http.StatusOK, map[string]string{
            "message": "pong",
        }, nil); err != nil {
            http.Error(w, "write response", http.StatusInternalServerError)
        }
    }); err != nil {
        log.Fatalf("register route: %v", err)
    }

    if err := app.Prepare(); err != nil {
        log.Fatalf("prepare server: %v", err)
    }
    srv, err := app.Server()
    if err != nil {
        log.Fatalf("get server: %v", err)
    }
    defer app.Shutdown(ctx)

    log.Println("server started at :8080")
    serveErr := srv.ListenAndServe()
    if srv.TLSConfig != nil {
        serveErr = srv.ListenAndServeTLS("", "")
    }
    if serveErr != nil && !errors.Is(serveErr, http.ErrServerClosed) {
        log.Fatalf("server stopped: %v", serveErr)
    }
}
```

## 配置基础
- 环境变量应在 `main` 包中显式加载。若应用本地工具需要知道当前 `.env` 路径，例如 devtools 热重载，请把它放在应用本地配置里，例如参考实现中的 `cfg.App.EnvFile`。
- `core` 现在走 config-first 构造：先从 `core.DefaultConfig()` 取得基线，再调整 typed `core.AppConfig`，最后传给 `core.New(cfg, ...)`。
- `core.New(cfg, ...)` 默认使用 `NoOpLogger`。如果希望有请求日志或运行期日志，请显式注入 `core.AppDependencies{Logger: ...}`。
- Logger 生命周期归调用方所有。`Prepare()` 和 `Shutdown(ctx)` 不会替你初始化、flush 或关闭注入的 logger 实现。
- 常用变量：`AUTH_TOKEN`（ops 组件默认鉴权配置）、`WS_SECRET`（WebSocket JWT 签名密钥，至少 32 字节）、`WEBHOOK_TRIGGER_TOKEN`、`GITHUB_WEBHOOK_SECRET` 和 `STRIPE_WEBHOOK_SECRET`（详见 `env.example`）。
- `core.AppConfig` 负责服务地址、TLS 以及 HTTP 服务超时/硬化设置。请求体限制与并发限制属于显式中间件 wiring，不属于 `core` 自身配置。
- TLS 仍走同一条显式启动路径：`Prepare()` 会把证书与私钥加载进准备好的 `*http.Server`，随后调用方基于 `Server()` 返回的实例选择 `ListenAndServe()` 或 `ListenAndServeTLS("", "")`。
- 安全基线建议通过 `app.Use(...)` 显式组合，例如 `middleware/security.SecurityHeaders(...)` 与 `middleware/ratelimit.AbuseGuard(...)`。
- 调试模式与 devtools 已拆分：调试开关应放在应用本地配置里，例如参考实现中的 `cfg.App.Debug`；如果需要 devtools，请在应用本地 wiring 中显式注册相关路由，不要把它视为 canonical kernel 的一部分。
- `/_debug` 下的调试端点（路由表、Middleware、配置快照、指标、pprof、手动重载）现在由 `x/devtools` 提供，而不是 `core` 内建。这些端点仅用于本地开发或受保护环境，生产环境应关闭或加访问控制。
- 当接入 `x/devtools` 时，`/_debug/config` 会暴露 first-party tooling 使用的稳定运行时快照：地址、env 文件、服务超时、drain 配置、TLS 配置以及内核的 `preparation_state`。

## Agent 优先工作流
- canonical 应用启动路径从 `reference/standard-service` 开始。
- 机器可读的任务入口规则位于 `specs/task-routing.yaml`。
- 模块 owner、risk 和默认验证入口统一写在各自的 `<模块>/module.yaml` 中。
- 标准变更 recipe 位于 `specs/change-recipes/*`。
- 模块 primer 文档位于 `docs/modules/*`，并应与各模块 manifest 的 `doc_paths` 保持一致。
- 次级任务族入口也已固定：前端静态资源从 `x/frontend` 开始，本地调试能力从 `x/devtools` 开始，服务发现从 `x/discovery` 开始，受保护管理端点从 `x/ops` 开始。
- 这些次级扩展根是能力入口，不是应用 bootstrap surface。

## 能力导览

根 README 作为入口页使用。具体能力说明统一放在 `docs/modules/*`。

稳定根级模块：

- [core](./docs/modules/core/README.md) — 应用内核、生命周期与共享运行时 wiring
- [router](./docs/modules/router/README.md) — 路由匹配、参数、分组与反向路由
- [middleware](./docs/modules/middleware/README.md) — 传输层中间件
- [contract](./docs/modules/contract/README.md) — 响应与错误契约
- [security](./docs/modules/security/README.md) — 认证、安全头、输入安全、防滥用
- [store](./docs/modules/store/README.md) — 持久化原语
- [health](./docs/modules/health/README.md) — 就绪状态与健康模型
- [log](./docs/modules/log/README.md) 与 [metrics](./docs/modules/metrics/README.md) — 基础日志与指标契约

面向应用的扩展能力族：

- [x/tenant](./docs/modules/x-tenant/README.md) — 多租户、配额、策略与租户感知数据路径
- [x/rest](./docs/modules/x-rest/README.md) — 资源 API 与 CRUD 标准化
- [x/websocket](./docs/modules/x-websocket/README.md) — WebSocket 传输
- [x/messaging](./docs/modules/x-messaging/README.md) — 消息能力入口
- [x/fileapi](./docs/modules/x-fileapi/README.md) — 文件上传/下载传输
- [x/gateway](./docs/modules/x-gateway/README.md) 与 [x/discovery](./docs/modules/x-discovery/README.md) — 边缘传输与服务发现
- [x/frontend](./docs/modules/x-frontend/README.md) — 前端静态资源托管
- [x/observability](./docs/modules/x-observability/README.md)、[x/ops](./docs/modules/x-ops/README.md) 与 [x/devtools](./docs/modules/x-devtools/README.md) — 可观测性、受保护运维面与本地调试工具
- [x/data](./docs/modules/x-data/README.md)、[x/cache](./docs/modules/x-cache/README.md) 与 [x/ai](./docs/modules/x-ai/README.md) — 拓扑型数据能力、缓存适配器与 AI 能力

## 参考应用
`reference/standard-service` 是 canonical 参考应用。它只依赖稳定根级包，并演示：

- 默认应用目录结构
- `main.go` 中的显式 bootstrap 流程
- `internal/app/routes.go` 中的显式路由注册
- `internal/config` 下的应用本地配置
- 最小稳定根级 wiring

运行方式：

```bash
go run ./reference/standard-service
```

## 延伸阅读

- [`docs/getting-started.md`](./docs/getting-started.md) — 最小可运行示例
- [`reference/standard-service`](./reference/standard-service) — canonical 参考应用
- [`docs/README.md`](./docs/README.md) — 文档入口
- [`env.example`](./env.example) — 环境变量参考
- [`cmd/plumego/DEV_SERVER.md`](./cmd/plumego/DEV_SERVER.md) — 开发服务器与仪表盘细节

## 开发与测试
- 安装 Go 1.24+（匹配 `go.mod`）。
- 运行与 CI 等价的完整门禁：`make gates`。
- 聚焦修改时，先运行目标包的 `go test -timeout 20s ./<package>` 和
  `go vet ./<package>`。
- 使用 `gofmt -w <paths>` 格式化改过的 Go 文件。

## 开发服务器与仪表盘

`plumego` CLI 包含一个强大的开发服务器，它本身就是使用 plumego 框架构建的。它提供热重载、实时监控和 Web 仪表盘，大大提升开发体验。

仪表盘**默认启用** - 只需运行 `plumego dev` 即可开始使用。

**定位差异与生产建议**
- 参考实现中的 `cfg.App.Debug = true` 会暴露应用级 `/_debug` 端点，属于应用自身调试接口，生产环境应关闭或加访问控制。
- `plumego dev` 仪表盘是本地开发工具，运行独立的仪表盘服务，不建议在生产环境对外暴露。
- 仪表盘可能读取应用的 `/_debug` 端点用于路由/配置/指标/pprof 展示，因此仅建议在本地或受控环境启用。

### 启动开发服务器

```bash
plumego dev
# 仪表盘：http://localhost:9999
# 你的应用：http://localhost:8080
```

### 仪表盘功能

每个 `plumego dev` 会话都包含：

- **实时日志**：流式传输应用程序的 stdout/stderr，支持过滤
- **路由浏览器**：自动发现并展示应用程序的所有 HTTP 路由
- **指标仪表盘**：监控运行时间、PID、健康状态和性能
- **构建管理**：查看构建输出并手动触发重新构建
- **应用控制**：从 UI 中启动、停止和重启应用程序
- **热重载**：文件更改时自动重新构建和重启（< 5 秒）

### 自定义配置

```bash
# 自定义应用端口
plumego dev --addr :3000

# 自定义仪表盘端口
plumego dev --dashboard-addr :8888

# 自定义监听模式
plumego dev --watch "**/*.go,**/*.yaml"

# 调整热重载灵敏度
plumego dev --debounce 1s
```

完整文档请参见 `cmd/plumego/DEV_SERVER.md`。

## 文档
规范文档入口与优先级顺序：`docs/README.md`。
