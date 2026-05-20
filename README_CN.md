# Plumego — 标准库优先的 Go Web 工具包

[![Go 版本](https://img.shields.io/badge/Go-1.26%2B-00ADD8?style=flat&logo=go)](https://go.dev/)
[![状态](https://img.shields.io/badge/status-v1.0.0-blue)](https://github.com/spcent/plumego/releases/tag/v1.0.0)
[![许可证](https://img.shields.io/badge/license-MIT-green)](LICENSE)

Plumego 是一个基于 Go 标准库构建的小型 HTTP 工具包。它把 `net/http`
兼容性放在核心位置：处理器仍是普通的
`func(http.ResponseWriter, *http.Request)`，中间件仍包装 `http.Handler`，
应用装配仍显式写在你自己的 `main` 包中。

稳定表面刻意保持收敛。先从 `core`、`router`、`contract` 和
`middleware` 开始；只有在职责需要时，再加入 `security`、`store`、
`health`、`log` 和 `metrics`。

## 快速开始

创建 `main.go`：

```go
package main

import (
	"log"
	"net/http"

	"github.com/spcent/plumego/contract"
	"github.com/spcent/plumego/core"
)

func main() {
	cfg := core.DefaultConfig()
	cfg.Addr = ":8080"
	app := core.New(cfg, core.AppDependencies{})
	if err := app.Get("/ping", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = contract.WriteResponse(w, r, http.StatusOK, map[string]string{"message": "pong"}, nil)
	})); err != nil {
		log.Fatal(err)
	}
	if err := app.Prepare(); err != nil {
		log.Fatal(err)
	}
	srv, err := app.Server()
	if err != nil {
		log.Fatal(err)
	}
	log.Fatal(srv.ListenAndServe())
}
```

运行：

```bash
go mod init example.com/hello
go get github.com/spcent/plumego@latest
go run main.go
```

打开 `http://localhost:8080/ping`。

完成这个示例后，再阅读
[`reference/standard-service`](./reference/standard-service)，查看生产风格的规范应用结构。

## 选择你的起点

根据你的项目场景选择对应入口：

| 我想构建… | 从这里开始 |
| --- | --- |
| 普通 JSON API | `reference/standard-service` → 仅使用稳定根包 |
| 带 CRUD 约定的 REST 资源 | `reference/with-rest` → `x/rest` |
| 多租户 SaaS API | `reference/with-tenant` → `x/tenant` |
| API 网关或反向代理 | `reference/with-gateway` → `x/gateway` |
| 实时 WebSocket 功能 | `reference/with-websocket` → `x/websocket` |
| AI 驱动的服务 | `reference/with-ai` → `x/ai/provider` |
| 带消息队列/Webhook 的服务 | `reference/with-messaging` → `x/messaging` |
| gRPC + HTTP 混合服务 | `reference/with-rpc` → `x/rpc` |
| 可观测性（Prometheus / OpenTelemetry） | `x/observability` |

所有路径都以 `reference/standard-service` 为基础结构。扩展包是显式增量，而不是另一套应用启动框架。

## 为什么选择 plumego

Plumego 适合那些需要比原生 `http.ServeMux` 更多结构、但又不想引入大型框架模型的 Go 服务。

| 原则 | plumego 的做法 |
| --- | --- |
| 标准库优先 | 使用普通 handler、中间件、request、response writer 和 `*http.Server`。 |
| 显式装配 | 路由、中间件、依赖和生命周期都能在构造位置直接看到。 |
| 小型稳定表面 | 稳定根包职责清晰，避免变成功能目录。 |
| 便于代理维护 | `specs/`、`tasks/` 和每模块 `module.yaml` 让范围与验证方式可发现。 |
| 可选能力 | 产品功能和协议适配放在稳定学习路径之外。 |

## 标准库对比

| 能力 | `http.ServeMux` | plumego |
| --- | --- | --- |
| 基础路由 | 方法处理由调用方自行判断。 | 通过 `Get`、`Post` 和 `AddRoute` 等方法注册一个方法、一个路径、一个处理器。 |
| `{param}` 路径提取 | 调用方手动解析路径段。 | 路由器匹配路径参数，并通过请求上下文读取。 |
| 路由分组 | 调用方手动重复路径前缀。 | 分组为相关路由注册统一前缀。 |
| 分组中间件 | 调用方手动为每棵子树组合处理器。 | 分组可携带共享中间件，同时保持 `http.Handler` 形状。 |
| 命名路由与反向 URL | 调用方手动拼接 URL。 | 命名路由通过 app/router API 支持反向 URL 生成。 |
| 准备后冻结路由 | 调用方只要继续改装配代码，路由就可能变化。 | `Prepare` 在服务启动前冻结路由变更。 |
| 结构化错误响应 | 调用方定义每一种响应形状。 | `contract.WriteError` 提供规范结构化错误路径。 |
| Request ID 上下文传递 | 调用方自行选择并传播约定。 | Request ID 辅助函数使用显式上下文访问器，并提供中间件支持。 |
| 优雅生命周期 | 调用方自行组织 server 设置与关闭策略。 | `Prepare`、`Server` 和 `Shutdown` 让生命周期显式且可复用。 |

## 包导览

| 包 | 职责 |
| --- | --- |
| `core` | 应用构造、路由注册入口、中间件挂载和服务生命周期。 |
| `router` | 路由匹配、路径参数、分组、路由元数据和反向 URL 生成。 |
| `contract` | 规范响应写入、结构化错误构造、请求元数据和传输绑定辅助。 |
| `middleware` | 传输层中间件组合和一方中间件包。 |
| `security` | 认证、JWT、密码、安全头、输入安全和防滥用原语。 |
| `store` | 面向缓存、KV、文件、数据库和幂等场景的稳定存储契约与内存原语。 |
| `health` | 面向应用与依赖状态的健康和就绪模型。 |
| `log` | 最小日志接口和默认日志器实现。 |
| `metrics` | 面向计数器、仪表、耗时和收集器的最小指标契约。 |

可选能力族位于 `x/*`。应把它们视为稳定根路径上的显式增量，而不是另一套应用结构。

## Agent-First Design

Plumego 使用 agent-first 控制面维护仓库：`docs/` 解释架构，`specs/`
记录可由机器检查的边界，`tasks/` 定义可审查的执行单元，`reference/`
展示规范装配方式。`internal/checks/` 下的检查会在本地和 CI 中执行关键边界验证，
让自动化修改带着可审查证据进入 review，而不是依赖隐含约定。完整模型和采用路径见
[`docs/AGENT_FIRST.md`](./docs/AGENT_FIRST.md)。

## 获取帮助

- 从 [`docs/getting-started.md`](./docs/getting-started.md) 开始，查看最小可运行教程。
- 使用 [`reference/standard-service`](./reference/standard-service) 作为规范应用结构。
- 阅读 [`docs/CANONICAL_STYLE_GUIDE.md`](./docs/CANONICAL_STYLE_GUIDE.md)，了解处理器、中间件、路由和依赖注入约定。
- 浏览 [`docs/modules`](./docs/modules)，查看各包的模块导读。
- 查看 [`docs/ADOPTION_PATH.md`](./docs/ADOPTION_PATH.md)，了解 5 分钟、30 分钟和 1 天采用路径。
- 查看 [`docs/troubleshooting.md`](./docs/troubleshooting.md)，解决常见问题：路由冻结、中间件顺序、JWT 验证错误和生命周期问题。
- 查看 [`docs/benchmarks/README.md`](./docs/benchmarks/README.md)，了解与 Chi、Gin、Echo 的性能对比。
