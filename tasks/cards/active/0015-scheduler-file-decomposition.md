# Card 0015

Priority: P2
State: active
Primary Module: x/scheduler
Owned Files:
  - x/scheduler/scheduler.go

Depends On: —

Goal:
`x/scheduler/scheduler.go` 是 1777 行的单文件 god object，包含 4 个不同职责的代码：

- **公共 API 层**（Start/Stop/AddCron/Schedule/Delay/Cancel/Pause/Resume/List/QueryJobs 等，~20 个方法）
- **执行引擎**（runLoop/dispatch/worker/execute/handleFailure/scheduleNext，~500 行）
- **队列与调度**（scheduleHeap、pushScheduleLocked、popDue、nextDue，~150 行）
- **依赖图**（wouldCreateDependencyCycleLocked/depPathExistsLocked/batch ops，~200 行）
- **辅助查询**（QueryJobs/sortJobStatuses/jobQueryMatcher/buildJobStatusLess，~150 行）

单文件无法局部阅读和修改，任何一处改动都需要 grep 全文定位上下文。

按职责拆分为多个文件（**同包内，不改变任何类型或函数签名**）：

| 新文件 | 内容 | 大约行数 |
|--------|------|----------|
| `scheduler.go` | Scheduler 结构体、Option、New()、Start/Stop | ~200 行 |
| `scheduler_api.go` | AddCron、Schedule、Delay、TriggerNow、Cancel、Pause、Resume 等公共写方法 | ~350 行 |
| `scheduler_query.go` | List、Status、QueryJobs、PruneTerminal、BatchByGroup/Tags | ~250 行 |
| `scheduler_executor.go` | runLoop、dispatch、worker、execute、handleFailure、scheduleNext | ~500 行 |
| `scheduler_queue.go` | scheduleHeap、pushScheduleLocked、popDue、nextDue、内部 queue 操作 | ~150 行 |
| `scheduler_deps.go` | wouldCreateDependencyCycleLocked、depPathExistsLocked | ~100 行 |

Scope:
- **仅移动代码到新文件，不改变任何逻辑、签名或导出名称**
- 删除 scheduler.go 中迁移出去的部分，最终 scheduler.go 仅保留结构体定义、Option 和 New()
- 保持所有函数、类型、常量在同一 package `scheduler` 内，无需修改 import
- 检查并保留所有 init-level 变量和 compile-time 断言的位置

Non-goals:
- 不改变任何公共 API（函数签名、类型名称、包名）
- 不拆分为子包
- 不优化算法或修改行为
- 不改变 metrics.go、types.go、job.go 等已独立的文件

Files:
  - x/scheduler/scheduler.go（缩减为结构体 + Option + New）
  - x/scheduler/scheduler_api.go（新建）
  - x/scheduler/scheduler_query.go（新建）
  - x/scheduler/scheduler_executor.go（新建）
  - x/scheduler/scheduler_queue.go（新建）
  - x/scheduler/scheduler_deps.go（新建）

Tests:
  - go build ./x/scheduler/...
  - go test -race -timeout 60s ./x/scheduler/...

Docs Sync: —

Done Definition:
- `wc -l x/scheduler/scheduler.go` ≤ 250 行
- `go build ./x/scheduler/...` 通过
- `go test -race ./x/scheduler/...` 通过，无测试变更（逻辑未动）

Outcome:
