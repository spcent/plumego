# 数据库分库分表与读写分离设计方案

## 1. 概述

本设计方案旨在为 plumego 的 `store/db` 包添加企业级数据库中间件能力,包括:

- **读写分离**: 主从架构支持,读写请求自动路由
- **分库分表**: 水平分片支持,支持多种分片策略
- **对业务透明**: 通过统一接口屏蔽底层复杂性
- **高可用性**: 故障转移、健康检查、自动恢复
- **可观测性**: 完整的监控指标和追踪支持

## 2. 设计原则

### 2.1 核心原则

1. **向后兼容**: 现有代码无需修改即可工作
2. **标准库优先**: 继续基于 `database/sql`,不引入第三方 ORM
3. **显式配置**: 分片规则和路由策略必须显式声明
4. **渐进式采用**: 可以只启用部分功能(如只用读写分离,不用分表)
5. **性能优先**: 路由决策必须在纳秒级完成
6. **测试友好**: 支持 mock 和单元测试

### 2.2 非目标

- ❌ 不提供自动化的数据迁移工具
- ❌ 不实现分布式事务(2PC/3PC)
- ❌ 不做 SQL 语法树完整解析(仅做关键字提取)
- ❌ 不处理跨分片 JOIN(由业务层处理)

## 3. 架构设计

### 3.1 整体架构

```
┌─────────────────────────────────────────────────────────┐
│                    Business Layer                        │
│              (使用 db.DB 接口,无感知)                      │
└────────────────────┬────────────────────────────────────┘
                     │
                     ▼
┌─────────────────────────────────────────────────────────┐
│                  Routing Layer                           │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐  │
│  │  SQL Parser  │→│ Shard Router │→│ RW Splitter  │  │
│  └──────────────┘  └──────────────┘  └──────────────┘  │
└────────────────────┬────────────────────────────────────┘
                     │
          ┌──────────┼──────────┐
          ▼          ▼          ▼
    ┌─────────┐ ┌─────────┐ ┌─────────┐
    │ Shard 0 │ │ Shard 1 │ │ Shard N │
    │ ┌─────┐ │ │ ┌─────┐ │ │ ┌─────┐ │
    │ │Write│ │ │ │Write│ │ │ │Write│ │
    │ └──┬──┘ │ │ └──┬──┘ │ │ └──┬──┘ │
    │ ┌──┴──┐ │ │ ┌──┴──┐ │ │ ┌──┴──┐ │
    │ │Read │ │ │ │Read │ │ │ │Read │ │
    │ │Read │ │ │ │Read │ │ │ │Read │ │
    │ └─────┘ │ │ └─────┘ │ │ └─────┘ │
    └─────────┘ └─────────┘ └─────────┘
```

### 3.2 模块划分

```
store/db/
├── sql.go                    # 现有基础包 (不变)
├── sql_test.go               # 现有测试 (不变)
├── cluster.go                # 集群管理 (NEW)
├── cluster_test.go
├── router.go                 # SQL 路由器 (NEW)
├── router_test.go
├── sharding/
│   ├── strategy.go           # 分片策略接口
│   ├── hash.go               # 哈希分片
│   ├── range.go              # 范围分片
│   ├── mod.go                # 取模分片
│   ├── resolver.go           # 分片键解析
│   └── strategy_test.go
├── rw/
│   ├── splitter.go           # 读写分离
│   ├── loadbalancer.go       # 负载均衡
│   ├── policy.go             # 路由策略
│   └── splitter_test.go
├── pool/
│   ├── manager.go            # 连接池管理
│   ├── shard_pool.go         # 分片连接池
│   └── manager_test.go
└── SHARDING_DESIGN.md        # 本文档
```

## 4. 读写分离设计

### 4.1 核心概念

- **主库 (Primary/Master)**: 处理所有写操作和强一致性读
- **从库 (Replicas/Slaves)**: 处理只读查询,可配置多个
- **负载均衡**: 从库之间的请求分发策略
- **故障转移**: 从库不可用时的降级策略

### 4.2 接口设计

```go
// ReadWriteCluster 读写分离集群
type ReadWriteCluster struct {
    primary   DB              // 主库连接
    replicas  []DB            // 从库连接列表
    lb        LoadBalancer    // 负载均衡器
    policy    RoutingPolicy   // 路由策略
    health    *HealthChecker  // 健康检查
}

// LoadBalancer 负载均衡策略
type LoadBalancer interface {
    // Next 返回下一个可用的从库索引
    Next(replicas []DB) (int, error)
}

// RoutingPolicy 路由策略
type RoutingPolicy interface {
    // ShouldUsePrimary 判断是否应该使用主库
    ShouldUsePrimary(ctx context.Context, query string) bool
}
```

### 4.3 负载均衡策略

#### 4.3.1 轮询 (Round Robin)
```go
type RoundRobinBalancer struct {
    counter atomic.Uint64
}

func (b *RoundRobinBalancer) Next(replicas []DB) (int, error) {
    if len(replicas) == 0 {
        return -1, ErrNoReplicasAvailable
    }
    idx := int(b.counter.Add(1) % uint64(len(replicas)))
    return idx, nil
}
```

#### 4.3.2 随机 (Random)
```go
type RandomBalancer struct {
    rng *rand.Rand
}
```

#### 4.3.3 最少连接 (Least Connections)
```go
type LeastConnBalancer struct {
    // 基于 sql.DB.Stats() 选择连接数最少的从库
}
```

#### 4.3.4 加权轮询 (Weighted Round Robin)
```go
type WeightedBalancer struct {
    weights []int  // 权重配置
}
```

### 4.4 路由策略

#### 4.4.1 基于 SQL 语句类型

```go
type SQLTypePolicy struct{}

func (p *SQLTypePolicy) ShouldUsePrimary(ctx context.Context, query string) bool {
    // 简单的关键字匹配
    upper := strings.ToUpper(strings.TrimSpace(query))

    // 写操作 -> 主库
    if strings.HasPrefix(upper, "INSERT") ||
       strings.HasPrefix(upper, "UPDATE") ||
       strings.HasPrefix(upper, "DELETE") ||
       strings.HasPrefix(upper, "CREATE") ||
       strings.HasPrefix(upper, "ALTER") ||
       strings.HasPrefix(upper, "DROP") {
        return true
    }

    // SELECT FOR UPDATE -> 主库
    if strings.Contains(upper, "FOR UPDATE") {
        return true
    }

    // 读操作 -> 从库
    return false
}
```

#### 4.4.2 基于上下文提示

```go
type ctxKey int

const (
    keyUsePrimary ctxKey = iota
    keyPreferReplica
)

// ForcePrimary 强制使用主库
func ForcePrimary(ctx context.Context) context.Context {
    return context.WithValue(ctx, keyUsePrimary, true)
}

// PreferReplica 优先使用从库(即使是写操作,用于数据导出等场景)
func PreferReplica(ctx context.Context) context.Context {
    return context.WithValue(ctx, keyPreferReplica, true)
}
```

#### 4.4.3 基于事务状态

```go
type TransactionAwarePolicy struct {
    base RoutingPolicy
}

func (p *TransactionAwarePolicy) ShouldUsePrimary(ctx context.Context, query string) bool {
    // 事务中的所有操作都走主库
    if isInTransaction(ctx) {
        return true
    }
    return p.base.ShouldUsePrimary(ctx, query)
}
```

### 4.5 健康检查

```go
type HealthChecker struct {
    interval      time.Duration        // 检查间隔
    timeout       time.Duration        // 超时时间
    unhealthy     map[int]time.Time    // 不健康的从库及发现时间
    mu            sync.RWMutex
}

// IsHealthy 检查指定从库是否健康
func (hc *HealthChecker) IsHealthy(idx int) bool {
    hc.mu.RLock()
    defer hc.mu.RUnlock()
    _, exists := hc.unhealthy[idx]
    return !exists
}

// StartHealthCheck 启动后台健康检查
func (hc *HealthChecker) StartHealthCheck(ctx context.Context, replicas []DB) {
    ticker := time.NewTicker(hc.interval)
    defer ticker.Stop()

    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            hc.checkAll(replicas)
        }
    }
}
```

## 5. 分库分表设计

### 5.1 核心概念

- **分片 (Shard)**: 一个独立的数据库实例或 schema
- **分片键 (Sharding Key)**: 用于决定数据分片位置的字段(如 user_id)
- **分片策略 (Sharding Strategy)**: 分片键到分片的映射算法
- **路由规则 (Routing Rule)**: 表级别的分片配置

### 5.2 分片策略接口

```go
// ShardingStrategy 分片策略接口
type ShardingStrategy interface {
    // Shard 根据分片键计算分片索引
    Shard(key any, numShards int) (int, error)

    // ShardRange 计算分片范围(用于批量查询)
    ShardRange(start, end any, numShards int) ([]int, error)

    // Name 返回策略名称
    Name() string
}

// ShardingRule 分片规则
type ShardingRule struct {
    // TableName 表名或表名模式 (支持通配符)
    TableName string

    // ShardKeyColumn 分片键列名
    ShardKeyColumn string

    // Strategy 分片策略
    Strategy ShardingStrategy

    // NumShards 分片数量
    NumShards int

    // ActualTableNames 实际表名列表 (可选,用于已分表场景)
    // 例如: ["user_0", "user_1", "user_2", "user_3"]
    ActualTableNames []string

    // DatabaseSharding 是否分库 (true: 分库, false: 只分表)
    DatabaseSharding bool
}
```

### 5.3 分片策略实现

#### 5.3.1 哈希分片 (Hash Sharding)

```go
type HashStrategy struct {
    hashFunc func(key any) uint64  // 哈希函数 (默认 FNV-1a)
}

func (s *HashStrategy) Shard(key any, numShards int) (int, error) {
    hash := s.hashFunc(key)
    return int(hash % uint64(numShards)), nil
}

// 使用示例:
// key="user123" -> hash(user123) % 4 = 2 -> shard_2
```

**优点**:
- 数据分布均匀
- 扩容时可使用一致性哈希减少迁移

**缺点**:
- 范围查询需要查所有分片
- 扩容需要数据迁移

**适用场景**:
- 用户数据分片 (user_id)
- 订单数据分片 (order_id)

#### 5.3.2 范围分片 (Range Sharding)

```go
type RangeStrategy struct {
    ranges []RangeDefinition  // 范围定义列表
}

type RangeDefinition struct {
    Start any   // 范围起始值
    End   any   // 范围结束值 (不含)
    Shard int   // 目标分片
}

func (s *RangeStrategy) Shard(key any, numShards int) (int, error) {
    for _, r := range s.ranges {
        if inRange(key, r.Start, r.End) {
            return r.Shard, nil
        }
    }
    return -1, ErrNoMatchingRange
}

// 使用示例:
// 配置: [0, 10000) -> shard_0
//      [10000, 20000) -> shard_1
//      [20000, ∞) -> shard_2
// key=15000 -> shard_1
```

**优点**:
- 范围查询高效
- 便于按时间/地区分片

**缺点**:
- 需要预先规划范围
- 数据可能不均匀

**适用场景**:
- 按时间分片 (created_at)
- 按地区分片 (region_id)

#### 5.3.3 取模分片 (Modulo Sharding)

```go
type ModStrategy struct {
    // 最简单的分片策略
}

func (s *ModStrategy) Shard(key any, numShards int) (int, error) {
    id, ok := toInt64(key)
    if !ok {
        return -1, ErrInvalidShardKey
    }
    return int(id % int64(numShards)), nil
}

// 使用示例:
// key=12345, numShards=4 -> 12345 % 4 = 1 -> shard_1
```

**优点**:
- 实现简单
- 性能极高

**缺点**:
- 扩容困难(需要重新取模)
- 只支持整数键

**适用场景**:
- 固定分片数的简单场景

#### 5.3.4 列表分片 (List Sharding)

```go
type ListStrategy struct {
    mapping map[any]int  // 显式映射表
}

func (s *ListStrategy) Shard(key any, numShards int) (int, error) {
    shard, ok := s.mapping[key]
    if !ok {
        return -1, ErrNoMatchingList
    }
    return shard, nil
}

// 使用示例:
// mapping: {"CN": 0, "US": 1, "EU": 2, "JP": 3}
// key="US" -> shard_1
```

**优点**:
- 灵活,适合枚举值
- 可手动调整分布

**缺点**:
- 需要维护映射表
- 不适合大量不同值

**适用场景**:
- 按国家/地区分片
- 按租户/商户分片

### 5.4 分片键解析

```go
// ShardKeyResolver 从 SQL 和参数中提取分片键
type ShardKeyResolver struct {
    rules map[string]*ShardingRule  // 表名 -> 分片规则
}

// ExtractShardKey 提取分片键值
func (r *ShardKeyResolver) ExtractShardKey(
    query string,
    args []any,
) (table string, key any, err error) {
    // 1. 解析表名
    table, err = r.parseTableName(query)
    if err != nil {
        return "", nil, err
    }

    // 2. 查找分片规则
    rule, ok := r.rules[table]
    if !ok {
        return table, nil, nil  // 无分片规则,返回 nil
    }

    // 3. 提取分片键
    key, err = r.extractKeyFromQuery(query, args, rule.ShardKeyColumn)
    return table, key, err
}

// parseTableName 从 SQL 中提取表名
func (r *ShardKeyResolver) parseTableName(query string) (string, error) {
    // 简化实现:使用正则表达式提取
    // INSERT INTO users ... -> "users"
    // SELECT * FROM orders WHERE ... -> "orders"
    // UPDATE products SET ... -> "products"

    upper := strings.ToUpper(strings.TrimSpace(query))

    // INSERT INTO table_name
    if strings.HasPrefix(upper, "INSERT INTO") {
        return extractTableAfter(upper, "INSERT INTO"), nil
    }

    // UPDATE table_name
    if strings.HasPrefix(upper, "UPDATE") {
        return extractTableAfter(upper, "UPDATE"), nil
    }

    // DELETE FROM table_name
    if strings.HasPrefix(upper, "DELETE FROM") {
        return extractTableAfter(upper, "DELETE FROM"), nil
    }

    // SELECT ... FROM table_name
    if strings.HasPrefix(upper, "SELECT") {
        fromIdx := strings.Index(upper, " FROM ")
        if fromIdx == -1 {
            return "", ErrNoTableFound
        }
        return extractTableAfter(upper[fromIdx:], "FROM"), nil
    }

    return "", ErrUnsupportedSQL
}

// extractKeyFromQuery 从 WHERE 子句或 VALUES 中提取分片键值
func (r *ShardKeyResolver) extractKeyFromQuery(
    query string,
    args []any,
    keyColumn string,
) (any, error) {
    // 场景 1: WHERE user_id = ?
    // 场景 2: WHERE user_id IN (?, ?, ?)  -> 返回第一个值或错误
    // 场景 3: INSERT INTO users (user_id, ...) VALUES (?, ...)
    //        -> 从 args 中根据列位置提取

    // 简化实现:支持最常见的场景
    upper := strings.ToUpper(query)

    // 查找 WHERE user_id = ? 模式
    pattern := fmt.Sprintf("%s =", strings.ToUpper(keyColumn))
    if idx := strings.Index(upper, pattern); idx != -1 {
        // 计算是第几个占位符
        placeholderIdx := countPlaceholdersBefore(query[:idx])
        if placeholderIdx < len(args) {
            return args[placeholderIdx], nil
        }
    }

    // INSERT 场景需要解析列顺序
    if strings.HasPrefix(upper, "INSERT INTO") {
        columnIdx := r.findColumnIndex(query, keyColumn)
        if columnIdx >= 0 && columnIdx < len(args) {
            return args[columnIdx], nil
        }
    }

    return nil, ErrShardKeyNotFound
}
```

### 5.5 SQL 改写

对于已经物理分表的场景(如 user_0, user_1, user_2),需要改写 SQL:

```go
type SQLRewriter struct {
    rules map[string]*ShardingRule
}

// Rewrite 改写 SQL 语句
func (rw *SQLRewriter) Rewrite(
    originalSQL string,
    table string,
    shardIdx int,
) (string, error) {
    rule, ok := rw.rules[table]
    if !ok || len(rule.ActualTableNames) == 0 {
        return originalSQL, nil  // 无需改写
    }

    if shardIdx >= len(rule.ActualTableNames) {
        return "", ErrInvalidShardIndex
    }

    actualTable := rule.ActualTableNames[shardIdx]

    // 简单替换表名
    // "SELECT * FROM users WHERE ..." -> "SELECT * FROM users_2 WHERE ..."
    return strings.Replace(originalSQL, table, actualTable, 1), nil
}
```

### 5.6 跨分片查询处理

对于无法确定分片的查询(如没有分片键的 WHERE 条件),需要查询所有分片并合并结果:

```go
type MultiShardQuery struct {
    shards []DB
}

// QueryAll 查询所有分片并合并结果
func (m *MultiShardQuery) QueryAll(
    ctx context.Context,
    query string,
    args ...any,
) (*sql.Rows, error) {
    // 方案 1: 顺序查询 (简单但慢)
    // 方案 2: 并发查询 + 内存合并 (快但内存占用大)
    // 方案 3: 流式合并 (推荐)

    // 这里不实现完整的合并逻辑,由业务层处理
    // 或者返回错误要求明确分片键
    return nil, ErrCrossShardQueryNotSupported
}
```

**建议**:
- 优先级 1: 要求业务层明确分片键,拒绝跨分片查询
- 优先级 2: 支持简单的并发查询 + 内存合并(限制结果集大小)
- 优先级 3: 复杂聚合查询由业务层使用 MapReduce 模式处理

## 6. 统一路由层

### 6.1 核心路由器

```go
// Router 统一路由器,整合读写分离和分片功能
type Router struct {
    // 分片配置
    shards        []*ShardCluster    // 分片集群列表
    shardingRules map[string]*ShardingRule
    resolver      *ShardKeyResolver
    rewriter      *SQLRewriter

    // 单库配置(向后兼容)
    singleDB      *ReadWriteCluster

    // 配置
    config        RouterConfig
}

// ShardCluster 单个分片的读写集群
type ShardCluster struct {
    index    int                // 分片索引
    cluster  *ReadWriteCluster  // 读写分离集群
}

// RouterConfig 路由器配置
type RouterConfig struct {
    // 是否启用分片
    EnableSharding bool

    // 是否启用读写分离
    EnableReadWrite bool

    // 跨分片查询策略
    CrossShardPolicy CrossShardPolicy

    // SQL 解析超时
    ParseTimeout time.Duration
}

// CrossShardPolicy 跨分片查询策略
type CrossShardPolicy int

const (
    CrossShardDeny   CrossShardPolicy = iota  // 拒绝跨分片查询
    CrossShardFirst                           // 只查询第一个分片
    CrossShardAll                             // 查询所有分片(并发)
)
```

### 6.2 路由决策流程

```go
func (r *Router) ExecContext(
    ctx context.Context,
    query string,
    args ...any,
) (sql.Result, error) {
    // 1. 解析分片键
    table, shardKey, err := r.resolver.ExtractShardKey(query, args)
    if err != nil {
        return nil, err
    }

    // 2. 确定目标分片
    var targetDB DB
    var finalSQL string

    if shardKey != nil && r.config.EnableSharding {
        // 有分片键 -> 路由到特定分片
        rule := r.shardingRules[table]
        shardIdx, err := rule.Strategy.Shard(shardKey, rule.NumShards)
        if err != nil {
            return nil, err
        }

        // 改写 SQL
        finalSQL, err = r.rewriter.Rewrite(query, table, shardIdx)
        if err != nil {
            return nil, err
        }

        // 获取分片的主库(写操作)
        targetDB = r.shards[shardIdx].cluster.primary
    } else {
        // 无分片或单库模式
        targetDB = r.singleDB.primary
        finalSQL = query
    }

    // 3. 执行查询
    return targetDB.ExecContext(ctx, finalSQL, args...)
}

func (r *Router) QueryContext(
    ctx context.Context,
    query string,
    args ...any,
) (*sql.Rows, error) {
    // 1. 解析分片键
    table, shardKey, err := r.resolver.ExtractShardKey(query, args)
    if err != nil {
        return nil, err
    }

    // 2. 确定目标数据库
    var targetDB DB
    var finalSQL string

    if shardKey != nil && r.config.EnableSharding {
        // 路由到特定分片
        rule := r.shardingRules[table]
        shardIdx, err := rule.Strategy.Shard(shardKey, rule.NumShards)
        if err != nil {
            return nil, err
        }

        finalSQL, _ = r.rewriter.Rewrite(query, table, shardIdx)

        // 读写分离:选择从库
        if r.config.EnableReadWrite {
            targetDB, err = r.shards[shardIdx].cluster.SelectReplica()
            if err != nil {
                // 降级到主库
                targetDB = r.shards[shardIdx].cluster.primary
            }
        } else {
            targetDB = r.shards[shardIdx].cluster.primary
        }
    } else if shardKey == nil && r.config.EnableSharding {
        // 没有分片键 -> 根据策略处理
        switch r.config.CrossShardPolicy {
        case CrossShardDeny:
            return nil, ErrCrossShardQueryNotSupported
        case CrossShardFirst:
            targetDB = r.shards[0].cluster.primary
            finalSQL = query
        case CrossShardAll:
            return r.queryAllShards(ctx, query, args...)
        }
    } else {
        // 单库模式 + 读写分离
        if r.config.EnableReadWrite {
            targetDB, _ = r.singleDB.SelectReplica()
        } else {
            targetDB = r.singleDB.primary
        }
        finalSQL = query
    }

    // 3. 执行查询
    return targetDB.QueryContext(ctx, finalSQL, args...)
}
```

### 6.3 对业务层透明的接口

```go
// ClusterDB 实现 db.DB 接口,对业务层完全透明
type ClusterDB struct {
    router *Router
}

// 实现 DB 接口
func (c *ClusterDB) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
    return c.router.ExecContext(ctx, query, args...)
}

func (c *ClusterDB) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
    return c.router.QueryContext(ctx, query, args...)
}

func (c *ClusterDB) QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
    return c.router.QueryRowContext(ctx, query, args...)
}

func (c *ClusterDB) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
    return c.router.BeginTx(ctx, opts)
}

func (c *ClusterDB) PingContext(ctx context.Context) error {
    return c.router.PingContext(ctx)
}

func (c *ClusterDB) Close() error {
    return c.router.Close()
}
```

### 6.4 使用示例

```go
// 示例 1: 单库 + 读写分离
db := cluster.New(cluster.Config{
    Primary: cluster.DataSource{
        Driver: "mysql",
        DSN:    "user:pass@tcp(primary:3306)/db",
    },
    Replicas: []cluster.DataSource{
        {Driver: "mysql", DSN: "user:pass@tcp(replica1:3306)/db"},
        {Driver: "mysql", DSN: "user:pass@tcp(replica2:3306)/db"},
    },
    LoadBalancer: cluster.RoundRobin(),
})

// 业务代码无需改变
rows, err := db.QueryContext(ctx, "SELECT * FROM users WHERE id = ?", 123)

// 示例 2: 分库分表 + 读写分离
db := cluster.New(cluster.Config{
    Shards: []cluster.ShardConfig{
        {
            Index: 0,
            Primary:  cluster.DataSource{Driver: "mysql", DSN: "..."},
            Replicas: []cluster.DataSource{...},
        },
        {
            Index: 1,
            Primary:  cluster.DataSource{Driver: "mysql", DSN: "..."},
            Replicas: []cluster.DataSource{...},
        },
    },
    ShardingRules: []cluster.ShardingRule{
        {
            TableName:      "users",
            ShardKeyColumn: "user_id",
            Strategy:       sharding.Hash(),
            NumShards:      2,
        },
    },
})

// 业务代码完全相同,自动路由到正确的分片
result, err := db.ExecContext(ctx,
    "INSERT INTO users (user_id, name) VALUES (?, ?)",
    12345, "Alice",
)
// 自动计算: hash(12345) % 2 = 1 -> shard_1.primary
```

## 7. 监控和可观测性

### 7.1 指标收集

```go
type RouterMetrics struct {
    // 请求计数
    TotalQueries     atomic.Uint64
    TotalWrites      atomic.Uint64
    TotalReads       atomic.Uint64

    // 路由统计
    PrimaryQueries   atomic.Uint64
    ReplicaQueries   atomic.Uint64
    CrossShardQueries atomic.Uint64

    // 分片统计
    ShardQueries     []atomic.Uint64  // 每个分片的查询数

    // 性能指标
    RoutingLatency   *histogram.Histogram
    QueryLatency     *histogram.Histogram

    // 错误统计
    RoutingErrors    atomic.Uint64
    ShardingErrors   atomic.Uint64
    ReplicaErrors    atomic.Uint64
}

// RecordQuery 记录一次查询
func (m *RouterMetrics) RecordQuery(
    duration time.Duration,
    shardIdx int,
    isPrimary bool,
    err error,
) {
    m.TotalQueries.Add(1)

    if isPrimary {
        m.PrimaryQueries.Add(1)
    } else {
        m.ReplicaQueries.Add(1)
    }

    if shardIdx >= 0 && shardIdx < len(m.ShardQueries) {
        m.ShardQueries[shardIdx].Add(1)
    }

    m.QueryLatency.Record(duration.Microseconds())

    if err != nil {
        m.RoutingErrors.Add(1)
    }
}
```

### 7.2 分布式追踪

```go
import "go.opentelemetry.io/otel/trace"

func (r *Router) QueryContext(
    ctx context.Context,
    query string,
    args ...any,
) (*sql.Rows, error) {
    ctx, span := r.tracer.Start(ctx, "db.Router.Query")
    defer span.End()

    span.SetAttributes(
        attribute.String("db.query", query),
        attribute.Int("db.args.count", len(args)),
    )

    // 路由决策
    table, shardKey, err := r.resolver.ExtractShardKey(query, args)
    if err != nil {
        span.RecordError(err)
        return nil, err
    }

    span.SetAttributes(
        attribute.String("db.table", table),
        attribute.String("db.shard_key", fmt.Sprint(shardKey)),
    )

    // ... 执行查询
}
```

### 7.3 日志记录

```go
type RouterLogger struct {
    logger log.Logger
    level  log.Level
}

func (l *RouterLogger) LogRouting(
    query string,
    table string,
    shardKey any,
    shardIdx int,
    isPrimary bool,
    duration time.Duration,
) {
    if l.level >= log.LevelDebug {
        l.logger.Debug("database routing",
            "query", truncateQuery(query, 100),
            "table", table,
            "shard_key", shardKey,
            "shard_idx", shardIdx,
            "is_primary", isPrimary,
            "duration_us", duration.Microseconds(),
        )
    }
}
```

## 8. 配置管理

### 8.1 配置文件格式

```yaml
database:
  # 是否启用分片
  sharding_enabled: true

  # 是否启用读写分离
  read_write_split_enabled: true

  # 跨分片查询策略: deny | first | all
  cross_shard_policy: deny

  # 分片配置
  shards:
    - index: 0
      primary:
        driver: mysql
        dsn: "user:pass@tcp(shard0-primary:3306)/db"
        max_open_conns: 20
        max_idle_conns: 10
      replicas:
        - driver: mysql
          dsn: "user:pass@tcp(shard0-replica1:3306)/db"
        - driver: mysql
          dsn: "user:pass@tcp(shard0-replica2:3306)/db"

    - index: 1
      primary:
        driver: mysql
        dsn: "user:pass@tcp(shard1-primary:3306)/db"
      replicas:
        - driver: mysql
          dsn: "user:pass@tcp(shard1-replica1:3306)/db"

  # 分片规则
  sharding_rules:
    - table: users
      shard_key: user_id
      strategy: hash
      num_shards: 2

    - table: orders
      shard_key: user_id
      strategy: hash
      num_shards: 2

    - table: logs
      shard_key: created_at
      strategy: range
      num_shards: 4
      ranges:
        - start: "2024-01-01"
          end: "2024-04-01"
          shard: 0
        - start: "2024-04-01"
          end: "2024-07-01"
          shard: 1

  # 负载均衡配置
  load_balancer:
    type: round_robin  # round_robin | random | least_conn | weighted

  # 健康检查配置
  health_check:
    enabled: true
    interval: 30s
    timeout: 5s

  # 故障转移配置
  failover:
    enabled: true
    retry_count: 3
    retry_delay: 100ms
```

### 8.2 环境变量配置

```bash
# 启用分片
DB_SHARDING_ENABLED=true

# 分片数量
DB_NUM_SHARDS=4

# 分片 0 - 主库
DB_SHARD_0_PRIMARY_DSN=user:pass@tcp(shard0-primary:3306)/db

# 分片 0 - 从库
DB_SHARD_0_REPLICA_1_DSN=user:pass@tcp(shard0-replica1:3306)/db
DB_SHARD_0_REPLICA_2_DSN=user:pass@tcp(shard0-replica2:3306)/db

# 分片规则
DB_SHARDING_RULE_USERS_KEY=user_id
DB_SHARDING_RULE_USERS_STRATEGY=hash
```

### 8.3 代码配置

```go
config := cluster.Config{
    EnableSharding:   true,
    EnableReadWrite:  true,
    CrossShardPolicy: cluster.CrossShardDeny,

    Shards: []cluster.ShardConfig{
        {
            Index: 0,
            Primary: cluster.DataSource{
                Driver: "mysql",
                DSN:    os.Getenv("DB_SHARD_0_PRIMARY_DSN"),
                Config: db.DefaultConfig("mysql", ""),
            },
            Replicas: []cluster.DataSource{
                {
                    Driver: "mysql",
                    DSN:    os.Getenv("DB_SHARD_0_REPLICA_1_DSN"),
                },
            },
        },
    },

    ShardingRules: []cluster.ShardingRule{
        {
            TableName:      "users",
            ShardKeyColumn: "user_id",
            Strategy:       sharding.NewHashStrategy(),
            NumShards:      2,
        },
    },

    LoadBalancer: cluster.NewRoundRobinBalancer(),

    HealthCheck: cluster.HealthCheckConfig{
        Enabled:  true,
        Interval: 30 * time.Second,
        Timeout:  5 * time.Second,
    },
}

database, err := cluster.New(config)
```

## 9. 实现计划

### 9.1 阶段 1: 读写分离 (2-3 周)

**目标**: 实现基本的读写分离功能

**交付物**:
- [ ] `rw/splitter.go` - 读写分离核心逻辑
- [ ] `rw/loadbalancer.go` - 负载均衡实现(轮询、随机、最少连接)
- [ ] `rw/policy.go` - 路由策略(SQL类型、上下文提示)
- [ ] `rw/health.go` - 健康检查
- [ ] `cluster.go` - 集群管理
- [ ] 完整的单元测试(覆盖率 > 85%)
- [ ] 集成测试(使用 docker-compose 模拟主从)
- [ ] 文档和示例

**技术要点**:
- 轻量级 SQL 解析(关键字匹配,不使用完整解析器)
- 原子操作保证负载均衡的线程安全
- 健康检查使用独立 goroutine
- 支持强制路由到主库(context hint)

### 9.2 阶段 2: 分片基础 (3-4 周)

**目标**: 实现分片路由和基本分片策略

**交付物**:
- [ ] `sharding/strategy.go` - 分片策略接口
- [ ] `sharding/hash.go` - 哈希分片
- [ ] `sharding/mod.go` - 取模分片
- [ ] `sharding/range.go` - 范围分片
- [ ] `sharding/list.go` - 列表分片
- [ ] `sharding/resolver.go` - 分片键解析
- [ ] `router.go` - 统一路由器
- [ ] 单元测试和集成测试
- [ ] 性能基准测试

**技术要点**:
- 分片键提取使用正则表达式
- 支持常见 SQL 模式(INSERT, UPDATE, DELETE, SELECT)
- 路由决策必须在微秒级完成
- 支持分片键缓存(避免重复解析)

### 9.3 阶段 3: 高级功能 (2-3 周)

**目标**: 完善功能,生产可用

**交付物**:
- [ ] SQL 改写支持(物理分表场景)
- [ ] 跨分片查询支持(并发查询 + 结果合并)
- [ ] 分布式追踪集成(OpenTelemetry)
- [ ] Prometheus 指标导出
- [ ] 配置文件加载(YAML/JSON)
- [ ] 故障转移和重试策略
- [ ] 压力测试报告

**技术要点**:
- SQL 改写保证语义不变
- 跨分片查询限制结果集大小(防止 OOM)
- 监控指标包含分片级别细节
- 支持热重载配置(不中断服务)

### 9.4 阶段 4: 优化和文档 (1-2 周)

**目标**: 性能优化和完善文档

**交付物**:
- [ ] 性能优化(减少内存分配、优化锁粒度)
- [ ] 完整的 API 文档(godoc)
- [ ] 使用指南和最佳实践
- [ ] 迁移指南(从单库迁移到分片)
- [ ] 故障排查手册
- [ ] 生产环境检查清单

## 10. 测试策略

### 10.1 单元测试

- 每个策略独立测试
- Mock 数据库连接
- 覆盖率目标: > 85%

### 10.2 集成测试

- 使用 docker-compose 启动真实 MySQL 实例
- 测试主从复制场景
- 测试分片数据分布

### 10.3 性能测试

- 路由决策延迟 < 100μs
- 吞吐量不低于单库 95%
- 并发 1000 连接稳定性测试

### 10.4 故障测试

- 从库宕机自动降级
- 主库宕机快速失败
- 网络分区恢复

## 11. 向后兼容性

### 11.1 兼容保证

- 现有 `db.DB` 接口不变
- 现有代码无需修改即可工作
- 新功能通过可选配置启用

### 11.2 迁移路径

```go
// 旧代码 (继续工作)
db, err := db.Open(db.DefaultConfig("mysql", dsn))

// 新代码 (启用读写分离)
clusterDB, err := cluster.New(cluster.Config{
    Primary:  cluster.DataSource{Driver: "mysql", DSN: primaryDSN},
    Replicas: []cluster.DataSource{{Driver: "mysql", DSN: replicaDSN}},
})

// 新代码 (启用分片)
shardedDB, err := cluster.New(cluster.Config{
    Shards: []cluster.ShardConfig{...},
    ShardingRules: []cluster.ShardingRule{...},
})
```

## 12. 安全考虑

### 12.1 SQL 注入防护

- 继承 `database/sql` 的参数绑定机制
- 不拼接 SQL 字符串
- 分片键提取不执行 SQL

### 12.2 访问控制

- 支持不同分片使用不同数据库凭证
- 支持只读用户连接从库

### 12.3 审计日志

- 记录所有跨分片查询
- 记录路由决策失败

## 13. 限制和已知问题

### 13.1 不支持的功能

- ❌ 分布式事务(跨分片)
- ❌ 跨分片 JOIN
- ❌ 跨分片聚合(SUM/AVG)
- ❌ 自动数据迁移

### 13.2 性能限制

- SQL 解析使用正则,复杂 SQL 可能解析失败
- 跨分片查询需要加载所有结果到内存

### 13.3 使用建议

- 优先设计支持分片的数据模型
- 避免跨分片查询,在应用层聚合
- 分片键选择要考虑查询模式
- 定期检查分片数据倾斜

## 14. 参考资料

### 14.1 类似项目

- [Vitess](https://vitess.io/) - MySQL 集群方案
- [ShardingSphere](https://shardingsphere.apache.org/) - Java 生态分片中间件
- [Gorm Sharding](https://github.com/go-gorm/sharding) - Gorm 分表插件

### 14.2 相关文档

- [MySQL Replication](https://dev.mysql.com/doc/refman/8.0/en/replication.html)
- [Database Sharding](https://en.wikipedia.org/wiki/Shard_(database_architecture))
- [Consistent Hashing](https://en.wikipedia.org/wiki/Consistent_hashing)

---

**文档版本**: v1.0
**最后更新**: 2026-01-30
**作者**: Claude (AI Assistant)
**状态**: 设计规划阶段
