# Sprint 5: Distributed Workflows - 详细规划

## 目标

实现分布式工作流执行系统，支持跨多个节点的 AI 代理协作，提供可靠的任务分发、状态管理和故障恢复能力。

---

## 核心功能模块

### 1. 工作流持久化 (Workflow Persistence)

**目标**: 将工作流定义和执行状态持久化到存储，支持中断恢复

**实现文件**:
- `persistence.go` - 持久化接口和实现
- `persistence_test.go` - 测试

**功能点**:
```go
type WorkflowPersistence interface {
    // 保存工作流定义
    SaveWorkflow(ctx context.Context, wf *orchestration.Workflow) error

    // 加载工作流定义
    LoadWorkflow(ctx context.Context, id string) (*orchestration.Workflow, error)

    // 保存执行状态快照
    SaveSnapshot(ctx context.Context, snapshot *ExecutionSnapshot) error

    // 加载执行状态快照
    LoadSnapshot(ctx context.Context, executionID string) (*ExecutionSnapshot, error)

    // 保存步骤结果
    SaveStepResult(ctx context.Context, executionID string, stepIndex int, result *orchestration.AgentResult) error

    // 清理过期数据
    CleanupExpired(ctx context.Context, before time.Time) (int, error)
}

type ExecutionSnapshot struct {
    ExecutionID   string
    WorkflowID    string
    State         map[string]any
    CompletedSteps []int           // 已完成的步骤索引
    CurrentStep   int              // 当前执行的步骤
    Results       []*orchestration.AgentResult
    Status        ExecutionStatus  // running/paused/completed/failed
    CreatedAt     time.Time
    UpdatedAt     time.Time
}

type ExecutionStatus string
const (
    StatusPending   ExecutionStatus = "pending"
    StatusRunning   ExecutionStatus = "running"
    StatusPaused    ExecutionStatus = "paused"
    StatusCompleted ExecutionStatus = "completed"
    StatusFailed    ExecutionStatus = "failed"
)
```

**实现方案**:
- 使用 `store/kv` 作为后端存储
- 工作流定义序列化为 JSON
- 执行快照每步保存，支持断点续传
- TTL 自动清理过期执行记录

---

### 2. 分布式任务队列 (Distributed Task Queue)

**目标**: 将工作流步骤分发到多个 worker 节点执行

**实现文件**:
- `queue.go` - 任务队列接口和实现
- `worker.go` - Worker 节点实现
- `queue_test.go`, `worker_test.go` - 测试

**功能点**:
```go
type TaskQueue interface {
    // 发布任务到队列
    Enqueue(ctx context.Context, task *WorkflowTask) error

    // Worker 订阅任务
    Subscribe(ctx context.Context, workerID string, handler TaskHandler) error

    // 更新任务状态
    UpdateTaskStatus(ctx context.Context, taskID string, status TaskStatus, result *TaskResult) error

    // 获取任务状态
    GetTaskStatus(ctx context.Context, taskID string) (*TaskStatus, error)
}

type WorkflowTask struct {
    ID            string
    ExecutionID   string
    WorkflowID    string
    StepIndex     int
    Step          orchestration.Step  // 序列化的步骤
    WorkflowState map[string]any      // 当前工作流状态
    Priority      int                  // 优先级
    CreatedAt     time.Time
    Timeout       time.Duration
}

type TaskHandler func(ctx context.Context, task *WorkflowTask) (*TaskResult, error)

type TaskResult struct {
    TaskID      string
    AgentResult *orchestration.AgentResult
    UpdatedState map[string]any
    Error       error
}

type TaskStatus struct {
    TaskID      string
    Status      string  // pending/running/completed/failed
    WorkerID    string
    StartedAt   time.Time
    CompletedAt time.Time
    Retries     int
    LastError   string
}
```

**实现方案**:
- 基于 `net/mq` 消息队列
- 使用 priority queue 支持优先级
- 自动负载均衡（多个 worker 竞争消费）
- ACK 机制确保任务不丢失

---

### 3. 分布式执行引擎 (Distributed Execution Engine)

**目标**: 协调分布式工作流执行，管理任务生命周期

**实现文件**:
- `engine.go` - 分布式执行引擎
- `coordinator.go` - 执行协调器
- `engine_test.go`, `coordinator_test.go` - 测试

**功能点**:
```go
type DistributedEngine struct {
    localEngine   *orchestration.Engine
    taskQueue     TaskQueue
    persistence   WorkflowPersistence
    coordinator   *ExecutionCoordinator
}

// 执行工作流（异步）
func (e *DistributedEngine) ExecuteAsync(
    ctx context.Context,
    workflowID string,
    initialState map[string]any,
    options ExecutionOptions,
) (executionID string, err error)

// 执行工作流（同步）
func (e *DistributedEngine) ExecuteSync(
    ctx context.Context,
    workflowID string,
    initialState map[string]any,
    options ExecutionOptions,
) ([]*orchestration.AgentResult, error)

// 恢复中断的工作流
func (e *DistributedEngine) Resume(
    ctx context.Context,
    executionID string,
) error

// 暂停执行
func (e *DistributedEngine) Pause(
    ctx context.Context,
    executionID string,
) error

// 获取执行状态
func (e *DistributedEngine) GetExecutionStatus(
    ctx context.Context,
    executionID string,
) (*ExecutionSnapshot, error)

type ExecutionOptions struct {
    Distributed bool          // 是否使用分布式执行
    Timeout     time.Duration
    RetryPolicy *RetryPolicy
    Checkpoints bool          // 是否启用检查点
}
```

**执行流程**:
1. 保存工作流执行快照（初始状态）
2. 遍历步骤，根据配置决定本地/远程执行
3. 对于远程步骤，创建任务并发布到队列
4. 等待任务完成，更新快照
5. 支持中断恢复：从快照加载，跳过已完成步骤

---

### 4. Worker 节点实现 (Worker Node)

**目标**: 独立运行的 worker 进程，从队列消费任务并执行

**实现文件**:
- `worker.go` - Worker 实现
- `worker_pool.go` - Worker 池管理
- `worker_test.go` - 测试

**功能点**:
```go
type Worker struct {
    ID          string
    queue       TaskQueue
    executor    TaskExecutor  // 执行器
    concurrency int           // 并发数
    heartbeat   time.Duration
}

// 启动 worker
func (w *Worker) Start(ctx context.Context) error

// 停止 worker
func (w *Worker) Stop(ctx context.Context) error

// Worker 池
type WorkerPool struct {
    workers []*Worker
    config  PoolConfig
}

type PoolConfig struct {
    WorkerCount  int
    Concurrency  int  // 每个 worker 的并发数
    QueueTopic   string
}

func (p *WorkerPool) Start(ctx context.Context) error
func (p *WorkerPool) Stop(ctx context.Context) error
func (p *WorkerPool) Scale(count int) error
```

**特性**:
- 支持优雅启动/关闭
- 心跳检测
- 自动重连
- 并发控制

---

### 5. 远程步骤类型 (Remote Step Types)

**目标**: 支持远程执行的步骤类型

**实现文件**:
- `remote_step.go` - 远程步骤包装器
- `remote_step_test.go` - 测试

**功能点**:
```go
// 远程执行步骤包装器
type RemoteStep struct {
    Step      orchestration.Step
    Timeout   time.Duration
    RetryPolicy *RetryPolicy
}

func (r *RemoteStep) Execute(ctx context.Context, wf *orchestration.Workflow) (*orchestration.AgentResult, error)
func (r *RemoteStep) Name() string

// 分布式并行步骤
type DistributedParallelStep struct {
    StepName   string
    Agents     []*orchestration.Agent
    InputFn    func(state map[string]any) []string
    OutputKeys []string
    MaxWorkers int  // 最大并发 worker 数
}
```

**实现逻辑**:
- RemoteStep 将任务发送到队列，等待完成
- DistributedParallelStep 并发发送多个任务到队列
- 支持超时和重试

---

### 6. 故障恢复 (Fault Tolerance)

**目标**: 自动处理节点故障、任务超时等异常

**实现文件**:
- `recovery.go` - 故障恢复逻辑
- `recovery_test.go` - 测试

**功能点**:
```go
type RecoveryManager struct {
    persistence WorkflowPersistence
    taskQueue   TaskQueue
}

// 扫描并恢复失败的任务
func (rm *RecoveryManager) RecoverStaleTasks(ctx context.Context) (int, error)

// 重试策略
type RetryPolicy struct {
    MaxRetries  int
    InitialDelay time.Duration
    MaxDelay     time.Duration
    Multiplier   float64  // 指数退避倍数
}

// 任务超时检测
func (rm *RecoveryManager) MonitorTimeouts(ctx context.Context, interval time.Duration) error
```

**策略**:
- 任务超时自动重试（基于 ACK 超时）
- 执行快照定期保存
- Worker 崩溃后任务自动重新调度
- 支持死信队列（DLQ）

---

### 7. 监控与可观测性 (Observability)

**目标**: 提供分布式执行的可见性

**实现文件**:
- `instrumentation.go` - 指标和追踪
- `instrumentation_test.go` - 测试

**功能点**:
```go
type DistributedMetrics struct {
    // 执行指标
    ExecutionsStarted   int64
    ExecutionsCompleted int64
    ExecutionsFailed    int64

    // 任务队列指标
    TasksEnqueued       int64
    TasksCompleted      int64
    TasksFailed         int64
    QueueDepth          int64

    // Worker 指标
    ActiveWorkers       int64
    TasksProcessing     int64

    // 性能指标
    AvgTaskDuration     time.Duration
    AvgExecutionTime    time.Duration
}

type InstrumentedDistributedEngine struct {
    engine    *DistributedEngine
    collector metrics.MetricsCollector
}

func (ie *InstrumentedDistributedEngine) ExecuteAsync(...) (string, error)
// 自动收集执行时间、成功/失败计数等指标
```

**集成**:
- 与现有 `metrics` 包集成
- 支持 Prometheus 导出
- 结构化日志

---

## 文件结构

```
ai/distributed/
├── engine.go                    # 分布式执行引擎
├── engine_test.go
├── coordinator.go               # 执行协调器
├── coordinator_test.go
├── persistence.go               # 工作流持久化
├── persistence_test.go
├── queue.go                     # 任务队列接口
├── queue_test.go
├── worker.go                    # Worker 实现
├── worker_test.go
├── worker_pool.go               # Worker 池
├── worker_pool_test.go
├── remote_step.go               # 远程步骤包装器
├── remote_step_test.go
├── recovery.go                  # 故障恢复
├── recovery_test.go
├── instrumentation.go           # 监控指标
├── instrumentation_test.go
├── serialization.go             # 工作流序列化
├── serialization_test.go
├── integration_test.go          # 集成测试
├── example_test.go              # 示例测试
└── README.md                    # 文档
```

预计：**~2,500 LOC**，**60+ 测试**

---

## 实现优先级

### P0 - 核心功能（必须完成）
1. ✅ WorkflowPersistence - 持久化接口和 KV 实现
2. ✅ TaskQueue - 基于 net/mq 的任务队列
3. ✅ Worker - 基础 worker 实现
4. ✅ DistributedEngine - 分布式执行引擎
5. ✅ RemoteStep - 远程步骤包装器

### P1 - 增强功能（建议完成）
6. ✅ RecoveryManager - 故障恢复和重试
7. ✅ WorkerPool - Worker 池管理
8. ✅ DistributedParallelStep - 分布式并行执行
9. ✅ InstrumentedEngine - 监控指标

### P2 - 高级功能（可选）
10. ⏸️ WorkflowVersioning - 工作流版本管理
11. ⏸️ DistributedLock - 分布式锁（避免重复执行）
12. ⏸️ WorkflowScheduling - 定时执行集成

---

## 技术依赖

- **orchestration** - 基础工作流编排
- **net/mq** - 消息队列（任务分发）
- **store/kv** - 持久化存储
- **scheduler** - 定时任务（可选集成）
- **metrics** - 指标收集
- **pubsub** - 事件通知

---

## 测试策略

### 单元测试
- 每个模块独立测试
- Mock 外部依赖（队列、存储）
- 覆盖率 > 80%

### 集成测试
- 完整的分布式工作流执行
- 多 worker 并发测试
- 故障注入测试（worker 崩溃、超时）

### 性能测试
- 高并发任务处理
- 大规模工作流（100+ 步骤）
- 内存和 CPU 使用监控

---

## 示例用例

```go
// 1. 创建分布式引擎
cfg := distributed.DefaultConfig()
cfg.QueueTopic = "ai-workflows"
cfg.WorkerCount = 4
cfg.EnablePersistence = true

engine := distributed.NewDistributedEngine(cfg)
defer engine.Close()

// 2. 定义工作流
workflow := orchestration.NewWorkflow("workflow-1", "Distributed Analysis", "")
workflow.AddStep(&orchestration.SequentialStep{
    StepName: "extract",
    Agent:    extractorAgent,
    InputFn:  func(state map[string]any) string { return state["input"].(string) },
    OutputKey: "extracted",
})

// 使用分布式并行步骤
workflow.AddStep(&distributed.DistributedParallelStep{
    StepName: "parallel-analysis",
    Agents: []*orchestration.Agent{
        sentimentAgent,
        entityAgent,
        summaryAgent,
    },
    InputFn: func(state map[string]any) []string {
        text := state["extracted"].(string)
        return []string{text, text, text}
    },
    OutputKeys: []string{"sentiment", "entities", "summary"},
    MaxWorkers: 3,
})

// 3. 异步执行
executionID, err := engine.ExecuteAsync(ctx, "workflow-1", map[string]any{
    "input": "Large document...",
}, distributed.ExecutionOptions{
    Distributed: true,
    Timeout:     10 * time.Minute,
    Checkpoints: true,
})

// 4. 监控执行状态
status, _ := engine.GetExecutionStatus(ctx, executionID)
fmt.Printf("Status: %s, Completed: %d/%d\n",
    status.Status, len(status.CompletedSteps), len(workflow.Steps))

// 5. 故障恢复
if status.Status == distributed.StatusFailed {
    engine.Resume(ctx, executionID)
}
```

---

## 性能目标

- **任务分发延迟**: < 10ms (P99)
- **执行吞吐量**: > 1000 任务/秒（4 workers）
- **故障恢复时间**: < 5 秒
- **检查点开销**: < 5% 执行时间
- **内存占用**: < 100 MB（单 worker）

---

## 风险与挑战

1. **序列化复杂度** - 工作流步骤包含函数指针，需要特殊处理
   - 解决方案：使用注册表模式，预定义步骤类型

2. **分布式一致性** - 多节点执行状态同步
   - 解决方案：单点协调器 + 持久化快照

3. **性能瓶颈** - 高频快照保存影响性能
   - 解决方案：异步保存 + 批量写入

4. **调试困难** - 分布式追踪复杂
   - 解决方案：结构化日志 + 执行 ID 传播

---

## 下一步行动

1. ✅ 创建 `ai/distributed/` 目录
2. ✅ 实现 P0 核心功能（持久化、队列、引擎）
3. ✅ 编写单元测试（目标 80% 覆盖率）
4. ✅ 实现 P1 增强功能（恢复、池、监控）
5. ✅ 集成测试和性能测试
6. ✅ 文档和示例
7. ⏸️ 可选：P2 高级功能

---

**预计完成时间**: 核心功能 (P0 + P1) 约 **2-3 小时**

**预计测试数量**: **60-80 个测试**

**预计代码行数**: **~2,500 LOC**
