# Plumego — 仅基于golang标准库的 Web 工具包

[![Go 版本](https://img.shields.io/badge/Go-1.24%2B-00ADD8?style=flat&logo=go)](https://go.dev/)
[![版本](https://img.shields.io/badge/version-v1.0.0--rc.1-blue)](https://github.com/spcent/plumego/releases)
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

## v1 支持矩阵

Plumego 的 v1 发布范围覆盖当前仓库中所有已存在模块，但不同层级的兼容性承诺并不相同。

| 范围 | v1 状态 | 兼容性承诺 | 模块 |
| --- | --- | --- | --- |
| 稳定根级库 | GA | 这些公开包构成 v1 的长期稳定 API 面 | `core`、`router`、`contract`、`middleware`、`security`、`store`、`health`、`log`、`metrics` |
| canonical 参考应用 | 支持的参考实现 | 保持与 canonical bootstrap 和稳定根级用法一致，但不作为扩展能力目录 | `reference/standard-service` |
| CLI | 纳入 v1 发布范围 | 作为命令行工具受支持，而不是 Go import 面；命令行为和生成产物必须与 canonical 文档保持一致 | `cmd/plumego` |
| 面向应用的扩展族 | Experimental | 纳入仓库质量门禁和发布范围，但 API / 配置兼容性尚未冻结 | `x/ai`、`x/data`、`x/devtools`、`x/discovery`、`x/frontend`、`x/gateway`、`x/messaging`、`x/observability`、`x/ops`、`x/rest`、`x/tenant`、`x/websocket` |
| 从属扩展原语 | Experimental | 保持维护和测试，但发现入口应先从所属能力族开始，兼容性尚未冻结 | `x/ipc`、`x/mq`、`x/pubsub`、`x/scheduler`、`x/webhook` |

## 亮点
- **路由器支持分组和参数**：基于 Trie 的匹配器，支持 `/:param` 段、路由冻结，以及每路由/分组的中件栈。
- **中间件链**：日志、恢复、gzip、CORS、超时（默认缓冲上限 10 MiB）、限流、并发限制、请求体大小限制、安全头，以及认证辅助工具，全部包装标准 `http.Handler`。
- **安全辅助**：JWT + 密码工具、安全头策略、输入安全校验与基础防滥用组件，便于进行安全基线加固。
- **集成扩展**：提供 `database/sql`、Redis 缓存，以及扩展层的服务发现与消息能力。优先从 `x/discovery` 与 `x/messaging` 入手；只有在需要直接操作队列原语时才进入 `x/mq`。
- **幂等工具**：提供 `store/idempotency` 的 KV/SQL 幂等存储接口。
- **结构化日志钩子**：接入自定义日志器，并通过中间件钩子收集指标/链路追踪。
- **优雅生命周期**：环境变量加载、连接排水、就绪标志，以及可选的 TLS/HTTP2 配置，带有合理默认值。
- **可选服务**：WebSocket、Webhook、前端托管、网关、消息等能力都位于 `x/*`，并且有意不进入 canonical 应用路径。
- **任务调度**：通过 `scheduler` 包提供进程内 cron、延迟任务与可重试任务。

新代码应在应用自己的 wiring 包中显式注册路由、中间件和后台任务。Plumego 已经移除了 `core` 中的兼容组件层。

## 快速开始

建议先读 [`docs/getting-started.md`](./docs/getting-started.md)，再打开 [`reference/standard-service`](./reference/standard-service) 查看 canonical 应用结构。

最小可运行示例：

```go
package main

import (
    "log"
    "net/http"

    "github.com/spcent/plumego/contract"
    "github.com/spcent/plumego/core"
    plog "github.com/spcent/plumego/log"
    "github.com/spcent/plumego/middleware/requestid"
    "github.com/spcent/plumego/middleware/recovery"
)

func main() {
    app := core.New(
        core.WithAddr(":8080"),
        core.WithLogger(plog.NewGLogger()),
    )

    if err := app.Use(
        requestid.Middleware(),
        recovery.Recovery(app.Logger()),
    ); err != nil {
        log.Fatalf("register middleware: %v", err)
    }

    app.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
        if err := contract.WriteResponse(w, r, http.StatusOK, map[string]string{
            "message": "pong",
        }, nil); err != nil {
            http.Error(w, "write response", http.StatusInternalServerError)
        }
    })

    log.Println("server started at :8080")
    log.Fatal(http.ListenAndServe(":8080", app))
}
```

## 配置基础
- 环境变量应在 `main` 包中显式加载。`core.WithEnvPath` 仅记录路径，供需要该信息的组件使用，例如 devtools 热重载。
- `core.New(...)` 默认使用 `NoOpLogger`。如果希望有请求日志或运行期日志，请显式注入 `core.WithLogger(...)`。
- 常用变量：`AUTH_TOKEN`（ops 组件默认鉴权配置）、`WS_SECRET`（WebSocket JWT 签名密钥，至少 32 字节）、`WEBHOOK_TRIGGER_TOKEN`、`GITHUB_WEBHOOK_SECRET` 和 `STRIPE_WEBHOOK_SECRET`（详见 `env.example`）。
- 应用默认包括 10485760 字节（10 MiB）请求体限制、256 并发请求限制（带队列）、HTTP 读/写超时，以及 5000ms（5 秒）优雅关闭窗口。可通过 `core.With...` 选项覆盖。
- 安全基线建议通过 `app.Use(...)` 显式组合，例如 `middleware/security.SecurityHeaders(...)` 与 `middleware/ratelimit.AbuseGuard(...)`。
- 调试模式与 devtools 已拆分：`core.WithDebug()` 只开启调试行为；如果需要 devtools，请在应用本地 wiring 中显式注册相关路由，不要把它视为 canonical kernel 的一部分。
- `/_debug` 下的调试端点（路由表、Middleware、配置快照、指标、pprof、手动重载）现在由 `x/devtools` 提供，而不是 `core` 内建。这些端点仅用于本地开发或受保护环境，生产环境应关闭或加访问控制。

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
- [`docs/README.md`](./docs/README.md) — 文档入口
- [`env.example`](./env.example) — 环境变量参考
- [`cmd/plumego/DEV_SERVER.md`](./cmd/plumego/DEV_SERVER.md) — 开发服务器与仪表盘细节

## 开发与测试
- 安装 Go 1.24+（匹配 `go.mod`）。
- 运行测试：`go test ./...`
- 使用 Go 工具链进行格式化和静态检查（`go fmt`、`go vet`）。

## 开发服务器与仪表盘

`plumego` CLI 包含一个强大的开发服务器，它本身就是使用 plumego 框架构建的。它提供热重载、实时监控和 Web 仪表盘，大大提升开发体验。

仪表盘**默认启用** - 只需运行 `plumego dev` 即可开始使用。

**定位差异与生产建议**
- `core.WithDebug` 会暴露应用级 `/_debug` 端点，属于应用自身调试接口，生产环境应关闭或加访问控制。
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
