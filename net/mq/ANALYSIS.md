# net/mq 包分析报告

**生成时间**: 2026-01-30
**分析范围**: /home/user/plumego/net/mq/

---

## 📊 包概述

net/mq 包是一个**进程内消息代理（In-Process Message Broker）**，基于 pubsub 包构建，提供消息队列的高级功能。

**当前状态**: Experimental（实验性功能，不建议生产使用）

---

## ✅ 已完成功能

### 1. 核心发布/订阅功能
- ✅ 基本的 Publish/Subscribe 操作
- ✅ 批量发布 (PublishBatch)
- ✅ 批量订阅 (SubscribeBatch)
- ✅ Context 取消支持
- ✅ Topic 和 Message 验证

### 2. 优先级队列（Priority Queue）
- ✅ 完整实现的优先级队列 (priorityQueue)
- ✅ 基于堆的优先级排序
- ✅ 序列号保证相同优先级的 FIFO 顺序
- ✅ 优先级调度器 (priorityDispatcher)
- ✅ Context 取消支持
- ✅ 测试覆盖完整

**实现质量**: 优秀，代码完整且测试充分

### 3. 可观测性
- ✅ 指标收集 (MetricsCollector 集成)
- ✅ Panic 恢复和处理
- ✅ 错误追踪
- ✅ 健康检查 (HealthCheck)
- ✅ 内存使用监控 (GetMemoryUsage)
- ✅ 指标快照 (Snapshot)

### 4. 配置管理
- ✅ 灵活的配置系统 (Config)
- ✅ 配置验证 (Validate)
- ✅ 动态配置更新 (UpdateConfig)
- ✅ 功能选项模式 (Functional Options)

### 5. 测试覆盖
- ✅ 基础功能测试
- ✅ 并发测试
- ✅ 边界条件测试
- ✅ 错误处理测试
- ✅ 优先级队列专项测试

---

## ⚠️ 未完成功能（需要实现）

### 1. TTL（消息过期）支持
**当前状态**: 仅有类型定义，无实现逻辑

**问题**:
- mq.go:26-30 定义了 `TTLMessage` 类型
- mq.go:615-617 有 TODO 注释但跳过了 TTL 检查
- mq.go:649-650 批量发布中也跳过了 TTL 验证

**需要实现**:
```go
// 1. 在发布时检查 TTL
func (b *InProcBroker) validateTTL(msg Message) error {
    if ttlMsg, ok := msg.(TTLMessage); ok {
        if !ttlMsg.ExpiresAt.IsZero() && time.Now().After(ttlMsg.ExpiresAt) {
            return ErrMessageExpired // 新的错误类型
        }
    }
    return nil
}

// 2. 后台清理过期消息
func (b *InProcBroker) startTTLCleanup() {
    // 定期清理过期消息
}
```

**优先级**: 中

---

### 2. ACK（消息确认）机制
**当前状态**: 框架存在，核心逻辑缺失

**问题**:
- mq.go:111-121 定义了 `AckPolicy` 枚举
- mq.go:248-254 定义了 `AckMessage` 类型
- mq.go:841-843 TODO: 实现确认逻辑
- mq.go:906-908 TODO: 实现确认追踪
- mq.go:932-934 TODO: 实现负面确认逻辑

**需要实现**:
```go
// 1. ACK 追踪器
type ackTracker struct {
    mu       sync.Mutex
    pending  map[string]*ackEntry  // messageID -> ackEntry
    timers   map[string]*time.Timer
}

type ackEntry struct {
    messageID  string
    topic      string
    timestamp  time.Time
    timeout    time.Duration
    retries    int
}

// 2. ACK 超时处理
func (b *InProcBroker) handleAckTimeout(entry *ackEntry) {
    // 重新投递或发送到死信队列
}

// 3. 消息重投递
func (b *InProcBroker) redeliverMessage(entry *ackEntry) error {
    // 重新发布消息
}
```

**优先级**: 高（这是消息队列的核心功能）

---

### 3. 集群模式（Cluster Mode）
**当前状态**: 配置和框架存在，无分布式逻辑

**问题**:
- mq.go:301-315 定义了集群配置
- mq.go:418-428 定义了 `ClusterStatus` 类型
- mq.go:994-1000 TODO: 实现集群复制逻辑

**需要实现**:
```go
// 1. 节点发现和健康检查
type clusterManager struct {
    nodes     map[string]*clusterNode
    replicator *messageReplicator
    syncer    *stateSyncer
}

// 2. 消息复制
func (cm *clusterManager) replicateMessage(msg Message, factor int) error {
    // 复制消息到指定数量的节点
}

// 3. 状态同步
func (cm *clusterManager) syncClusterState() error {
    // 定期同步集群状态
}

// 4. 故障转移
func (cm *clusterManager) handleNodeFailure(nodeID string) error {
    // 处理节点故障
}
```

**优先级**: 低（分布式特性，复杂度高）

---

### 4. 持久化存储（Persistence）
**当前状态**: 配置存在，无实现

**问题**:
- mq.go:316-320 定义了持久化配置
- 没有任何持久化逻辑

**需要实现**:
```go
// 1. WAL (Write-Ahead Log) 支持
type persistenceLayer struct {
    wal     *store.WAL  // 可以复用 store/kv 的 WAL
    dataDir string
}

// 2. 消息持久化
func (pl *persistenceLayer) persistMessage(msg Message) error {
    // 写入 WAL
}

// 3. 崩溃恢复
func (pl *persistenceLayer) recover() ([]Message, error) {
    // 从 WAL 恢复消息
}
```

**优先级**: 中（可选特性，但对可靠性很重要）

---

### 5. 死信队列（Dead Letter Queue）
**当前状态**: 配置和框架存在，核心逻辑缺失

**问题**:
- mq.go:322-326 定义了死信队列配置
- mq.go:430-437 定义了 `DeadLetterStats` 类型
- mq.go:1195-1200 TODO: 实现死信队列逻辑
- mq.go:1213-1217 TODO: 实现死信统计收集

**需要实现**:
```go
// 1. 死信队列管理器
type deadLetterQueue struct {
    mu        sync.Mutex
    messages  []deadLetterMessage
    stats     DeadLetterStats
}

type deadLetterMessage struct {
    original      Message
    originalTopic string
    reason        string
    timestamp     time.Time
    retries       int
}

// 2. 发送到死信队列
func (dlq *deadLetterQueue) add(msg Message, topic, reason string) {
    // 添加到死信队列
    // 更新统计信息
}

// 3. 死信队列重试策略
func (dlq *deadLetterQueue) retryMessage(msgID string) error {
    // 重试死信消息
}
```

**优先级**: 中高（与 ACK 机制配合使用）

---

### 6. 事务支持（Transactions）
**当前状态**: 配置和框架存在，无事务逻辑

**问题**:
- mq.go:328-332 定义了事务配置
- mq.go:1092-1098 TODO: 实现事务逻辑
- mq.go:1124-1129 TODO: 实现事务提交逻辑
- mq.go:1154-1159 TODO: 实现事务回滚逻辑

**需要实现**:
```go
// 1. 事务管理器
type transactionManager struct {
    mu           sync.Mutex
    transactions map[string]*transaction
}

type transaction struct {
    id         string
    messages   []Message
    startTime  time.Time
    timeout    time.Duration
    state      txState  // pending/committed/rolledback
}

type txState int

const (
    txPending txState = iota
    txCommitted
    txRolledback
    txTimedOut
)

// 2. 事务操作
func (tm *transactionManager) begin(txID string, timeout time.Duration) error
func (tm *transactionManager) commit(txID string) error
func (tm *transactionManager) rollback(txID string) error

// 3. 事务超时处理
func (tm *transactionManager) handleTimeout(txID string) {
    // 自动回滚超时事务
}
```

**优先级**: 低（高级特性，使用场景有限）

---

### 7. MQTT 协议支持
**当前状态**: 配置存在，无实现

**问题**:
- mq.go:334-337 定义了 MQTT 配置
- mq.go:1225-1242 TODO: 实现 MQTT 服务器

**需要实现**:
- MQTT 协议解析器
- MQTT broker 实现
- 桥接到内部 pubsub

**优先级**: 很低（需要外部依赖或复杂的协议实现）

---

### 8. AMQP 协议支持
**当前状态**: 配置存在，无实现

**问题**:
- mq.go:340-344 定义了 AMQP 配置
- mq.go:1244-1261 TODO: 实现 AMQP 服务器

**需要实现**:
- AMQP 协议解析器
- AMQP broker 实现
- 桥接到内部 pubsub

**优先级**: 很低（需要外部依赖或复杂的协议实现）

---

### 9. Trie 模式匹配
**当前状态**: 仅有配置标志

**问题**:
- mq.go:298-299 定义了 `EnableTriePattern` 配置
- 测试中有验证（mq_test.go:481-496）
- 实际实现应该在 pubsub 层

**需要实现**:
- 主题通配符匹配（如 `user.*.created`）
- Trie 树结构用于高效匹配

**优先级**: 低（依赖 pubsub 层实现）

---

## 🐛 发现的 Bug 和问题

### 1. 内存限制检查未实际使用
**位置**: mq.go:938-955

**问题**:
- `checkMemoryLimit()` 方法已实现
- 但从未在 Publish 或其他操作中调用
- 测试 mq_test.go:466-479 显示发布总是成功

**影响**: 内存限制配置无效

**修复建议**:
```go
func (b *InProcBroker) Publish(ctx context.Context, topic string, msg Message) error {
    return b.executeWithObservability(ctx, OpPublish, topic, func() error {
        // 添加内存检查
        if err := b.checkMemoryLimit(); err != nil {
            return err
        }
        // ... 现有逻辑
    })
}
```

---

### 2. Topic 长度验证不一致
**位置**: mq.go:479

**问题**:
- 最大长度硬编码为 1024
- 没有配置选项可以调整

**建议**: 添加到 Config 中
```go
type Config struct {
    // ...
    MaxTopicLength int  // 默认 1024
}
```

---

### 3. 配置验证不完整
**位置**: mq.go:377-391

**问题**:
- `Config.Validate()` 只检查了基础字段
- 未验证集群配置（如 `ClusterNodeID` 不能为空当 `EnableCluster=true`）
- 未验证持久化路径（如 `PersistencePath` 必须存在当 `EnablePersistence=true`）

**建议**: 增强验证逻辑
```go
func (c Config) Validate() error {
    // 现有验证...

    if c.EnableCluster && c.ClusterNodeID == "" {
        return fmt.Errorf("%w: ClusterNodeID required when cluster is enabled", ErrInvalidConfig)
    }

    if c.EnablePersistence && c.PersistencePath == "" {
        return fmt.Errorf("%w: PersistencePath required when persistence is enabled", ErrInvalidConfig)
    }

    // ...更多验证
}
```

---

### 4. 资源泄漏风险
**位置**: mq.go:511-537

**问题**:
- `ensurePriorityDispatcher()` 创建 goroutine (mq.go:179)
- 如果创建多个 topic 的 dispatcher，goroutine 会累积
- Close 时有清理逻辑（mq.go:539-558），但未验证是否完全

**建议**:
- 添加 dispatcher 数量限制
- 增强测试验证 goroutine 清理

---

### 5. 错误处理不一致
**位置**: 多处

**问题**:
- 有些方法返回 `fmt.Errorf("%w", ErrNotInitialized)`（正确）
- 有些方法返回 `fmt.Errorf("%w: xxx", ErrInvalidConfig, ...)`（正确）
- 但错误信息格式不统一

**建议**: 统一错误包装风格

---

## 🔧 代码质量问题

### 1. 类型断言缺失错误处理
**位置**: mq.go:1294-1303

**问题**:
```go
if snapper, ok := b.ps.(interface {
    ListTopics() []string
    GetSubscriberCount(topic string) int
}); ok {
    // 使用 snapper
}
```
- 依赖接口检查，但不保证实现正确性
- 应该添加文档说明 pubsub 实现需要提供这些方法

---

### 2. 重复代码
**位置**: 多个 Publish*/Subscribe* 方法

**问题**:
- 每个方法都重复相同的验证逻辑：
  - Context 验证
  - Broker 初始化检查
  - Topic 验证
  - Message 验证

**建议**: 提取公共验证函数
```go
func (b *InProcBroker) validateOperation(ctx context.Context, topic string, msg *Message) error {
    if ctx != nil && ctx.Err() != nil {
        return ctx.Err()
    }
    if b == nil || b.ps == nil {
        return ErrNotInitialized
    }
    if err := validateTopic(topic); err != nil {
        return err
    }
    if msg != nil {
        if err := validateMessage(*msg); err != nil {
            return err
        }
    }
    return nil
}
```

---

### 3. Magic Numbers
**位置**: 多处

**问题**:
- mq.go:479: `1024` (topic 长度)
- mq.go:353: `16` (默认缓冲区大小)

**建议**: 定义为常量
```go
const (
    DefaultBufferSize = 16
    MaxTopicLength    = 1024
)
```

---

### 4. 测试覆盖不足

**已覆盖**:
- 基本发布/订阅 ✅
- 并发操作 ✅
- 配置管理 ✅
- 优先级队列 ✅

**缺失**:
- TTL 消息过期测试
- ACK 超时和重试测试
- 内存限制强制执行测试
- 资源泄漏测试（goroutine/内存）
- Benchmark 测试（性能基线）

---

## 📈 性能优化建议

### 1. Priority Queue 优化
**当前实现**: 每个 topic 一个 dispatcher goroutine

**建议**:
- 考虑使用 goroutine 池处理优先级消息
- 避免为每个 topic 创建独立的 goroutine

```go
type priorityWorkerPool struct {
    workers   int
    taskQueue chan *priorityTask
}

type priorityTask struct {
    topic string
    msg   PriorityMessage
}
```

---

### 2. 内存分配优化
**位置**: mq.go:643-652

**问题**:
```go
validMsgs := make([]Message, 0, len(msgs))
for _, msg := range msgs {
    // 验证后添加
    validMsgs = append(validMsgs, msg)
}
```

**建议**: 就地验证，避免创建新切片
```go
for i, msg := range msgs {
    if err := validateMessage(msg); err != nil {
        return fmt.Errorf("message %d: %w", i, err)
    }
}
// 直接使用 msgs
```

---

### 3. 锁粒度优化
**位置**: mq.go:519-521

**问题**:
```go
b.priorityMu.Lock()
defer b.priorityMu.Unlock()
```
- 整个 map 操作都持有全局锁
- 可以改用 sync.Map 或分片锁

---

### 4. 指标收集性能
**位置**: mq.go:1405-1421

**问题**:
- 每次操作都收集指标
- `ObserveMQ` 调用可能成为瓶颈

**建议**:
- 批量收集指标
- 使用无锁数据结构（如 atomic）
- 可选的采样机制

---

## 🏗️ 架构改进建议

### 1. 分层架构
**当前**: InProcBroker 直接包装 pubsub

**建议**: 引入中间层
```
InProcBroker
    ├── PriorityManager (优先级队列)
    ├── AckManager (确认管理)
    ├── TTLManager (过期管理)
    ├── DeadLetterManager (死信队列)
    └── PubSub (底层)
```

### 2. 插件化设计
**建议**: 将高级特性设计为可插拔的组件

```go
type Feature interface {
    Name() string
    OnPublish(ctx context.Context, topic string, msg Message) error
    OnSubscribe(ctx context.Context, topic string) error
    Close() error
}

type InProcBroker struct {
    ps       pubsub.PubSub
    features map[string]Feature
}
```

### 3. 接口隔离
**当前**: InProcBroker 实现了所有方法（包括未实现的）

**建议**: 使用接口组合
```go
type BasicBroker interface {
    Publish(ctx context.Context, topic string, msg Message) error
    Subscribe(ctx context.Context, topic string, opts SubOptions) (Subscription, error)
    Close() error
}

type PriorityBroker interface {
    PublishPriority(ctx context.Context, topic string, msg PriorityMessage) error
}

type AckBroker interface {
    PublishWithAck(ctx context.Context, topic string, msg AckMessage) error
    Ack(ctx context.Context, topic string, messageID string) error
    Nack(ctx context.Context, topic string, messageID string) error
}

// InProcBroker 可选实现这些接口
```

---

## 📝 文档改进建议

### 1. 缺失的文档
- 包级别的使用示例
- 各个特性的启用指南
- 性能特性和限制说明
- 与 pubsub 的区别说明

### 2. 建议添加示例
```go
// Example: Basic usage
func ExampleInProcBroker_basic() {
    broker := NewInProcBroker(pubsub.New())
    defer broker.Close()

    // Subscribe
    sub, _ := broker.Subscribe(context.Background(), "events", SubOptions{BufferSize: 10})
    defer sub.Cancel()

    // Publish
    broker.Publish(context.Background(), "events", Message{ID: "1", Data: "hello"})

    // Receive
    msg := <-sub.C()
    fmt.Println(msg.Data)
}

// Example: Priority queue
func ExampleInProcBroker_priority() {
    cfg := DefaultConfig()
    cfg.EnablePriorityQueue = true
    broker := NewInProcBroker(pubsub.New(), WithConfig(cfg))
    defer broker.Close()

    // Publish with priority
    broker.PublishPriority(context.Background(), "tasks", PriorityMessage{
        Message:  Message{ID: "1", Data: "urgent task"},
        Priority: PriorityHigh,
    })
}
```

---

## 🎯 优先级总结

### 高优先级（应该立即实现）
1. **ACK 机制** - 消息队列的核心功能
2. **内存限制强制执行** - 修复已有的 bug
3. **配置验证增强** - 防止错误配置
4. **重复代码重构** - 提升代码质量
5. **测试覆盖增强** - 确保稳定性

### 中优先级（可以逐步实现）
1. **TTL 支持** - 常用的功能
2. **死信队列** - 与 ACK 配合使用
3. **持久化存储** - 提升可靠性
4. **性能优化** - 提升吞吐量
5. **文档和示例** - 提升可用性

### 低优先级（可选功能）
1. **集群模式** - 复杂度高，使用场景有限
2. **事务支持** - 使用场景有限
3. **Trie 模式匹配** - 依赖 pubsub 层
4. **MQTT/AMQP 协议** - 需要外部依赖

---

## 🔍 与 plumego 其他模块的关系

### 1. 依赖关系
- **pubsub**: 直接依赖，是底层实现
- **metrics**: 用于可观测性
- **store/kv**: 可以用于持久化（未使用）

### 2. 集成建议
- 可以在 `core.App` 中提供 MQ 组件支持
- 类似 WebSocket hub 的集成方式

```go
// 在 core/ 中添加
func WithMessageQueue(cfg mq.Config) Option {
    return func(app *App) {
        broker := mq.NewInProcBroker(pubsub.New(), mq.WithConfig(cfg))
        app.RegisterComponent(broker)
    }
}
```

---

## 📊 代码统计

- **总行数**: ~1422 行（mq.go）
- **测试行数**: ~799 行（两个测试文件）
- **测试覆盖率**: 约 60-70%（估计，基于已实现功能）
- **TODO 数量**: 12 个
- **已实现特性**: ~40%
- **配置选项**: 25 个

---

## 🚀 实施路线图

### Phase 1: 稳定核心功能（2-3 周）
- [ ] 修复内存限制强制执行 bug
- [ ] 增强配置验证
- [ ] 重构重复代码
- [ ] 添加更多单元测试
- [ ] 添加 benchmark 测试

### Phase 2: 实现 ACK 机制（2-3 周）
- [ ] 实现 ACK 追踪器
- [ ] 实现超时处理
- [ ] 实现消息重投递
- [ ] 集成死信队列
- [ ] 完整测试覆盖

### Phase 3: 完善高级特性（3-4 周）
- [ ] 实现 TTL 支持
- [ ] 实现死信队列
- [ ] 实现持久化存储
- [ ] 性能优化
- [ ] 文档和示例

### Phase 4: 可选功能（时间允许）
- [ ] 集群模式（需要详细设计）
- [ ] 事务支持
- [ ] 协议支持（MQTT/AMQP）

---

## ✅ 建议

1. **移除实验性标记之前**，必须完成：
   - ACK 机制实现
   - TTL 支持
   - 死信队列
   - 完整的测试覆盖（>80%）
   - 文档和示例

2. **重新评估特性范围**：
   - 当前功能列表过于庞大
   - 建议聚焦核心消息队列功能
   - 将集群、事务、协议支持移到单独的包或扩展

3. **改进 API 设计**：
   - 使用接口隔离原则
   - 只导出已实现的功能
   - 未实现的功能不要导出

4. **性能基准**：
   - 添加 benchmark 测试
   - 建立性能基线
   - 监控性能退化

---

## 📚 参考资料

如果要完善这些功能，可以参考：
- **RabbitMQ**: ACK 机制、死信队列设计
- **Kafka**: 消息持久化、分区策略
- **NATS**: 轻量级 MQ 架构
- **Redis Streams**: 消息 ID、消费者组
- **plumego/store/kv**: 已有的 WAL 实现可用于持久化

---

**报告结束**
