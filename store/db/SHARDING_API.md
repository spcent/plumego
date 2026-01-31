# 分库分表 API 接口定义与使用示例

## 1. 核心接口定义

### 1.1 集群配置

```go
package cluster

import (
    "time"
    "github.com/spcent/plumego/store/db"
    "github.com/spcent/plumego/store/db/sharding"
)

// Config 集群配置
type Config struct {
    // 单库模式配置(与分片模式互斥)
    Primary  *DataSource
    Replicas []DataSource

    // 分片模式配置
    Shards []ShardConfig

    // 分片规则
    ShardingRules []ShardingRule

    // 功能开关
    EnableSharding  bool  // 是否启用分片
    EnableReadWrite bool  // 是否启用读写分离

    // 路由策略
    CrossShardPolicy CrossShardPolicy  // 跨分片查询策略
    RoutingPolicy    RoutingPolicy     // 读写路由策略

    // 负载均衡
    LoadBalancer LoadBalancer  // 负载均衡器

    // 健康检查
    HealthCheck HealthCheckConfig

    // 故障转移
    Failover FailoverConfig

    // 监控
    Metrics MetricsConfig

    // 高级配置
    SQLParseTimeout time.Duration  // SQL 解析超时
    RouterTimeout   time.Duration  // 路由决策超时
}

// DataSource 数据源配置
type DataSource struct {
    Driver string      // 数据库驱动
    DSN    string      // 数据源名称
    Config db.Config   // 连接池配置
    Weight int         // 权重(用于加权负载均衡)
}

// ShardConfig 分片配置
type ShardConfig struct {
    Index    int          // 分片索引
    Primary  DataSource   // 主库
    Replicas []DataSource // 从库列表
}

// ShardingRule 分片规则
type ShardingRule struct {
    TableName        string                  // 表名
    ShardKeyColumn   string                  // 分片键列名
    Strategy         sharding.Strategy       // 分片策略
    NumShards        int                     // 分片数量
    ActualTableNames []string                // 实际表名(用于已分表场景)
    DatabaseSharding bool                    // 是否分库(false 表示只分表)
}

// CrossShardPolicy 跨分片查询策略
type CrossShardPolicy int

const (
    CrossShardDeny  CrossShardPolicy = iota  // 拒绝跨分片查询
    CrossShardFirst                          // 只查询第一个分片
    CrossShardAll                            // 查询所有分片
)

// HealthCheckConfig 健康检查配置
type HealthCheckConfig struct {
    Enabled  bool          // 是否启用
    Interval time.Duration // 检查间隔
    Timeout  time.Duration // 超时时间
}

// FailoverConfig 故障转移配置
type FailoverConfig struct {
    Enabled    bool          // 是否启用
    RetryCount int           // 重试次数
    RetryDelay time.Duration // 重试延迟
}

// MetricsConfig 监控配置
type MetricsConfig struct {
    Enabled        bool   // 是否启用
    EnableTracing  bool   // 是否启用追踪
    ServiceName    string // 服务名称(用于追踪)
}
```

### 1.2 负载均衡接口

```go
package cluster

// LoadBalancer 负载均衡器接口
type LoadBalancer interface {
    // Next 选择下一个副本索引
    Next(replicas []Replica) (int, error)

    // Reset 重置状态(用于配置变更)
    Reset()
}

// Replica 副本信息
type Replica struct {
    Index     int           // 索引
    DB        db.DB         // 数据库连接
    Weight    int           // 权重
    IsHealthy bool          // 是否健康
    Stats     ReplicaStats  // 统计信息
}

// ReplicaStats 副本统计信息
type ReplicaStats struct {
    OpenConnections int           // 打开的连接数
    InUse           int           // 使用中的连接数
    WaitCount       int64         // 等待次数
    WaitDuration    time.Duration // 等待时长
}

// 预定义负载均衡器
func NewRoundRobinBalancer() LoadBalancer
func NewRandomBalancer() LoadBalancer
func NewLeastConnBalancer() LoadBalancer
func NewWeightedBalancer(weights []int) LoadBalancer
```

### 1.3 路由策略接口

```go
package cluster

// RoutingPolicy 路由策略接口
type RoutingPolicy interface {
    // ShouldUsePrimary 判断是否应该使用主库
    ShouldUsePrimary(ctx context.Context, query string) bool
}

// 预定义路由策略
func NewSQLTypePolicy() RoutingPolicy          // 基于 SQL 类型
func NewContextAwarePolicy() RoutingPolicy     // 基于上下文提示
func NewTransactionAwarePolicy(base RoutingPolicy) RoutingPolicy  // 事务感知
func NewCompositePolicy(policies ...RoutingPolicy) RoutingPolicy  // 组合策略
```

### 1.4 分片策略接口

```go
package sharding

// Strategy 分片策略接口
type Strategy interface {
    // Shard 计算分片索引
    Shard(key any, numShards int) (int, error)

    // ShardRange 计算分片范围(用于范围查询)
    ShardRange(start, end any, numShards int) ([]int, error)

    // Name 返回策略名称
    Name() string
}

// 预定义分片策略
func NewHashStrategy() Strategy                // FNV-1a 哈希
func NewMurmurHashStrategy() Strategy          // MurmurHash3
func NewModStrategy() Strategy                 // 简单取模
func NewRangeStrategy(ranges []RangeDefinition) Strategy
func NewListStrategy(mapping map[any]int) Strategy

// RangeDefinition 范围定义
type RangeDefinition struct {
    Start any   // 起始值(包含)
    End   any   // 结束值(不包含)
    Shard int   // 目标分片
}
```

## 2. 使用示例

### 2.1 场景 1: 单库 + 读写分离

```go
package main

import (
    "context"
    "database/sql"
    "log"

    "github.com/spcent/plumego/store/db/cluster"
    _ "github.com/go-sql-driver/mysql"
)

func main() {
    // 配置读写分离集群
    clusterDB, err := cluster.New(cluster.Config{
        Primary: &cluster.DataSource{
            Driver: "mysql",
            DSN:    "user:pass@tcp(primary.db:3306)/myapp",
            Config: db.DefaultConfig("mysql", ""),
        },
        Replicas: []cluster.DataSource{
            {
                Driver: "mysql",
                DSN:    "user:pass@tcp(replica1.db:3306)/myapp",
                Weight: 1,
            },
            {
                Driver: "mysql",
                DSN:    "user:pass@tcp(replica2.db:3306)/myapp",
                Weight: 2,  // 更高权重
            },
        },

        EnableReadWrite: true,
        LoadBalancer:    cluster.NewWeightedBalancer([]int{1, 2}),

        HealthCheck: cluster.HealthCheckConfig{
            Enabled:  true,
            Interval: 30 * time.Second,
            Timeout:  5 * time.Second,
        },
    })
    if err != nil {
        log.Fatal(err)
    }
    defer clusterDB.Close()

    ctx := context.Background()

    // 写操作 -> 自动路由到主库
    result, err := clusterDB.ExecContext(ctx,
        "INSERT INTO users (name, email) VALUES (?, ?)",
        "Alice", "alice@example.com",
    )
    if err != nil {
        log.Fatal(err)
    }

    // 读操作 -> 自动路由到从库(加权轮询)
    rows, err := clusterDB.QueryContext(ctx,
        "SELECT id, name, email FROM users WHERE id = ?",
        123,
    )
    if err != nil {
        log.Fatal(err)
    }
    defer rows.Close()

    // SELECT FOR UPDATE -> 自动路由到主库
    rows, err = clusterDB.QueryContext(ctx,
        "SELECT balance FROM accounts WHERE user_id = ? FOR UPDATE",
        123,
    )

    // 强制使用主库(用于需要强一致性的场景)
    ctx = cluster.ForcePrimary(ctx)
    rows, err = clusterDB.QueryContext(ctx,
        "SELECT * FROM users WHERE id = ?",
        123,
    )
}
```

### 2.2 场景 2: 哈希分库 + 读写分离

```go
package main

import (
    "context"
    "log"

    "github.com/spcent/plumego/store/db/cluster"
    "github.com/spcent/plumego/store/db/sharding"
    _ "github.com/go-sql-driver/mysql"
)

func main() {
    // 配置 4 个分片,每个分片有 1 主 2 从
    clusterDB, err := cluster.New(cluster.Config{
        EnableSharding:  true,
        EnableReadWrite: true,
        CrossShardPolicy: cluster.CrossShardDeny,  // 拒绝跨分片查询

        Shards: []cluster.ShardConfig{
            // Shard 0
            {
                Index: 0,
                Primary: cluster.DataSource{
                    Driver: "mysql",
                    DSN:    "user:pass@tcp(shard0-primary:3306)/db",
                },
                Replicas: []cluster.DataSource{
                    {Driver: "mysql", DSN: "user:pass@tcp(shard0-replica1:3306)/db"},
                    {Driver: "mysql", DSN: "user:pass@tcp(shard0-replica2:3306)/db"},
                },
            },
            // Shard 1
            {
                Index: 1,
                Primary: cluster.DataSource{
                    Driver: "mysql",
                    DSN:    "user:pass@tcp(shard1-primary:3306)/db",
                },
                Replicas: []cluster.DataSource{
                    {Driver: "mysql", DSN: "user:pass@tcp(shard1-replica1:3306)/db"},
                },
            },
            // Shard 2, 3 配置类似...
        },

        // 分片规则
        ShardingRules: []cluster.ShardingRule{
            {
                TableName:      "users",
                ShardKeyColumn: "user_id",
                Strategy:       sharding.NewHashStrategy(),
                NumShards:      4,
            },
            {
                TableName:      "orders",
                ShardKeyColumn: "user_id",  // 与 users 使用相同分片键
                Strategy:       sharding.NewHashStrategy(),
                NumShards:      4,
            },
        },

        LoadBalancer: cluster.NewRoundRobinBalancer(),

        HealthCheck: cluster.HealthCheckConfig{
            Enabled:  true,
            Interval: 30 * time.Second,
        },
    })
    if err != nil {
        log.Fatal(err)
    }
    defer clusterDB.Close()

    ctx := context.Background()

    // 写入用户数据
    // user_id=12345 -> hash(12345) % 4 = 1 -> shard_1.primary
    _, err = clusterDB.ExecContext(ctx,
        "INSERT INTO users (user_id, name, email) VALUES (?, ?, ?)",
        12345, "Alice", "alice@example.com",
    )
    if err != nil {
        log.Fatal(err)
    }

    // 查询用户数据
    // user_id=12345 -> shard_1.replica (轮询选择)
    var name, email string
    err = clusterDB.QueryRowContext(ctx,
        "SELECT name, email FROM users WHERE user_id = ?",
        12345,
    ).Scan(&name, &email)
    if err != nil {
        log.Fatal(err)
    }

    // 创建订单(使用相同的 user_id,路由到同一分片)
    _, err = clusterDB.ExecContext(ctx,
        "INSERT INTO orders (order_id, user_id, amount) VALUES (?, ?, ?)",
        "ORD123", 12345, 99.99,
    )

    // 跨分片查询会被拒绝(没有分片键)
    _, err = clusterDB.QueryContext(ctx,
        "SELECT COUNT(*) FROM users",
    )
    // err == ErrCrossShardQueryNotSupported
}
```

### 2.3 场景 3: 范围分表(按月分表)

```go
package main

import (
    "context"
    "log"
    "time"

    "github.com/spcent/plumego/store/db/cluster"
    "github.com/spcent/plumego/store/db/sharding"
)

func main() {
    // 日志表按月分表: logs_202401, logs_202402, ...
    clusterDB, err := cluster.New(cluster.Config{
        EnableSharding: true,

        Shards: []cluster.ShardConfig{
            {
                Index: 0,
                Primary: cluster.DataSource{
                    Driver: "mysql",
                    DSN:    "user:pass@tcp(localhost:3306)/logs_db",
                },
            },
        },

        ShardingRules: []cluster.ShardingRule{
            {
                TableName:        "logs",
                ShardKeyColumn:   "created_at",
                Strategy: sharding.NewRangeStrategy([]sharding.RangeDefinition{
                    {
                        Start: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
                        End:   time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC),
                        Shard: 0,
                    },
                    {
                        Start: time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC),
                        End:   time.Date(2024, 3, 1, 0, 0, 0, 0, time.UTC),
                        Shard: 1,
                    },
                    // ... 更多月份
                }),
                NumShards: 12,
                ActualTableNames: []string{
                    "logs_202401", "logs_202402", "logs_202403",
                    "logs_202404", "logs_202405", "logs_202406",
                    "logs_202407", "logs_202408", "logs_202409",
                    "logs_202410", "logs_202411", "logs_202412",
                },
                DatabaseSharding: false,  // 只分表,不分库
            },
        },
    })
    if err != nil {
        log.Fatal(err)
    }
    defer clusterDB.Close()

    ctx := context.Background()

    // 插入 1 月的日志
    // created_at=2024-01-15 -> shard 0 -> logs_202401
    _, err = clusterDB.ExecContext(ctx,
        "INSERT INTO logs (created_at, level, message) VALUES (?, ?, ?)",
        time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
        "INFO",
        "User logged in",
    )
    // SQL 自动改写为: INSERT INTO logs_202401 ...

    // 查询 2 月的日志
    rows, err := clusterDB.QueryContext(ctx,
        "SELECT level, message FROM logs WHERE created_at >= ? AND created_at < ?",
        time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC),
        time.Date(2024, 2, 2, 0, 0, 0, 0, time.UTC),
    )
    // SQL 自动改写为: SELECT ... FROM logs_202402 WHERE ...
}
```

### 2.4 场景 4: 列表分片(按租户/商户)

```go
package main

import (
    "context"
    "log"

    "github.com/spcent/plumego/store/db/cluster"
    "github.com/spcent/plumego/store/db/sharding"
)

func main() {
    // 多租户场景:大客户独立分片,小客户共享分片
    clusterDB, err := cluster.New(cluster.Config{
        EnableSharding: true,

        Shards: []cluster.ShardConfig{
            {Index: 0, Primary: cluster.DataSource{Driver: "mysql", DSN: "..."}},  // 大客户 A
            {Index: 1, Primary: cluster.DataSource{Driver: "mysql", DSN: "..."}},  // 大客户 B
            {Index: 2, Primary: cluster.DataSource{Driver: "mysql", DSN: "..."}},  // 小客户共享
        },

        ShardingRules: []cluster.ShardingRule{
            {
                TableName:      "tenant_data",
                ShardKeyColumn: "tenant_id",
                Strategy: sharding.NewListStrategy(map[any]int{
                    "tenant_big_a":    0,  // 大客户 A -> 独立分片 0
                    "tenant_big_b":    1,  // 大客户 B -> 独立分片 1
                    "tenant_small_1":  2,  // 小客户 -> 共享分片 2
                    "tenant_small_2":  2,
                    "tenant_small_3":  2,
                }),
                NumShards: 3,
            },
        },
    })
    if err != nil {
        log.Fatal(err)
    }
    defer clusterDB.Close()

    ctx := context.Background()

    // 大客户 A 的数据 -> 分片 0
    _, err = clusterDB.ExecContext(ctx,
        "INSERT INTO tenant_data (tenant_id, key, value) VALUES (?, ?, ?)",
        "tenant_big_a", "config", "{}",
    )

    // 小客户的数据 -> 分片 2(共享)
    _, err = clusterDB.ExecContext(ctx,
        "INSERT INTO tenant_data (tenant_id, key, value) VALUES (?, ?, ?)",
        "tenant_small_1", "config", "{}",
    )
}
```

### 2.5 场景 5: 事务处理

```go
package main

import (
    "context"
    "database/sql"
    "log"

    "github.com/spcent/plumego/store/db/cluster"
)

func main() {
    clusterDB, err := cluster.New(cluster.Config{
        // ... 配置
    })
    if err != nil {
        log.Fatal(err)
    }
    defer clusterDB.Close()

    ctx := context.Background()

    // 方式 1: 手动管理事务
    tx, err := clusterDB.BeginTx(ctx, &sql.TxOptions{
        Isolation: sql.LevelReadCommitted,
    })
    if err != nil {
        log.Fatal(err)
    }

    // 事务中的所有操作都路由到主库
    _, err = tx.ExecContext(ctx,
        "UPDATE accounts SET balance = balance - 100 WHERE user_id = ?",
        12345,
    )
    if err != nil {
        tx.Rollback()
        log.Fatal(err)
    }

    _, err = tx.ExecContext(ctx,
        "INSERT INTO transactions (user_id, amount) VALUES (?, ?)",
        12345, -100,
    )
    if err != nil {
        tx.Rollback()
        log.Fatal(err)
    }

    if err := tx.Commit(); err != nil {
        log.Fatal(err)
    }

    // 方式 2: 使用 WithTransaction 辅助函数
    err = db.WithTransaction(ctx, clusterDB, nil, func(tx *sql.Tx) error {
        _, err := tx.ExecContext(ctx,
            "UPDATE accounts SET balance = balance + 100 WHERE user_id = ?",
            12345,
        )
        if err != nil {
            return err
        }

        _, err = tx.ExecContext(ctx,
            "INSERT INTO transactions (user_id, amount) VALUES (?, ?)",
            12345, 100,
        )
        return err
    })
    if err != nil {
        log.Fatal(err)
    }
}
```

### 2.6 场景 6: 上下文控制路由

```go
package main

import (
    "context"
    "log"

    "github.com/spcent/plumego/store/db/cluster"
)

func main() {
    clusterDB, err := cluster.New(cluster.Config{
        EnableReadWrite: true,
        // ... 其他配置
    })
    if err != nil {
        log.Fatal(err)
    }
    defer clusterDB.Close()

    // 场景 1: 刚写入后立即读取,需要强一致性
    ctx := context.Background()

    _, err = clusterDB.ExecContext(ctx,
        "INSERT INTO users (user_id, name) VALUES (?, ?)",
        12345, "Alice",
    )
    if err != nil {
        log.Fatal(err)
    }

    // 强制从主库读取(避免主从延迟)
    ctx = cluster.ForcePrimary(ctx)
    var name string
    err = clusterDB.QueryRowContext(ctx,
        "SELECT name FROM users WHERE user_id = ?",
        12345,
    ).Scan(&name)

    // 场景 2: 数据导出,可以接受延迟,优先使用从库
    ctx = cluster.PreferReplica(ctx)
    rows, err := clusterDB.QueryContext(ctx,
        "SELECT * FROM users WHERE created_at >= ?",
        time.Now().AddDate(0, 0, -30),
    )

    // 场景 3: 自定义路由提示
    ctx = cluster.WithRoutingHint(ctx, cluster.RoutingHint{
        PreferShard:   1,           // 优先使用分片 1
        PreferReplica: 0,           // 优先使用第 0 个从库
        MaxLatency:    50 * time.Millisecond,  // 最大延迟要求
    })
}
```

### 2.7 场景 7: 健康检查和监控

```go
package main

import (
    "context"
    "log"
    "time"

    "github.com/spcent/plumego/store/db/cluster"
)

func main() {
    clusterDB, err := cluster.New(cluster.Config{
        EnableReadWrite: true,

        HealthCheck: cluster.HealthCheckConfig{
            Enabled:  true,
            Interval: 30 * time.Second,
            Timeout:  5 * time.Second,
        },

        Metrics: cluster.MetricsConfig{
            Enabled:       true,
            EnableTracing: true,
            ServiceName:   "myapp",
        },
    })
    if err != nil {
        log.Fatal(err)
    }
    defer clusterDB.Close()

    // 获取健康状态
    health := clusterDB.Health()
    log.Printf("Cluster health: %+v", health)
    // Output:
    // {
    //   "status": "healthy",
    //   "primary": {"status": "healthy", "latency": "2ms"},
    //   "replicas": [
    //     {"index": 0, "status": "healthy", "latency": "3ms"},
    //     {"index": 1, "status": "unhealthy", "error": "connection refused"}
    //   ],
    //   "shards": [...]
    // }

    // 获取指标
    metrics := clusterDB.Metrics()
    log.Printf("Total queries: %d", metrics.TotalQueries)
    log.Printf("Primary queries: %d", metrics.PrimaryQueries)
    log.Printf("Replica queries: %d", metrics.ReplicaQueries)
    log.Printf("Avg routing latency: %v", metrics.AvgRoutingLatency)

    // 手动触发健康检查
    ctx := context.Background()
    if err := clusterDB.Ping(ctx); err != nil {
        log.Printf("Ping failed: %v", err)
    }
}
```

## 3. 高级用法

### 3.1 自定义分片策略

```go
package main

import (
    "hash/crc32"
    "github.com/spcent/plumego/store/db/sharding"
)

// CustomShardingStrategy 自定义分片策略
type CustomShardingStrategy struct {
    // 自定义字段
}

func (s *CustomShardingStrategy) Shard(key any, numShards int) (int, error) {
    // 实现自定义分片逻辑
    str, ok := key.(string)
    if !ok {
        return -1, sharding.ErrInvalidShardKey
    }

    // 例如:使用 CRC32
    hash := crc32.ChecksumIEEE([]byte(str))
    return int(hash % uint32(numShards)), nil
}

func (s *CustomShardingStrategy) ShardRange(start, end any, numShards int) ([]int, error) {
    // 对于自定义策略,可能需要查询所有分片
    shards := make([]int, numShards)
    for i := 0; i < numShards; i++ {
        shards[i] = i
    }
    return shards, nil
}

func (s *CustomShardingStrategy) Name() string {
    return "custom_crc32"
}

// 使用自定义策略
func main() {
    clusterDB, err := cluster.New(cluster.Config{
        ShardingRules: []cluster.ShardingRule{
            {
                TableName:      "users",
                ShardKeyColumn: "user_id",
                Strategy:       &CustomShardingStrategy{},
                NumShards:      4,
            },
        },
    })
}
```

### 3.2 自定义负载均衡器

```go
package main

import (
    "math/rand"
    "sync"
    "github.com/spcent/plumego/store/db/cluster"
)

// WeightedRandomBalancer 加权随机负载均衡
type WeightedRandomBalancer struct {
    weights []int
    total   int
    mu      sync.Mutex
    rng     *rand.Rand
}

func NewWeightedRandomBalancer(weights []int) *WeightedRandomBalancer {
    total := 0
    for _, w := range weights {
        total += w
    }
    return &WeightedRandomBalancer{
        weights: weights,
        total:   total,
        rng:     rand.New(rand.NewSource(time.Now().UnixNano())),
    }
}

func (b *WeightedRandomBalancer) Next(replicas []cluster.Replica) (int, error) {
    if len(replicas) == 0 {
        return -1, cluster.ErrNoReplicasAvailable
    }

    b.mu.Lock()
    defer b.mu.Unlock()

    // 随机选择
    r := b.rng.Intn(b.total)
    sum := 0
    for i, w := range b.weights {
        sum += w
        if r < sum && i < len(replicas) && replicas[i].IsHealthy {
            return i, nil
        }
    }

    // 降级:返回第一个健康的副本
    for i, r := range replicas {
        if r.IsHealthy {
            return i, nil
        }
    }

    return -1, cluster.ErrNoHealthyReplicas
}

func (b *WeightedRandomBalancer) Reset() {
    // 重置状态
}
```

### 3.3 自定义路由策略

```go
package main

import (
    "context"
    "strings"
    "github.com/spcent/plumego/store/db/cluster"
)

// CustomRoutingPolicy 自定义路由策略
type CustomRoutingPolicy struct {
    // 业务规则配置
}

func (p *CustomRoutingPolicy) ShouldUsePrimary(ctx context.Context, query string) bool {
    // 检查上下文提示
    if cluster.IsForcePrimary(ctx) {
        return true
    }

    // 检查 SQL 类型
    upper := strings.ToUpper(strings.TrimSpace(query))

    // 所有写操作 -> 主库
    writeKeywords := []string{"INSERT", "UPDATE", "DELETE", "CREATE", "ALTER", "DROP"}
    for _, kw := range writeKeywords {
        if strings.HasPrefix(upper, kw) {
            return true
        }
    }

    // SELECT FOR UPDATE -> 主库
    if strings.Contains(upper, "FOR UPDATE") {
        return true
    }

    // 业务自定义规则:重要表的查询走主库
    criticalTables := []string{"ACCOUNTS", "ORDERS"}
    for _, table := range criticalTables {
        if strings.Contains(upper, table) {
            return true
        }
    }

    // 其他读操作 -> 从库
    return false
}
```

## 4. 迁移指南

### 4.1 从单库迁移到读写分离

```go
// Step 1: 原始代码(单库)
db, err := sql.Open("mysql", "user:pass@tcp(localhost:3306)/db")
if err != nil {
    log.Fatal(err)
}

// Step 2: 包装为 plumego db
import "github.com/spcent/plumego/store/db"
database, err := db.Open(db.DefaultConfig("mysql", "user:pass@tcp(localhost:3306)/db"))

// Step 3: 启用读写分离(代码无需修改!)
import "github.com/spcent/plumego/store/db/cluster"
database, err := cluster.New(cluster.Config{
    Primary: &cluster.DataSource{
        Driver: "mysql",
        DSN:    "user:pass@tcp(primary:3306)/db",
    },
    Replicas: []cluster.DataSource{
        {Driver: "mysql", DSN: "user:pass@tcp(replica:3306)/db"},
    },
    EnableReadWrite: true,
})

// 业务代码完全不变
rows, err := database.QueryContext(ctx, "SELECT * FROM users")
```

### 4.2 从读写分离迁移到分片

```go
// Step 1: 评估分片需求
// - 确定分片键(通常是 user_id, tenant_id 等)
// - 确定分片数量(建议 2 的幂次,便于扩容)
// - 确定分片策略(哈希/范围/列表)

// Step 2: 准备分片数据库
// - 创建多个数据库实例
// - 迁移数据(可以使用一致性哈希减少迁移量)

// Step 3: 更新配置
database, err := cluster.New(cluster.Config{
    EnableSharding:  true,
    EnableReadWrite: true,

    Shards: []cluster.ShardConfig{
        {Index: 0, Primary: ..., Replicas: ...},
        {Index: 1, Primary: ..., Replicas: ...},
        // ... 更多分片
    },

    ShardingRules: []cluster.ShardingRule{
        {
            TableName:      "users",
            ShardKeyColumn: "user_id",
            Strategy:       sharding.NewHashStrategy(),
            NumShards:      4,
        },
    },
})

// Step 4: 修改查询(确保包含分片键)
// ❌ 坏的查询(没有分片键,会失败或查询所有分片)
rows, err := db.QueryContext(ctx, "SELECT * FROM users WHERE name = ?", "Alice")

// ✅ 好的查询(包含分片键)
rows, err := db.QueryContext(ctx, "SELECT * FROM users WHERE user_id = ?", 12345)
```

## 5. 最佳实践

### 5.1 分片键选择

1. **选择不可变的字段**: 如 user_id, order_id(避免分片键变更导致数据迁移)
2. **选择查询中常用的字段**: 确保大部分查询都能包含分片键
3. **考虑数据分布均匀性**: 避免热点分片
4. **考虑业务关联性**: 相关数据应在同一分片(如 user + orders)

```go
// ✅ 好的分片键选择
ShardingRule{
    TableName:      "orders",
    ShardKeyColumn: "user_id",  // 用户的所有订单在同一分片
    Strategy:       sharding.NewHashStrategy(),
}

// ❌ 坏的分片键选择
ShardingRule{
    TableName:      "orders",
    ShardKeyColumn: "created_at",  // 查询通常用 user_id,不用时间
    Strategy:       sharding.NewRangeStrategy(...),
}
```

### 5.2 事务处理

1. **单分片事务**: 正常使用 `BeginTx`
2. **跨分片事务**: **不支持**,需要在应用层实现最终一致性

```go
// ✅ 单分片事务(支持)
tx, err := db.BeginTx(ctx, nil)
tx.ExecContext(ctx, "UPDATE users SET balance = balance - 100 WHERE user_id = ?", 12345)
tx.ExecContext(ctx, "INSERT INTO transactions (user_id, amount) VALUES (?, ?)", 12345, -100)
tx.Commit()

// ❌ 跨分片事务(不支持)
tx, err := db.BeginTx(ctx, nil)
tx.ExecContext(ctx, "UPDATE users SET balance = balance - 100 WHERE user_id = ?", 12345)  // 分片 1
tx.ExecContext(ctx, "UPDATE users SET balance = balance + 100 WHERE user_id = ?", 67890)  // 分片 2
tx.Commit()  // 失败!

// ✅ 应用层最终一致性
// 1. 写入转账记录(带状态)
// 2. 异步任务处理双方账户变更
// 3. 失败重试 + 幂等性保证
```

### 5.3 查询优化

```go
// ❌ 避免:没有分片键的查询
rows, err := db.QueryContext(ctx, "SELECT COUNT(*) FROM users")
// 需要查询所有分片,性能差

// ✅ 推荐:包含分片键
rows, err := db.QueryContext(ctx, "SELECT * FROM users WHERE user_id = ?", 12345)
// 精准路由到单个分片

// ✅ 推荐:如果必须统计,使用分片聚合
var totalCount int64
for shardIdx := 0; shardIdx < numShards; shardIdx++ {
    ctx := cluster.WithShardHint(ctx, shardIdx)
    var count int64
    db.QueryRowContext(ctx, "SELECT COUNT(*) FROM users").Scan(&count)
    totalCount += count
}
```

### 5.4 监控告警

```go
// 设置告警规则
alertRules := []cluster.AlertRule{
    {
        Name:      "replica_down",
        Condition: func(health cluster.HealthStatus) bool {
            return health.ReplicasHealthy < health.ReplicasTotal / 2
        },
        Severity: cluster.Critical,
    },
    {
        Name:      "routing_error_rate_high",
        Condition: func(metrics cluster.Metrics) bool {
            errorRate := float64(metrics.RoutingErrors) / float64(metrics.TotalQueries)
            return errorRate > 0.01  // 1% 错误率
        },
        Severity: cluster.Warning,
    },
}

clusterDB.SetAlertRules(alertRules)
clusterDB.OnAlert(func(alert cluster.Alert) {
    log.Printf("[%s] %s: %s", alert.Severity, alert.Name, alert.Message)
    // 发送到告警系统(PagerDuty, Slack, etc.)
})
```

## 6. 故障排查

### 6.1 常见错误

#### 错误 1: `ErrCrossShardQueryNotSupported`

```
Error: cross-shard query not supported
```

**原因**: 查询没有包含分片键,需要查询所有分片,但配置为拒绝跨分片查询。

**解决方案**:
1. 修改查询,添加分片键
2. 或者修改配置: `CrossShardPolicy: cluster.CrossShardAll`

#### 错误 2: `ErrShardKeyNotFound`

```
Error: shard key not found in query
```

**原因**: SQL 解析器无法从查询中提取分片键值。

**解决方案**:
1. 确保分片键出现在 WHERE 子句中
2. 或者使用上下文提示明确指定分片: `ctx = cluster.WithShardHint(ctx, shardIdx)`

#### 错误 3: `ErrNoHealthyReplicas`

```
Error: no healthy replicas available
```

**原因**: 所有从库都不健康。

**解决方案**:
1. 检查从库连接配置
2. 启用自动降级到主库: `Failover: cluster.FailoverConfig{Enabled: true}`

### 6.2 性能调优

#### 优化 1: 减少路由开销

```go
// 使用连接池预热
clusterDB.Warmup(ctx)

// 启用路由缓存(对于相同模式的 SQL)
config.EnableRoutingCache = true
config.RoutingCacheSize = 1000
```

#### 优化 2: 调整负载均衡

```go
// 根据从库性能调整权重
LoadBalancer: cluster.NewWeightedBalancer([]int{
    1,  // replica-0: 低配置
    2,  // replica-1: 中配置
    4,  // replica-2: 高配置
})
```

#### 优化 3: 连接池调优

```go
Shards: []cluster.ShardConfig{
    {
        Primary: cluster.DataSource{
            Config: db.Config{
                MaxOpenConns:    50,   // 根据负载调整
                MaxIdleConns:    25,
                ConnMaxLifetime: 30 * time.Minute,
                ConnMaxIdleTime: 5 * time.Minute,
            },
        },
    },
}
```

---

**文档版本**: v1.0
**最后更新**: 2026-01-30
**相关文档**: SHARDING_DESIGN.md
