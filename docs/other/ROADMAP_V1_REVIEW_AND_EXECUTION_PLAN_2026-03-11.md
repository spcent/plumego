# Plumego Roadmap V1 评审与修订执行方案

状态: Active  
日期: 2026-03-11  
评审对象: `docs/ROADMAP_V1.md`

## 1. 结论摘要

`docs/ROADMAP_V1.md` 作为“设计原则草案”有价值，但作为当前 Plumego 的 `v1.0.0` 发布路线图并不合理，原因不是它不够激进，而是它把 **v1 定义成一次破坏兼容性的体系重写**，这与仓库当前已经公开的产品定位、文档承诺、API 状态和发布节奏冲突。

更准确的判断是：

1. `ROADMAP_V1.md` 适合作为一个 **v2 架构探索提案**，不适合作为当前 `v1 GA` 计划。
2. 当前 Plumego 的合理目标不是“核心再造”，而是“**冻结稳定面、清理文档漂移、收敛实验模块边界、完成 GA 发布**”。
3. 对 v1 来说，最重要的不是新增统一入口或推倒 `core/router/contract`，而是把已经存在的标准库兼容能力整理成一个清晰、可维护、可验证的稳定产品面。

## 2. 判断依据

### 2.1 当前公开定位

从 `README.md` 可以确认 Plumego 当前定位是：

- 基于标准库的轻量 Go HTTP toolkit
- 可嵌入用户自己的 `main` 包
- `core` 是稳定主入口
- 顶层 `plumego` 包只是便捷 re-export
- 当前版本已标记为 `v1.0.0-rc.1`

这说明项目已经处于 **GA 收敛期**，不是架构推翻期。

### 2.2 当前硬约束

`AGENTS.md` 明确要求：

- 保持 `net/http` 兼容
- 主模块保持 stdlib-only
- 保持模块边界清晰
- 优先考虑 backward compatibility
- 以小步、可逆修改推进

因此，任何以“打破向后兼容性”为前提的 v1 路线图，都与当前项目治理规则直接冲突。

### 2.3 当前代码事实

仓库当前已经具备下列事实：

- `core.App` 是公开主入口，承担构造、路由注册、middleware 挂载、生命周期管理
- `core.Component` 已有稳定接口：`RegisterRoutes` / `RegisterMiddleware` / `Start` / `Stop` / `Health`
- `router.Router` 已提供显式路由注册能力：`AddRoute` / `AddRouteWithName`
- middleware 已统一在 `func(http.Handler) http.Handler`
- README、README_CN 和多份 GA 文档都已围绕 `v1.0.0-rc.1` 组织

这意味着 roadmap 中很多任务并不是“尚未开始”，而是“**方向与现状不一致**”。

## 3. 对原 Roadmap 的合理性评审

## 3.1 总体判断

Boundary Status: FAIL for `v1 GA`, WARN for `v2 proposal`

Violations:

- 把 v1 设定为破坏兼容的大重构，违背仓库当前的兼容性承诺
- 多处任务试图重定义 `core`、`router`、`contract` 的主职责，存在模块边界漂移风险
- 以“统一唯一入口”替换现有 `core.New` / `Boot` / `http.ListenAndServe` 双路径，会破坏标准库嵌入式定位
- 把 AI Gateway 作为 v1 核心目标，会冲淡 Plumego 作为通用 HTTP toolkit 的主线

Minimal Fix Plan:

1. 将 `docs/ROADMAP_V1.md` 降级为历史提案或 `v2 draft`
2. 以 GA 范围重新定义 v1：`core/router/middleware/contract/security/store`
3. 将 AI、tenant、net/mq 等扩展能力明确拆分为 GA 外围或实验轨
4. 所有后续改动按小步卡片推进，禁止跨模块推翻式重写

Required Extra Tests:

- `core`: 生命周期、组件挂载、启动/关闭回归
- `router`: static/param/group/reverse routing 回归
- `middleware`: 顺序、panic/error-path 回归
- `security`: 负面用例矩阵
- `docs`: canonical snippet compile gate

## 3.2 分阶段评审

### Phase 1: 核心再造

结论: **原则合理，落点失真，时机错误**

合理部分：

- 强调显式控制流
- 强调 DI 和零隐式全局
- 希望让 `core` 更聚焦

不合理部分：

1. “重塑 `core.App` 为纯协调器”与当前 canonical style 不一致。  
   当前 style guide 明确 `core` 负责应用构造、路由注册入口、middleware attachment、server startup。把这些从 `core` 剥离，会削弱当前最重要的学习路径。

2. “统一应用启动入口 `plumego.Run(http.Handler, ...Option)`”不符合现有定位。  
   Plumego 现在的优势之一是既可 `app.Boot()`，也可作为 `http.Handler` 交给标准库 server。强行统一成 `Run(...)` 反而减少嵌入灵活性。

3. 提议的 `Component` 接口不如当前接口贴近边界。  
   当前 `RegisterRoutes` / `RegisterMiddleware` 分离得更清楚；把注册压缩成 `Register(app *App)` 会让组件更容易跨边界碰 `core` 内部状态。

4. “彻底根除全局状态”方向对，但不应通过大面积 API 替换推进。  
   这里适合做兼容性收口，而不是破坏式重构。

判断: **应改为 v1.x 内部收敛项，而不是 v1 阶段性重写项**

### Phase 2: 路由与中间件重塑

结论: **大部分已经是既有方向，不需要以重写形式推进**

合理部分：

- 显式路由注册是正确方向
- 标准 middleware 签名是正确方向
- 反对 context service-locator 完全正确

不合理部分：

1. “移除所有复杂的链式 API 和隐式路由分组”过度。  
   当前 README 已明确把 groups 作为正式能力暴露。问题不在 group 本身，而在 group 是否承担了超出“路径前缀 + 共享 transport middleware”的职责。

2. “设计全新 `plumego.Context`”并非当前优先级。  
   当前 canonical handler 仍是 `func(http.ResponseWriter, *http.Request)`。`contract.Ctx` 应保持为辅助能力，而不是再定义一轮主路径。

判断: **应转化为“契约收敛 + 文档统一 + 回归测试补齐”**

### Phase 3: AI Gateway 与工具链

结论: **方向有价值，但不应定义 v1 主线**

合理部分：

- AI provider 抽象、tool calling、SSE 都符合仓库已有能力沉淀
- 能形成 Plumego 的差异化能力层

不合理部分：

1. 将 AI Gateway 设为 v1 核心阶段，会让核心定位从 “standard library HTTP toolkit” 偏向 “AI application framework”。
2. AI 能力变化速度快、接口不稳定，更适合作为实验或次级稳定面，而不是 v1 的发布阻塞项。

判断: **保留为 v1.1/v1.2 能力路线，或继续以 experimental/capability track 推进**

### Phase 4: 生产力与可观测性

结论: **应保留，但必须保持可选、模块化、非侵入**

合理部分：

- 统一 observability 组件有明确工程价值
- security 组合件有明确用户价值
- benchmark 进入门禁很合理

边界提醒：

1. observability 不能让主模块引入非 stdlib 依赖污染主路径。
2. security 应保留为 transport middleware / helper 组合，不要让 middleware 承担业务策略。
3. benchmark 门禁应分级，避免每次普通改动都承担高昂成本。

判断: **可作为 GA 之后的增强计划，或以 optional component 形式前置一部分**

### Phase 5: 文档、示例、发布

结论: **这是当前 v1 最合理的主战场**

合理部分：

- 文档、示例、质量门禁、发布脚本都直接服务 GA
- 这些工作与当前仓库已有 `V1_GA_EXECUTION_PLAN`、`V1_RELEASE_RUNBOOK` 一致

需要修正的地方：

1. 不应“基于未来新架构”重写 `AGENTS.md`。应基于当前稳定边界更新。
2. `CANONICAL_STYLE_GUIDE` 也应服务当前 API 冻结，而不是为未落地的新 API 立法。

判断: **应提升为当前执行优先级最高的阶段**

## 4. 修订后的 v1 目标定义

## 4.1 v1 的正确目标

Plumego v1.0.0 的目标应定义为：

> 交付一个以 `core/router/middleware/contract/security/store` 为稳定面的、标准库兼容的、文档和示例可验证的 Go HTTP toolkit；将 `tenant/*`、`net/mq/*`、部分 AI 能力明确标识为 experimental 或次级稳定面。

## 4.2 v1 不应该做的事

以下内容不应作为 v1 发布阻塞项：

- 推翻 `core.App` 的主入口地位
- 用 `plumego.Run` 替代现有双入口模型
- 取消 group/router 现有能力
- 重做 `contract.Context` 主路径
- 让 AI Gateway 成为 v1 发布主线
- 大规模跨模块重构以追求“更纯”的理论架构

## 5. 修订后的执行方案

## 5.1 总体分轨

将工作拆成三条轨道：

### Track A: v1 GA 阻塞项

目标: 最终发布 `v1.0.0`

范围:

- `core`
- `router`
- `middleware`
- `contract`
- `security`
- `store`
- canonical docs / examples / release gates

### Track B: v1.x 稳定化增强

目标: 不破坏兼容前提下继续提升体验

范围:

- observability optional component
- security preset/component 化
- benchmark 基线与 CI 分级
- 全局状态兼容层收口

### Track C: v2 / Experimental 架构探索

目标: 用 ADR / design doc 验证更激进的架构改动

范围:

- `plumego.Run` 是否需要
- `core.App` 是否进一步瘦身
- `contract.Ctx` 是否继续弱化
- AI Gateway 是否形成一级子产品线

## 5.2 Phase-by-Phase 执行顺序

### Phase A: 冻结 v1 GA 范围

目标:

- 明确 GA 稳定模块与 experimental 模块
- 停止把 v1 定义为重写工程

交付物:

- 统一的 scope 文档
- `ROADMAP_V1.md` 的状态说明或引用修订文档
- README / README_CN 中的模块稳定性矩阵

验收标准:

- 所有面向用户的入口文档，对 v1 范围描述一致
- 没有文档继续把 v1 表述为 breaking rewrite

### Phase B: 清理 canonical API 漂移

目标:

- 文档、示例、注释全部与当前 API 一致

交付物:

- README / README_CN / getting-started / core docs / router docs / middleware docs 的统一示例
- snippet compile gate
- drift scan 报告

验收标准:

- canonical docs 全部通过编译级校验
- 已移除 API 不再出现在 canonical 文档

### Phase C: 核心稳定面回归补强

目标:

- 为 `core/router/middleware/security` 建立完整回归信心

交付物:

- core lifecycle/component tests
- router static/param/group/reverse-routing tests
- middleware ordering/error-path tests
- security negative tests

验收标准:

- `go test -timeout 20s ./...`
- `go test -race -timeout 60s ./...`
- `go vet ./...`
- 关键路径 benchmark 可运行

### Phase D: experimental 边界标记

目标:

- 避免用户把实验模块误判为 GA 承诺

交付物:

- `tenant/*`、`net/mq/*`、高变化 AI 子模块的 experimental 标注
- README 与模块文档一致

验收标准:

- 所有实验模块入口文档都带稳定性说明
- 没有 README 示例把实验模块包装成默认主路径

### Phase E: GA 发布准备

目标:

- 完成最终发布而非继续设计漂移

交付物:

- release readiness script
- release runbook
- final changelog / release notes
- `v1.0.0` tag checklist

验收标准:

- 本地与 CI 门禁全部通过
- 发布说明覆盖稳定面、实验面、已知限制、升级说明

## 6. 任务卡片拆解

以下卡片遵循单模块优先、少文件、可逆、可单独提交的规则。

### Now

Goal:
- 将 `docs/ROADMAP_V1.md` 从“v1 发布路线图”纠偏为“历史提案/非当前 GA 计划”，避免继续误导贡献者

Scope:
- `docs/ROADMAP_V1.md`
- `docs/other/ROADMAP_V1_REVIEW_AND_EXECUTION_PLAN_2026-03-11.md`

Non-goals:
- 不修改核心代码
- 不同时重写 README 全量内容

Files:
- `docs/ROADMAP_V1.md`
- `docs/other/ROADMAP_V1_REVIEW_AND_EXECUTION_PLAN_2026-03-11.md`

Tests:
- 文档一致性人工复核
- `rg -n "打破向后兼容性|breaking"` 在 canonical 文档中复扫

Docs Sync:
- 如修改 `ROADMAP_V1.md`，同步在相关 GA 文档中补一个引用即可

Done Definition:
- 贡献者打开 `ROADMAP_V1.md` 时，不会再把它理解为当前 v1 GA 官方计划

### Queue 1

Goal:
- 清理 router / middleware canonical 文档漂移

Scope:
- `docs/modules/router/*`
- `docs/modules/middleware/*`

Non-goals:
- 不引入新 router API

Files:
- 最多 5 个高流量文档

Tests:
- `bash scripts/check-doc-snippets-compile.sh`
- `go test -timeout 20s ./router ./middleware/...`

Docs Sync:
- `README.md`
- `README_CN.md`

Done Definition:
- 高流量路由/中间件文档不再出现陈旧 API

### Queue 2

Goal:
- 收敛 `core` 生命周期与组件文档

Scope:
- `docs/modules/core/*`
- `core/*_test.go`

Non-goals:
- 不修改 `core.App` 主体架构

Files:
- 文档 3-4 个
- 测试 1-2 个

Tests:
- `go test -timeout 20s ./core`
- `go test -race -timeout 60s ./core`

Docs Sync:
- `README.md`
- `README_CN.md`

Done Definition:
- `core` 文档与测试对当前 API 的表述一致

### Queue 3

Goal:
- 给 experimental 模块补齐稳定性标签和入口提示

Scope:
- `docs/modules/tenant/*`
- `docs/tenant/*`
- `docs/modules/mq/*` 或相关入口

Non-goals:
- 不推进 tenant/mq 新功能开发

Files:
- 每次 3-5 个入口文档

Tests:
- `rg -n "experimental|实验"` 一致性检查
- 相关模块基础测试

Docs Sync:
- `README.md`
- `README_CN.md`

Done Definition:
- 用户可以在首屏判断哪些模块属于 GA，哪些不是

## 7. 风险与缓解

### 风险 1: 路线图与真实状态继续分叉

表现:
- 新贡献者按 `ROADMAP_V1.md` 发起破坏式重构

缓解:
- 明确标记该文档状态
- 以本执行方案和 GA plan 作为唯一当前路线

### 风险 2: AI 能力线抢占核心主线

表现:
- v1 发布被 AI Gateway 设计问题拖延

缓解:
- 将 AI track 改为 v1.x / experimental
- 明确发布阻塞只围绕核心稳定面

### 风险 3: 文档收敛慢于代码收敛

表现:
- API 实际可用，但 docs 不可信，用户误用

缓解:
- 把 snippet compile gate 设为固定门禁
- 先修 canonical，再修历史/规划文档

## 8. 最终建议

### 建议 1

不要按当前 `docs/ROADMAP_V1.md` 继续推进 v1。  
它应被重新定性为“架构理想蓝图”或“v2 draft”，而不是当前发布计划。

### 建议 2

以 `docs/other/V1_GA_EXECUTION_PLAN_2026-03-10.md` 为当前主计划，以本文档作为对 `ROADMAP_V1.md` 的解释和纠偏。

### 建议 3

当前最优执行策略是：

1. 冻结 v1 GA 范围
2. 清理 canonical docs 漂移
3. 补强核心稳定面测试
4. 标注 experimental 边界
5. 运行 release gate 并发布 `v1.0.0`

这条路线最符合 Plumego 作为“标准库兼容、轻量、可嵌入、显式控制流”的定位。
