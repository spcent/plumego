# pubsub 包重构计划

> 依据 `docs/CANONICAL_STYLE_GUIDE.md`，不考虑向后兼容，按最佳实践重新规划。

---

## 一、现状诊断

### 1.1 文件规模

`pubsub/` 目前有 **50+ Go 文件**，单包承载了六个不同关注点：

| 关注点 | 文件 |
|--------|------|
| 核心投递 | `pubsub.go`, `types.go`, `errors.go`, `options.go`, `shard.go`, `ringbuffer.go`, `workerpool.go`, `message_clone.go` |
| 消息处理扩展 | `dedup.go`, `ordering.go`, `priority.go`, `ratelimit.go`, `scheduler.go`, `replay.go` |
| 可靠性附件 | `ack.go`, `dlq.go`, `backpressure.go`, `request_reply.go` |
| 企业级功能 | `distributed.go`, `multitenant.go`, `consumergroup.go`, `persistence.go`, `audit.go` |
| 可观测性 | `metrics.go`, `metrics_enhanced.go`, `prometheus.go`, `health.go` |
| 模式匹配 | `mqtt_pattern.go`, `history.go` |

这违反了 Style Guide §2（Package Roles）的单一职责原则。

### 1.2 接口膨胀

`types.go` 定义了六层深的继承链，`InProcPubSub` 一个结构体满足全部：

```
PubSub
  └─ PatternPubSub
       └─ ContextPubSub
            └─ BatchPubSub
                 └─ DrainablePubSub
MQTTPubSub    (独立分支)
HistoryPubSub (独立分支)
RequestReplyPubSub (独立分支)
```

`InProcPubSub` 是典型的 **God Object**，实现了所有接口的全部方法。

### 1.3 功能标志反模式

`Config` 用布尔字段在运行时切换能力：

```go
EnableScheduler    bool
EnableTTL          bool
EnableRingBuffer   bool
EnableRequestReply bool
EnableHistory      bool
```

这导致代码路径分叉，每处 "功能未启用时返回 ErrXxxDisabled" 都是一个隐式运行时状态，违反 Style Guide §1（explicit over implicit）。

### 1.4 双重错误模型

`errors.go` 同时维护：
- 哨兵错误（`ErrClosed`、`ErrInvalidTopic`……共 17 个）
- 结构化 `PubSubError`（含 `ErrorCode`）
- 复杂的 `Is()` 双向映射（用于"向后兼容"）

两套模型并行存在，`determineErrorCode()` 和 `Is()` 都在做相同的码→sentinel 映射，互相重复。

### 1.5 `policyName` 重复

`types.go` 中 `BackpressurePolicy.String()` 与 `metrics.go` 中的 `policyName()` 功能完全相同，却共存。

### 1.6 两套 Backpressure 命名空间

| 类型 | 位置 | 作用 |
|------|------|------|
| `BackpressurePolicy` | `types.go` | 单个 subscriber 的缓冲满策略 |
| `BackpressurePropagationPolicy` | `backpressure.go` | 全局系统级背压传播策略 |

命名混淆，概念层次不清。

### 1.7 Metrics 三分裂

- `metrics.go` — 内部原子计数器
- `metrics_enhanced.go` — 延迟直方图（与前者独立实现）
- `prometheus.go` — Prometheus 导出

三套实现散乱，无统一收口。

### 1.8 `BackpressureController` 硬依赖

```go
func NewBackpressureController(ps *InProcPubSub, ...) *BackpressureController
```

控制器持有具体类型指针，造成循环耦合隐患，且无法替换实现。

### 1.9 `ContextPubSub` 接口冗余

`ContextPubSub` 仅是为每个方法加了 `ctx context.Context` 参数前缀的重复接口。Context 应在调用方按需传入，而非在接口层面独立存在一个"上下文版本"。

### 1.10 `Hooks` 暴露在 Config 中

`Config.Hooks` 将生命周期回调直接嵌入配置结构，混淆了"构建参数"与"可观测性探针"两类关注点。

---

## 二、重构目标

1. **单一 Broker 接口**：最小化、稳定、可测试
2. **能力按包隔离**：可选特性独立子包，零依赖核心
3. **单一错误类型**：`Error` 结构体 + `ErrorCode`，移除哨兵层
4. **无功能标志**：特性通过构造器组合，而非 `Enable*` 布尔值
5. **消除重复**：`policyName` 统一走 `.String()`，metrics 合一收口
6. **接口解耦**：`BackpressureController` 依赖接口不依赖实现
7. **可观测性独立**：Prometheus 导出移出核心包

---

## 三、新包结构

```
pubsub/
├── broker.go           # Broker 接口 + New() 构造器
├── types.go            # Message, SubOptions, Subscription 接口
├── errors.go           # Error 结构体 + ErrorCode（统一单一模型）
├── options.go          # Config + functional options（无 Enable* 字段）
├── pubsub.go           # InProcBroker 内部实现
├── shard.go            # 内部分片（unexported）
├── ringbuffer.go       # 内部环形缓冲（unexported）
├── workerpool.go       # 内部工作池（unexported）
├── metrics.go          # 内部计数器 + 延迟直方图（合并 enhanced）
├── backpressure.go     # BackpressurePolicy（subscriber 级），移除 Controller
└── mqtt_pattern.go     # MQTT 模式匹配（unexported helper）

pubsub/history/
├── history.go          # HistoryStore 接口 + 内存实现
└── history_test.go

pubsub/ack/
├── ack.go              # AckableSubscription + DeadLetterQueue
└── ack_test.go

pubsub/dlq/
├── dlq.go              # DLQ（RetryStrategy、DLQMessage）
└── dlq_test.go

pubsub/ordering/
├── ordering.go         # OrderLevel + OrderedBroker 包装器
└── ordering_test.go

pubsub/dedup/
├── dedup.go            # Deduplicator
└── dedup_test.go

pubsub/ratelimit/
├── ratelimit.go        # RateLimiter
└── ratelimit_test.go

pubsub/scheduler/
├── scheduler.go        # DelayedPublisher
└── scheduler_test.go

pubsub/replay/
├── replay.go           # ReplayBroker
└── replay_test.go

pubsub/priority/
├── priority.go         # PriorityBroker
└── priority_test.go

pubsub/requestreply/
├── requestreply.go     # RequestReplyBroker
└── requestreply_test.go

pubsub/pressure/
├── pressure.go         # BackpressureController（依赖 Broker 接口）
└── pressure_test.go

pubsub/consumergroup/
├── consumergroup.go    # ConsumerGroupManager
└── consumergroup_test.go

pubsub/distributed/
├── distributed.go      # DistributedBroker（HTTP 集群）
└── distributed_test.go

pubsub/multitenant/
├── multitenant.go      # TenantBroker（配额隔离）
└── multitenant_test.go

pubsub/audit/
├── audit.go            # AuditLogger（SHA-256 链）
└── audit_test.go

pubsub/persistence/
├── persistence.go      # WAL + Snapshot
└── persistence_test.go

pubsub/prometheus/
├── prometheus.go       # Prometheus 指标导出器
└── prometheus_test.go
```

---

## 四、各阶段详细计划

### Phase 1 — 核心接口精简（`broker.go` + `types.go`）

#### 目标

用单一 `Broker` 接口替代现有六个继承接口，能力接口扁平化、可选化。

#### 新接口设计

```go
// Broker 是核心发布订阅接口。最小化、无上下文变体。
type Broker interface {
    Publish(topic string, msg Message) error
    Subscribe(topic string, opts SubOptions) (Subscription, error)
    Close() error
}

// PatternSubscriber 扩展 Broker，支持 glob 模式订阅。
type PatternSubscriber interface {
    Broker
    SubscribePattern(pattern string, opts SubOptions) (Subscription, error)
}

// MQTTSubscriber 扩展 Broker，支持 MQTT 风格模式订阅。
type MQTTSubscriber interface {
    Broker
    SubscribeMQTT(pattern string, opts SubOptions) (Subscription, error)
}

// BatchPublisher 扩展 Broker，支持批量发布。
type BatchPublisher interface {
    Broker
    PublishBatch(topic string, msgs []Message) error
    PublishMulti(msgs map[string][]Message) error
}

// Drainable 扩展 Broker，支持优雅排空。
type Drainable interface {
    Broker
    Drain(ctx context.Context) error
}
```

**删除**：
- `ContextPubSub`：context 参数在调用方按需传入，不需要为此独立定义接口
- 原有的深层继承链

`InProcBroker`（原 `InProcPubSub`）实现 `PatternSubscriber`、`MQTTSubscriber`、`BatchPublisher`、`Drainable`，但调用方依赖最小化接口。

#### `Subscription` 接口不变

保留 `C()`、`Cancel()`、`ID()`、`Topic()`、`Stats()`、`Done()`，这是稳定的最小集合。

---

### Phase 2 — 错误模型统一（`errors.go`）

#### 目标

一套错误类型，清晰的 `errors.Is` / `errors.As` 语义，移除双轨并行。

#### 新错误设计

```go
// ErrorCode 是机器可读错误代码。
type ErrorCode string

const (
    ErrCodeClosed         ErrorCode = "PUBSUB_CLOSED"
    ErrCodeInvalidTopic   ErrorCode = "INVALID_TOPIC"
    ErrCodeInvalidPattern ErrorCode = "INVALID_PATTERN"
    ErrCodeInvalidOptions ErrorCode = "INVALID_OPTIONS"
    ErrCodeBufferFull     ErrorCode = "BUFFER_FULL"
    ErrCodeTimeout        ErrorCode = "TIMEOUT"
    ErrCodeBackpressure   ErrorCode = "BACKPRESSURE"
    ErrCodeNotFound       ErrorCode = "NOT_FOUND"
    ErrCodeDuplicate      ErrorCode = "DUPLICATE"
    ErrCodeExpired        ErrorCode = "EXPIRED"
    ErrCodeInternal       ErrorCode = "INTERNAL"
)

// Error 是 pubsub 包唯一的错误类型。
type Error struct {
    Code    ErrorCode
    Op      string
    Topic   string
    Message string
    Cause   error
}

func (e *Error) Error() string { ... }
func (e *Error) Unwrap() error { return e.Cause }

// Is 仅比较 Code，不做哨兵映射。
func (e *Error) Is(target error) bool {
    var t *Error
    if errors.As(target, &t) {
        return e.Code == t.Code
    }
    return false
}
```

**删除**：
- 17 个哨兵变量（`ErrClosed`、`ErrPublishToClosed`、`ErrSubscribeToClosed`……）
- `PubSubError`（重命名为 `Error`）
- `determineErrorCode()`（不再需要反向映射）
- `WrapError()`（直接构造 `&Error{}`）
- `IsClosedError()`、`IsTransientError()`、`IsPermanentError()`（调用方用 `errors.As + Code` 替代）

**新增工厂函数**（替代现有的 5 个构造函数）：

```go
func newError(code ErrorCode, op, topic, message string, cause error) *Error
```

---

### Phase 3 — Config 清理（`options.go`）

#### 目标

Config 只包含"运行时参数"，移除"功能开关"布尔值；`Hooks` 移至独立观察接口。

#### 新 Config

```go
type Config struct {
    ShardCount          int
    DefaultBufferSize   int
    DefaultPolicy       BackpressurePolicy
    DefaultBlockTimeout time.Duration
    WorkerPoolSize      int
    MetricsCollector    metrics.MetricsCollector
    EnablePanicRecovery bool
    OnPanic             func(topic string, subID uint64, recovered any)
}
```

**删除**：
- `EnableScheduler` — 调度功能由 `pubsub/scheduler` 子包包装器提供
- `EnableTTL` — TTL 由 `SubOptions.Filter` 或 `pubsub/dedup` 实现
- `EnableRingBuffer` — 统一在 `DropOldest` 策略下默认使用环形缓冲
- `EnableRequestReply` — 由 `pubsub/requestreply` 子包包装器提供
- `EnableHistory` / `HistoryConfig` — 由 `pubsub/history` 子包提供
- `Hooks` 字段（见下方）

#### Hooks → 独立 Observer 接口

```go
// BrokerObserver 用于外部监控，通过 WithObserver() 选项注入。
type BrokerObserver interface {
    OnPublish(topic string, msg *Message)
    OnSubscribe(topic string, subID uint64)
    OnUnsubscribe(topic string, subID uint64)
    OnDeliver(topic string, subID uint64, msg *Message)
    OnDrop(topic string, subID uint64, msg *Message, policy BackpressurePolicy)
}

// WithObserver 注册一个观察者（可注册多个）。
func WithObserver(o BrokerObserver) Option { ... }
```

这样测试和监控工具实现接口，核心代码不感知具体观察者。

---

### Phase 4 — Backpressure 命名厘清（`backpressure.go`）

#### 目标

两个概念赋予无歧义的名称，`BackpressureController` 移至 `pubsub/pressure` 子包并依赖接口。

#### 核心包保留

`BackpressurePolicy`（subscriber 级，原有四个值）保持不变，仅移除 `policyName()` 函数，统一用 `.String()` 方法。

#### 子包 `pubsub/pressure`

```go
// Publisher 是 pressure 包依赖的最小接口，解除对 InProcBroker 的硬依赖。
type Publisher interface {
    Publish(topic string, msg pubsub.Message) error
}

// PropagationPolicy 定义系统级背压传播行为（原 BackpressurePropagationPolicy）。
type PropagationPolicy int

const (
    PropagationNone     PropagationPolicy = iota
    PropagationWarn
    PropagationThrottle
    PropagationBlock
)

type Controller struct { ... }

func New(pub Publisher, cfg Config) *Controller { ... }
```

**删除**：
- `BackpressurePropagationPolicy`（改名 `PropagationPolicy`，移至子包）
- `ErrBackpressureActive`、`ErrBackpressureClosed`（移至子包）
- `PublishWithBackpressure()`（Controller 不再代理 Publish，职责归还调用方）

---

### Phase 5 — Metrics 合并（`metrics.go`）

#### 目标

`metrics.go` + `metrics_enhanced.go` 合并为单一内部实现；`prometheus.go` 移至 `pubsub/prometheus` 子包。

#### 合并后 `metrics.go`

保留原子计数器（publish、deliver、drop、subscriber gauge）+ 延迟直方图（原 `metrics_enhanced.go`），对外暴露：

```go
type Snapshot struct {
    Topics map[string]TopicSnapshot
}

type TopicSnapshot struct {
    Published  uint64
    Delivered  uint64
    Dropped    map[string]uint64  // key: policy.String()
    Subscribers int
    PublishP50  time.Duration
    PublishP99  time.Duration
    DeliverP50  time.Duration
    DeliverP99  time.Duration
}
```

**删除**：
- `metrics_enhanced.go`（内容合入 `metrics.go`）
- `policyName()` 函数（改用 `BackpressurePolicy.String()`）
- `EnhancedMetrics` 类型（合并后无需独立类型）

#### `pubsub/prometheus` 子包

```go
// Collector 将 pubsub.Snapshot 桥接到 prometheus.Registerer。
type Collector struct { ... }

func New(broker interface{ Snapshot() pubsub.Snapshot }, reg prometheus.Registerer) *Collector
func (c *Collector) Describe(ch chan<- *prometheus.Desc)
func (c *Collector) Collect(ch chan<- prometheus.Metric)
```

---

### Phase 6 — 企业功能子包迁移

以下四个功能完整迁入独立子包，核心包零依赖：

#### `pubsub/distributed`

```go
// DistributedBroker 包装 pubsub.Broker，通过 HTTP 实现多节点广播。
type DistributedBroker struct { ... }

func New(local pubsub.Broker, cfg ClusterConfig) (*DistributedBroker, error)
```

- `ClusterConfig`、`DefaultClusterConfig` 随之迁入
- 所有集群错误变量随之迁入
- `contract.WriteError` 调用保留（该包已引用 contract）

#### `pubsub/multitenant`

```go
// TenantBroker 为每个租户提供隔离的 pubsub 实例和配额管理。
type TenantBroker struct { ... }

func New(factory func() pubsub.Broker, cfg Config) *TenantBroker
func (tb *TenantBroker) For(tenantID string) (pubsub.Broker, error)
func (tb *TenantBroker) Register(id string, quota TenantQuota) error
```

#### `pubsub/consumergroup`

```go
// Manager 管理多消费者对 topic 的协调分配。
type Manager struct { ... }

func New(broker pubsub.Broker, cfg Config) *Manager
```

#### `pubsub/audit`

```go
// Logger 记录 pubsub 操作的不可变审计日志（SHA-256 链式）。
type Logger struct { ... }

func New(cfg Config) (*Logger, error)
func (l *Logger) Wrap(broker pubsub.Broker) pubsub.Broker
```

通过包装器模式（Decorator）在不修改核心的前提下注入审计。

---

### Phase 7 — 能力子包接口约定

所有能力子包遵循统一模式：

```
pubsub/xxx/
  xxx.go        # 公开类型 + 构造函数 New(broker pubsub.Broker, ...) *Xxx
  xxx_test.go   # 表格驱动测试，使用 pubsub.New() 真实实例
```

构造函数**接受 `pubsub.Broker` 接口**，不接受具体类型，保证可替换性：

```go
// 示例：history 子包
func New(broker pubsub.Broker, cfg Config) *Store

// 示例：scheduler 子包
func New(broker pubsub.Broker) *Scheduler
func (s *Scheduler) PublishDelayed(topic string, msg pubsub.Message, delay time.Duration) error
func (s *Scheduler) PublishAt(topic string, msg pubsub.Message, at time.Time) error
```

---

### Phase 8 — `HistoryPubSub` 接口迁移

`types.go` 中 `HistoryPubSub` 接口（7 个方法）迁至 `pubsub/history` 子包，作为 `Store` 类型的方法集合，不再作为 `Broker` 的扩展接口。

```go
// pubsub/history
type Store struct { ... }

func New(broker pubsub.Broker, cfg Config) *Store

func (s *Store) Get(topic string) ([]pubsub.Message, error)
func (s *Store) Since(topic string, seq uint64) ([]pubsub.Message, error)
func (s *Store) Recent(topic string, n int) ([]pubsub.Message, error)
func (s *Store) ByTTL(topic string, ttl time.Duration) ([]pubsub.Message, error)
func (s *Store) Clear(topic string) error
func (s *Store) Stats() (map[string]Stats, error)
func (s *Store) Sequence(topic string) (uint64, error)
```

`HistoryStats` 类型同步迁入 `pubsub/history`。

---

### Phase 9 — RequestReply 子包迁移

`types.go` 中 `RequestReplyPubSub` + `request_reply.go` 的实现迁至 `pubsub/requestreply`：

```go
// pubsub/requestreply
type Client struct { ... }

func New(broker pubsub.Broker) *Client
func (c *Client) Request(ctx context.Context, topic string, msg pubsub.Message) (pubsub.Message, error)
func (c *Client) RequestWithTimeout(topic string, msg pubsub.Message, timeout time.Duration) (pubsub.Message, error)
func (c *Client) Reply(reqMsg pubsub.Message, respMsg pubsub.Message) error
func IsRequest(msg pubsub.Message) bool
```

---

### Phase 10 — Health 检查解耦

`health.go` 中 `Health()`、`IsHealthy()`、`ReadinessCheck()`、`LivenessCheck()`、`DiagnosticInfo()` 均为特定部署场景的运维方法，从 `InProcBroker` 方法中移除，改为独立函数或通过 `Snapshot()` 方法提供数据给调用方自行决策：

```go
// Snapshot 替代 DiagnosticInfo，返回结构化数据，不做判断。
func (b *InProcBroker) Snapshot() Snapshot

// 调用方自行实现 HTTP handler：
func healthHandler(b *pubsub.InProcBroker) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        s := b.Snapshot()
        if s.Closed {
            contract.WriteError(w, contract.ServiceUnavailable(...))
            return
        }
        contract.WriteJSON(w, http.StatusOK, s)
    }
}
```

这符合 Style Guide §6（Handler 职责）和 §2（Package Roles）。

---

## 五、文件变动一览

### 核心包 `pubsub/` 保留文件

| 文件 | 变动 |
|------|------|
| `broker.go` | **新增**：`Broker`、`PatternSubscriber`、`MQTTSubscriber`、`BatchPublisher`、`Drainable` 接口 |
| `types.go` | **精简**：删除 6 个接口层级，保留 `Message`、`SubOptions`、`Subscription`、`BackpressurePolicy`、`SubscriptionStats` |
| `errors.go` | **重写**：单一 `Error` 类型，删除 17 个哨兵变量和双向映射逻辑 |
| `options.go` | **精简**：删除 `Enable*` 字段、`Hooks` 字段，`BrokerObserver` 接口新增 |
| `pubsub.go` | **重命名结构体**：`InProcPubSub` → `InProcBroker`；删除被迁出的方法 |
| `metrics.go` | **合并**：吸收 `metrics_enhanced.go` 内容，删除 `policyName()` |
| `backpressure.go` | **精简**：只保留 `BackpressurePolicy`（subscriber 级），删除 `Controller`、`BackpressurePropagationPolicy` |
| `shard.go` | 无变动（unexported） |
| `ringbuffer.go` | 无变动（unexported） |
| `workerpool.go` | 无变动（unexported） |
| `message_clone.go` | 无变动（unexported） |
| `mqtt_pattern.go` | 无变动（unexported，供 `InProcBroker` 内部使用） |

### 删除文件（迁出或废除）

| 文件 | 去向 |
|------|------|
| `metrics_enhanced.go` | 合入 `metrics.go` |
| `prometheus.go` | → `pubsub/prometheus/` |
| `health.go` | 删除，数据通过 `Snapshot()` 提供 |
| `ack.go` | → `pubsub/ack/` |
| `dlq.go` | → `pubsub/dlq/` |
| `ordering.go` | → `pubsub/ordering/` |
| `dedup.go` | → `pubsub/dedup/` |
| `ratelimit.go` | → `pubsub/ratelimit/` |
| `scheduler.go` | → `pubsub/scheduler/` |
| `replay.go` | → `pubsub/replay/` |
| `priority.go` | → `pubsub/priority/` |
| `request_reply.go` | → `pubsub/requestreply/` |
| `history.go` | → `pubsub/history/` |
| `distributed.go` | → `pubsub/distributed/` |
| `multitenant.go` | → `pubsub/multitenant/` |
| `consumergroup.go` | → `pubsub/consumergroup/` |
| `audit.go` | → `pubsub/audit/` |
| `persistence.go` | → `pubsub/persistence/` |

---

## 六、质量门限

每个阶段完成后必须通过：

```bash
go test -timeout 20s ./pubsub/...
go vet ./pubsub/...
gofmt -w ./pubsub/
go test -race ./pubsub/...
```

---

## 七、执行顺序建议

重构应按依赖关系由内向外进行，避免循环依赖：

```
Phase 1 (接口精简)
  → Phase 2 (错误模型)
    → Phase 3 (Config 清理)
      → Phase 4 (Backpressure 命名)
        → Phase 5 (Metrics 合并)
          → Phase 6-10 (子包迁移，可并行)
```

子包迁移（Phase 6-10）各自独立，可并行推进，但均依赖 Phase 1-5 的接口稳定。

---

## 八、关键设计判断说明

### 为何删除 `ContextPubSub`？

Context 参数的存在是为了取消和超时，这是**调用方关注点**。在接口层面复制一套"有 ctx 的版本"违反 DRY，且迫使所有实现者提供双倍方法。
正确做法：Broker 的 `Subscribe` 返回 `Subscription`，其 `Done()` channel 可与 `ctx.Done()` 联用；`Drain` 本身接受 `ctx`。

### 为何删除功能标志？

`EnableScheduler = true` 让核心代码在 `PublishDelayed` 被调用时返回 `ErrSchedulerDisabled`——这是隐式的运行时状态门控。
正确做法：`scheduler.New(broker)` 返回一个 `*Scheduler`，不存在"未启用"状态，构造即启用。

### 为何 `BackpressureController` 要依赖接口？

原代码 `NewBackpressureController(ps *InProcPubSub, ...)` 将控制器绑定到具体实现，导致：
- 无法对 `DistributedBroker` 使用背压控制器
- 无法在测试中替换 broker

改为依赖 `Publisher` 接口后，控制器成为通用包装器。

### 为何合并 `metrics_enhanced.go`？

该文件是 `metrics.go` 的延伸，没有独立存在的理由。两个文件描述同一主题（pubsub 内部指标），分开维护增加了认知负担且可能出现不一致。
