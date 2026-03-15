# Plumego Roadmap

This roadmap reflects the current repository direction after the agent-first restructuring work. It is architecture-first and optimized for low-regret iteration.

## Current Position

Plumego now has:

- stable roots with narrow responsibilities
- a canonical application path in `reference/standard-service`
- explicit `x/*` extension discovery rules
- machine-readable workflow metadata under `specs/*`
- removal of compatibility-style component APIs from `core` and extension entrypoints
- a formal deprecation policy for stable roots
- a full scaffold and code generation CLI in `cmd/plumego`
- deep AI gateway capabilities in `x/ai` (experimental)
- multi-tenancy middleware in `x/tenant` (experimental)

The next stages harden and grow the ecosystem rather than restructure it.

## Roadmap Principles

- Keep `core` as a kernel, not a feature catalog.
- Preserve `net/http` compatibility.
- Prefer explicit app-local wiring over hidden registration.
- Maintain one canonical entrypoint per capability family.
- Bias toward machine-readable repo rules so agents can work with low ambiguity.
- Reduce stable-root migration debt before adding new broad capabilities.

---

## Phase 1: Finish the Agent Workflow Control Plane

Status: complete

Goals:

- make task discovery deterministic
- make module ownership and validation mechanically discoverable
- make repo-native task slicing part of normal development

Planned work:

- add `tasks/cards/*.md` for reversible work items
- add a checker that verifies every manifest `doc_paths` target exists
- add a checker that verifies `specs/change-recipes/*` remain discoverable from canonical repo metadata
- expand `specs/agent-entrypoints.yaml` to cover more task families at finer granularity
- keep `docs/modules/*` synchronized with module manifests

Exit criteria:

- agents can choose a start path without reading broad prose first
- manifests, primers, ownership data, and recipes stay in sync under checks
- new contributors can follow the same workflow without tribal knowledge

Execution surface:

- `tasks/cards/` is the repo-native queue for reversible work items
- task cards should stay small enough for one focused implementation pass
- task cards are workflow assets, not archival prose

## Phase 2: Complete Extension Taxonomy Convergence

Status: substantially complete

Goals:

- remove ambiguity between sibling extension packages
- keep one obvious app-facing surface per capability family
- keep primitive packages available without turning them into discovery traps

Planned work:

- keep `x/messaging` as the only app-facing messaging family entrypoint
- keep `x/mq`, `x/pubsub`, `x/webhook`, and `x/scheduler` explicitly subordinate
- keep `x/gateway` and `x/rest` as non-overlapping entrypoints
- keep `x/ops` and `x/observability` as non-overlapping discovery roots
- keep `x/frontend`, `x/devtools`, and `x/discovery` clearly documented as secondary capability roots, not bootstrap surfaces

Completed in this phase:

- converged messaging-family discovery onto `x/messaging`
- clarified subordinate status for `x/mq`, `x/pubsub`, and `x/webhook`
- separated `x/gateway` from `x/rest` at the metadata and primer level
- separated `x/ops` from `x/observability` at the metadata and primer level
- completed secondary discovery guidance for `x/frontend`, `x/devtools`, and `x/discovery`

Remaining work:

- keep `x/scheduler` explicitly subordinate in family-level documentation
- watch for taxonomy drift as new extensions or adapters are added
- add stronger checks if future ambiguity reappears in manifests or top-level docs

Exit criteria:

- app-facing docs never send users to primitive sibling packages first
- capability family entrypoints are consistent across docs, manifests, and examples
- extension naming no longer forces agents to guess where to start

## Phase 3: Harden `x/rest` as the Reusable Resource Interface Layer

Status: substantially complete

Goals:

- make `x/rest` the reusable home for resource-interface conventions
- preserve explicit route binding while maximizing CRUD reuse
- keep `x/rest` distinct from bootstrap and proxy topology concerns

Planned work:

- keep route-level tests around `RegisterContextResourceRoutes(...)` current as the public registration surface evolves
- keep examples showing `ResourceSpec -> repository -> routes` current as the canonical reuse path
- continue moving query, pagination, hooks, and transformer behavior toward spec-driven configuration
- keep reusable error and response conventions aligned with `contract`
- keep layering guidance between handlers, repositories, and `x/rest` controllers explicit

Completed in this phase:

- locked down the public route registration surface with focused `x/rest` route tests
- added a canonical `x/rest` example for `ResourceSpec -> NewDBResource -> RegisterContextResourceRoutes(...)`
- tightened `ApplyResourceSpec(...)` so controller defaults flow through one orchestration path
- clarified `x/rest` response guidance and the layering boundary between app wiring, controllers, repositories, and domain logic

Remaining work:

- deepen spec-driven orchestration only when new duplication or ambiguity appears
- add further examples if new reusable controller patterns emerge
- keep `x/rest` guidance synchronized as the public API evolves

Guidance constraints:

- keep response and error conventions aligned with `contract`
- do not turn `x/rest` into a bootstrap or transport-contract replacement layer

Execution approach:

- land route-level tests first to lock the public registration surface
- add examples before making deeper orchestration changes
- keep each `x/rest` hardening step reversible and independently testable

Exit criteria:

- resource APIs can be standardized without inventing per-service scaffolding
- `x/rest` stays clearly distinct from `reference/standard-service` and `x/gateway`
- spec-driven behavior is covered by focused tests and examples

## Phase 4: Shrink Stable-Root Migration Debt

Status: complete

Goals:

- keep stable roots narrow and durable
- move topology-heavy or feature-heavy logic out of stable packages
- align package placement with the architecture blueprint

Completed work:

- migrated `store/cache/distributed` → `x/cache/distributed` (consistent-hashing distributed cache)
- migrated `store/cache/redis` → `x/cache/redis` (Redis adapter)
- `store/cache` now contains only abstract types, interfaces, and in-memory implementations
- audited stable `store` topology debt; established rule that no new topology-heavy packages may be added under stable `store`
- confirmed transport health endpoints remain outside `health`; `health` owns models and readiness state only
- tightened boundary documentation blocking tenant leakage into stable `middleware` and `store`
- reduced observability catch-all drift in stable `middleware`; stable middleware owns transport primitives, `x/observability` owns adapters and export wiring
- all `specs/check-baseline` migration debt files removed; checks now enforce boundaries without open exceptions

Exit criteria met:

- stable roots read as long-lived primitives rather than convenience catalogs
- migration debt files have been removed; checks block new drift instead of documenting it
- architecture checks pass cleanly with no baseline suppression

## Phase 5: Reference and Scaffold System

Status: complete

Goals:

- make the canonical app path easy to copy without reintroducing hidden patterns
- ensure scaffolds follow the same rules as the reference app

Completed work:

- `reference/standard-service` is the single canonical application layout demonstrating
  explicit route registration, constructor-based wiring, and stdlib-only dependencies
- `cmd/plumego new` scaffold command generates projects from templates (`minimal`, `api`,
  `fullstack`, `microservice`) each based on the canonical `reference/standard-service` structure
- `cmd/plumego generate` code generation command produces handlers, middleware, and related
  boilerplate aligned with the canonical style
- scaffold templates follow the same explicit wiring rules as the canonical reference; no hidden
  registration or global init patterns are introduced

Completed in this phase:

- added `canonical` template to `plumego new` that mirrors `reference/standard-service` exactly
- added `reference/with-messaging` demo: in-process broker wired into the standard-service shape
- added `reference/with-gateway` demo: reverse proxy wired into the standard-service shape
- added `reference/with-websocket` demo: WebSocket server wired into the standard-service shape
- added `reference/with-webhook` demo: inbound webhook receiver wired into the standard-service shape
- each feature reference is clearly marked non-canonical in its README and doc comment
- added `FindReferenceXImports` check in `internal/checks/checkutil` to detect x/* drift in `reference/standard-service`
- enhanced `internal/checks/reference-layout` to enforce the drift check and require feature reference paths
- updated `specs/agent-entrypoints.yaml` with a `scaffold` task entrypoint

Exit criteria:

- feature demos do not pollute the canonical learning path (reference variants still needed)

## Phase 6: Release Readiness Toward v1

Status: substantially complete

Goals:

- turn the architecture cleanup into a durable release baseline
- make quality gates strict enough for stable public adoption

Completed work:

- audited all stable-root public API surfaces; removed implementation details that leaked into
  the exported API before v1 freeze:
  - `router`: unexported `CacheEntry`, `PatternCacheEntry`, `RouteMatcher`, `NewRouteMatcher`,
    `IsParameterized` — internal implementation details with no external callers
  - `metrics`: removed `MetricsMiddleware` and `MetricsHandler` — middleware helpers that
    violated the module boundary (use `middleware/httpmetrics.Middleware` instead)
- expanded negative-path coverage for critical roots:
  - `contract`: `WriteBindError` is now tested against all sentinel errors with HTTP status and
    error code assertions; field-level validation errors are also covered
  - `router`: frozen-router registration, duplicate route, param validation failure, unknown path,
    and double-slash path are all covered with negative assertions
- defined a formal deprecation policy in `docs/DEPRECATION.md` covering the compatibility
  promise, four-step deprecation process, extension package exemption, and governance rules
- all quality gates pass (`go test -race ./...`, `go vet ./...`, all `internal/checks/*`)

Remaining work:

- Phase 5 scaffold work is substantially complete; the remaining Phase 5 item (non-canonical
  extension reference variants) does not block the API freeze
- keep `x/*` extension packages aligned with stable-root changes as the canonical reference
  evolves

Exit criteria met:

- public docs describe only the supported explicit APIs
- quality gates are green without new temporary exceptions
- maintainers can describe the supported architecture without caveats

See `docs/DEPRECATION.md` for the formal extension evolution policy.

---

## Phase 7: CLI 完善与代码生成质量提升

Status: planned

Goals:

- 消除 `cmd/plumego generate` 生成代码中残留的 `// TODO` 占位符
- 让脚手架生成的代码可以直接编译运行，而不只是提示用户还需要填写
- 对齐 `plumego new` 各模板与 `reference/standard-service` 最新结构

背景：

当前 `cmd/plumego/internal/codegen` 和 `cmd/plumego/internal/scaffold` 在生成
handler、repository、service 骨架时会输出 `// TODO: define service methods` 以及
`json.Encode(CreateXResponse{ID: "TODO"})` 等占位内容。这会让初次使用者困惑并且
无法直接运行。

Planned work:

- 将 handler 生成模板替换为可直接编译的最小实现（使用 `contract.WriteResponse` 返回 stub 响应）
- 将 repository 生成模板替换为接口+内存实现骨架，而非空注释
- 将 service 生成模板替换为带方法签名的可编译 stub
- 为 `plumego new` 的 `fullstack` 和 `microservice` 模板补全缺失文件
- 补充 `plumego generate` 的集成测试，确保每条代码生成路径产出可编译的 Go 文件
- 对 `cmd/plumego` 命令行参数做更友好的错误提示

Non-goals:

- 不改变生成代码的结构约定（仍遵循 canonical style）
- 不引入模板引擎外部依赖

Exit criteria:

- `plumego new <name>` 生成的任意模板都可直接 `go build ./...` 通过
- `plumego generate handler <Name>` 生成的 handler 文件零编译错误
- `plumego generate` 生成的文件不含裸 `// TODO: ...` 占位行

---

## Phase 8: x/ai 扩展层稳定路径

Status: planned

Goals:

- 明确 `x/ai` 各子包从实验性到稳定的升级路径
- 收敛 AI provider 接口，覆盖当前已实现的 Claude 和 OpenAI 适配器
- 补齐 orchestration、semanticcache、multimodal 的集成测试

背景：

`x/ai` 已经有相当深度的实现，包括：provider 抽象、Claude/OpenAI 适配器、
orchestration 工作流、语义缓存（含向量存储）、多模态输入、SSE 流式输出、
AI 专用限流与熔断、工具调用（tool use）框架。
但整体仍标记为 experimental，且内部存在 `SPRINT2_ENHANCEMENTS.md` 等
开发草稿文档，说明部分功能还在迭代中。

Planned work:

- 确定 `x/ai` 内哪些子包可以升级为稳定 API（候选：`provider`、`session`、`streaming`、`tool`）
- 确定哪些子包保持实验性（候选：`orchestration`、`semanticcache`、`marketplace`、`distributed`）
- 清理内部开发草稿（如 `SPRINT2_ENHANCEMENTS.md`），将相关内容迁移到正式 doc 或 CHANGELOG
- 为 `provider.Provider` 接口补充 mock 实现，供下游测试用例使用
- 为 `orchestration.Workflow` 的多步 agent 场景补充集成测试
- 为 `semanticcache` 补充离线向量检索和 cache hit/miss 的端到端测试
- 在 `docs/modules/x-ai` 中明确各子包的稳定等级和使用建议
- 将 `x/ai/marketplace` 的 provider 注册模式与 `x/ai/provider` 接口对齐

Non-goals:

- 不将 `x/ai` 提升为稳定根包
- 不在 `x/ai` 内部实现业务提示词（prompt flow）逻辑
- 不依赖 `x/tenant` 或核心 bootstrap 层

Exit criteria:

- `x/ai/provider`、`x/ai/session`、`x/ai/streaming`、`x/ai/tool` 的公开接口在文档中
  明确标注稳定等级
- 无内部开发草稿文档遗留在包目录下
- AI provider mock 可被下游测试直接引用
- 所有测试在 `-race` 下通过

---

## Phase 9: x/tenant 多租户稳定路径

Status: planned

Goals:

- 让 `x/tenant` 从实验性推进到可用于生产的参考实现
- 提供生产级 JWT 租户 ID 校验（替代简单的 header 提取）
- 为分布式场景提供租户配额的持久化存储适配器

背景：

`x/tenant` 目前拥有完整的多租户中间件栈（resolve → ratelimit → quota → policy），
但存在以下生产短板：
1. 默认租户解析使用 header（`X-Tenant-ID`），未提供 JWT claim 提取路径
2. 配额存储默认为内存实现（`InMemoryQuotaStore`），多实例部署下无法共享状态
3. 租户配置数据库后端（`DBTenantConfigManager`）的 SQL schema 和迁移脚本
   需要补全文档
4. 租户隔离的数据库查询过滤（`TenantDB`）目前仅支持单表 `WHERE tenant_id = ?`，
   尚不支持 JOIN 查询的透明注入

Planned work:

- 在 `x/tenant/resolve` 中增加 JWT claim 提取方式的租户解析适配器
- 在 `x/tenant/store` 中增加基于 Redis 的分布式配额存储适配器
- 补全 `x/tenant/config` 的数据库 schema（SQL 迁移文件）到 `docs/migrations/`
- 明确 `TenantDB` 的查询拦截范围限制，在文档中标注不支持的 SQL 模式
- 为 `x/tenant` 补充端到端集成测试：配额耗尽 + retry-after header 验证
- 在 `docs/architecture/X_TENANT_BLUEPRINT.md` 中增加生产部署建议章节

Non-goals:

- 不在 `x/tenant` 内实现租户 CRUD 业务接口（属于各应用的 domain 层）
- 不在稳定根包中引入任何租户感知代码

Exit criteria:

- JWT 租户解析路径有完整示例和测试
- Redis 配额存储适配器通过并发测试
- `docs/migrations/` 包含租户相关 schema
- `TenantDB` 的 SQL 支持范围在 README 中明确标注

---

## Phase 10: x/discovery 服务发现后端扩展

Status: planned

Goals:

- 实现 `x/discovery` 包注释中已标注为 "future" 的服务发现后端
- 保持与 `x/gateway` 反向代理的显式集成路径

背景：

`x/discovery` 目前只有两个后端实现：
- `discovery.NewStatic(...)` — 静态配置
- `discovery.NewConsul(...)` — Consul

包注释中已明确提到 Kubernetes 和 etcd 作为未来后端，但尚未实现。
`x/gateway` 的反向代理依赖 `discovery.Provider` 接口，这是两者之间的唯一耦合点。

Planned work:

- 实现 `x/discovery` Kubernetes 后端（基于 Endpoints API 或 EndpointSlices）
- 实现 `x/discovery` etcd 后端（基于 etcd v3 client 接口 + 显式依赖注入）
- 为两个新后端提供独立的集成测试（使用 Docker 或 test container）
- 在 `docs/modules/x-discovery` 更新后端选择指南
- 确保新后端的依赖通过显式构造函数注入，不引入全局 init 注册

Non-goals:

- 不实现 ZooKeeper 或其他注册中心后端（优先 k8s + etcd）
- 不将 discovery 逻辑放入稳定根包

Exit criteria:

- Kubernetes 和 etcd 后端各有完整示例和 mock 测试
- `x/discovery` 模块所有后端的接口契约一致
- 与 `x/gateway` 的集成路径有 reference demo

---

## Phase 11: x/data 数据层稳定化

Status: planned

Goals:

- 明确 `x/data` 三个子包（`file`、`rw`、`sharding`）的稳定化优先级
- 为读写分离（`rw`）和分片（`sharding`）提供清晰的生产使用说明
- 收敛 `x/fileapi` 在扩展层分类中的位置

背景：

`x/data` 包含：
- `x/data/file`：租户感知文件存储实现
- `x/data/rw`：读写分离 DB 适配器
- `x/data/sharding`：分片路由（支持 JSON 配置的 cluster + 负载均衡策略）

`x/fileapi` 是独立的 HTTP 文件上传/下载 handler，目前未出现在
`docs/architecture/AGENT_FIRST_REPO_BLUEPRINT.md` 或 `specs/repo.yaml` 的扩展包清单中，
属于架构孤儿。

Planned work:

- 将 `x/fileapi` 添加到 `specs/repo.yaml` 的 `extension.paths` 清单
- 将 `x/fileapi` 添加到 `docs/architecture/AGENT_FIRST_REPO_BLUEPRINT.md`
- 明确 `x/data/rw` 的使用场景和限制（只读副本 lag 处理、failover 策略）
- 补充 `x/data/sharding` 的分片策略文档和配置示例
- 确认 `x/data` 和 `x/fileapi` 模块不依赖稳定根包之外的内容
- 为 `x/fileapi` 补充端到端 handler 测试（multipart upload + tenant context）

Non-goals:

- 不将 `x/data` 提升为稳定根包
- 不在 `x/data` 中实现业务级数据访问层

Exit criteria:

- `x/fileapi` 出现在架构蓝图和 specs 清单中
- `x/data/rw` 和 `x/data/sharding` 各有使用说明文档
- `x/fileapi` 的 handler 测试覆盖上传、下载、租户隔离三条主路径

---

## Phase 12: x/observability 与 x/gateway 测试覆盖加固

Status: planned

Goals:

- 将 `x/observability` 和 `x/gateway` 的测试覆盖从当前最小水平提升到功能级覆盖
- 保持 `x/gateway` 的熔断、负载均衡和缓存行为有显式测试保护

背景：

当前测试数量：
- `x/observability`：1 个测试文件（仅基础 config 覆盖）
- `x/gateway`：4 个测试文件（主要覆盖路由和代理转发）

`x/gateway` 含有 `cache`、`protocol`、`protocolmw` 等子包，但测试主要集中在主包级别；
`x/observability` 的 Prometheus + OpenTelemetry 集成路径几乎没有测试。

Planned work:

- 为 `x/observability` 添加 Prometheus 指标注册和导出路径的测试
- 为 `x/observability` 添加 OpenTelemetry tracer hook 的 span 属性测试
- 为 `x/gateway/cache` 添加缓存命中/未命中/失效的独立测试
- 为 `x/gateway` 负载均衡策略（round-robin、weighted）添加分发均匀性测试
- 为 `x/gateway` 的熔断（circuit breaker）添加开路/半开路状态转换测试
- 为 `x/gateway` 的 TLS 透传和协议适配器补充配置级测试

Non-goals:

- 不引入外部服务依赖（所有测试用 httptest 或 mock）
- 不更改已稳定的 public API

Exit criteria:

- `x/observability` 测试文件数量 ≥ 4，覆盖 metrics + tracing 两条路径
- `x/gateway` 负载均衡和熔断行为均有显式测试
- 所有新增测试在 `-race` 下通过

---

## Phase 13: 开发者体验与文档体系完善

Status: planned

Goals:

- 让新用户可以在 15 分钟内从 `plumego new` 到第一个可运行的服务
- 让 README_CN.md 与 README.md 内容保持同步
- 补全 `x/ai`、`x/data`、`x/tenant` 的实用示例

背景：

- `README_CN.md` 存在内容落后于 `README.md` 的风险，特别是在 Phase 5（脚手架）
  和 Phase 6（v1 API 冻结）完成之后
- `x/ai` 虽有大量功能，但缺少一个端到端可运行的最小示例（query one LLM and print response）
- `examples/multi-tenant-saas/` 在 README.md 中被引用，但实际路径位于 `reference/` 下，
  路径引用需要修正

Planned work:

- 同步 `README_CN.md` 的以下章节至与 `README.md` 一致：
  - Agent-First Workflow
  - v1 Support Matrix
  - Development Server with Dashboard
  - 配置参考表（Configuration Reference）
- 在 `reference/` 下添加 `with-ai` demo：单 provider 完成一次 chat completion 请求
- 修正 README.md 中 `examples/multi-tenant-saas/` 的路径引用
- 在 `docs/modules/x-ai` 中补充子包选择指南（何时用 provider，何时用 orchestration）
- 确认 `env.example` 包含 AI provider 相关变量（`OPENAI_API_KEY`、`ANTHROPIC_API_KEY`）

Non-goals:

- 不在 reference 示例中引入网络依赖（AI provider 调用可用 mock）
- 不修改 `README.md` 的整体结构

Exit criteria:

- `README_CN.md` 在所有章节上与 `README.md` 一致（内容等价，语言不同）
- `env.example` 包含所有 x/* 扩展需要的环境变量
- `reference/with-ai` demo 存在且 `go run .` 可执行（使用 mock provider）
- README.md 中所有 `examples/` 路径引用指向实际存在的目录

---

## Phase 14: 扩展层 API 冻结候选评估

Status: planned（依赖 Phase 7–13 完成）

Goals:

- 对使用量最高的 `x/*` 包进行稳定性评估
- 确定哪些 `x/*` 包可以在 v2 版本中升级为 GA
- 更新 `docs/DEPRECATION.md` 覆盖扩展层的稳定化流程

背景：

当前所有 `x/*` 包均为 `status: experimental`，不提供兼容性保证。
随着 `x/tenant`、`x/rest`、`x/ai` 等包逐渐成熟，需要建立一个正式的评估和
晋级机制，而不是永久维持在 experimental 状态。

Planned work:

- 定义 `x/*` 包从 experimental → stable-candidate → GA 的三段式晋级标准：
  - experimental：API 可能变更，无兼容性保证
  - stable-candidate：API 冻结期 ≥ 2 个小版本，无重大变更，测试覆盖完整
  - GA：进入主版本兼容性承诺范围
- 更新每个扩展包的 `module.yaml` 增加 `stability_track` 字段
- 对以下候选包进行首轮评估：`x/rest`、`x/websocket`、`x/webhook`、`x/scheduler`
- 更新 `docs/DEPRECATION.md` 加入扩展层晋级章节

Non-goals:

- 不在此阶段实际提升任何包的兼容性等级（仅完成评估框架）
- 不修改稳定根包的现有兼容性承诺

Exit criteria:

- 晋级标准文档存在并通过 agent 可发现
- 每个候选包的 `module.yaml` 记录了当前稳定轨迹
- `docs/DEPRECATION.md` 的扩展层章节内容完整

---

## Cross-Cutting Workstreams

### Documentation

- keep `README.md`, `README_CN.md`, `AGENTS.md`, and module primers synchronized
- ensure examples use explicit route registration and constructor-based wiring
- avoid documenting speculative abstractions before code exists

### Testing

- preserve required repo-wide gates in `AGENTS.md`
- keep targeted tests near changed behavior
- bias toward negative-path coverage for `security`, `tenant`, and `webhook`
- add smoke tests around canonical references and resource registration helpers

### Tooling

- keep checks fast enough for routine iteration
- extend repo checks only when they reduce ambiguity or drift
- prefer machine-readable rules over tribal knowledge

## Suggested Execution Order

1. Finish agent workflow checks and task cards.
2. Complete extension taxonomy convergence.
3. Deepen `x/rest` examples and route coverage.
4. Reduce stable-root migration debt.
5. Build canonical scaffolds and feature references.
6. Run formal release-readiness work toward v1.
7. Polish CLI code generation quality (Phase 7).
8. Stabilize `x/ai` sub-package tiers and clean internal drafts (Phase 8).
9. Advance `x/tenant` toward production readiness (Phase 9).
10. Add Kubernetes and etcd discovery backends (Phase 10).
11. Place `x/fileapi` in taxonomy; document `x/data` data patterns (Phase 11).
12. Harden `x/observability` and `x/gateway` test coverage (Phase 12).
13. Sync docs, add `with-ai` reference demo (Phase 13).
14. Run first x/* stability evaluation cycle (Phase 14).

## What Not to Do

- do not reintroduce component-style compatibility APIs
- do not add new broad legacy roots for short-term convenience
- do not let feature demos become the canonical app path
- do not push tenant or topology-heavy logic back into stable roots
- do not let docs get ahead of code behavior
- do not mark `x/*` packages as GA without completing the Phase 14 evaluation framework
- do not add Kubernetes or cloud-provider dependencies to stable root packages
