# Read/Write Splitting Example

This example demonstrates the read/write splitting functionality of plumego's database layer.

## Features Demonstrated

1. **Automatic Routing**
   - Write operations (INSERT, UPDATE, DELETE) → Primary database
   - Read operations (SELECT) → Replica databases

2. **Load Balancing**
   - Multiple replicas with round-robin distribution
   - Automatic failover to primary if replicas are down

3. **Health Monitoring**
   - Background health checks for replicas
   - Automatic exclusion of unhealthy replicas

4. **Transaction Safety**
   - All transaction operations use primary database
   - Prevents read-your-writes issues

5. **Context Control**
   - `ForcePrimary()`: Force queries to use primary for strong consistency
   - Automatic detection of `SELECT FOR UPDATE`

## Prerequisites

### Option 1: Using Existing MySQL

If you have MySQL running locally:

```bash
# Create test database
mysql -u root -p -e "CREATE DATABASE IF NOT EXISTS testdb"
```

Edit `main.go` and update the DSN if needed:
```go
primaryDSN := "root:your_password@tcp(localhost:3306)/testdb?parseTime=true"
```

### Option 2: Using Docker

Start MySQL using Docker:

```bash
docker run --name mysql-primary \
  -e MYSQL_ROOT_PASSWORD=password \
  -e MYSQL_DATABASE=testdb \
  -p 3306:3306 \
  -d mysql:8.0
```

Wait for MySQL to be ready:
```bash
docker exec mysql-primary mysqladmin ping -h localhost -u root -ppassword
```

## Running the Example

```bash
# From the plumego root directory
cd examples/rw

# Run the example
go run main.go
```

## Expected Output

```
=== Plumego Read/Write Splitting Example ===

1. Creating read/write cluster...
✓ Cluster created with 1 primary + 2 replicas

2. Executing WRITE operations (→ primary)...
  ✓ Inserted user (id=1)
  ✓ Updated user

3. Executing READ operations (→ replicas, round-robin)...
  ✓ Query 1: Read 1 users (distributed to replica)
  ✓ Query 2: Read 1 users (distributed to replica)
  ✓ Query 3: Read 1 users (distributed to replica)
  ✓ Query 4: Read 1 users (distributed to replica)
  ✓ Query 5: Read 1 users (distributed to replica)

4. Forcing primary for strong consistency...
  ✓ Forced primary: Found 1 users

5. Executing transaction (→ primary automatically)...
  ✓ Transaction committed

6. Executing SELECT FOR UPDATE (→ primary automatically)...
  ✓ Locked row: id=1, name=Alice Smith

7. Cluster Metrics:
  • Total queries:   13
  • Primary queries: 8
  • Replica queries: 5
  • Fallback count:  0
  • Health checks:   0

8. Replica Health Status:
  • Replica 0: ✓ healthy
  • Replica 1: ✓ healthy

9. Connection Pool Stats:
  Primary:  Open=1, InUse=0, Idle=1
  Replica 0: Open=1, InUse=0, Idle=1
  Replica 1: Open=1, InUse=0, Idle=1

=== Example Complete ===
```

## Code Walkthrough

### 1. Create Cluster

```go
cluster, err := rw.New(rw.Config{
    Primary:  primary,
    Replicas: []*sql.DB{replica1, replica2},
    HealthCheck: rw.HealthCheckConfig{
        Enabled:           true,
        Interval:          30 * time.Second,
        Timeout:           5 * time.Second,
        FailureThreshold:  3,
        RecoveryThreshold: 2,
    },
    FallbackToPrimary: true,
})
```

### 2. Write Operations

```go
// Automatically routed to primary
result, err := cluster.ExecContext(ctx,
    "INSERT INTO users (name, email) VALUES (?, ?)",
    "Alice", "alice@example.com",
)
```

### 3. Read Operations

```go
// Automatically routed to replicas with round-robin
rows, err := cluster.QueryContext(ctx, "SELECT id, name FROM users")
```

### 4. Force Primary

```go
// Override automatic routing
ctx := rw.ForcePrimary(context.Background())
rows, err := cluster.QueryContext(ctx, "SELECT * FROM users")
```

### 5. Transactions

```go
// Transactions always use primary
tx, err := cluster.BeginTx(context.Background(), nil)
tx.ExecContext(ctx, "INSERT INTO users ...")
tx.Commit()
```

## Advanced Usage

### Custom Load Balancer

```go
cluster, err := rw.New(rw.Config{
    Primary:  primary,
    Replicas: []*sql.DB{replica1, replica2},
    // Use weighted load balancer
    LoadBalancer: rw.NewWeightedBalancer([]int{1, 2}), // replica2 gets 2x traffic
})
```

### Custom Routing Policy

```go
cluster, err := rw.New(rw.Config{
    Primary:  primary,
    Replicas: []*sql.DB{replica1},
    // Always use primary (maintenance mode)
    RoutingPolicy: rw.NewAlwaysPrimaryPolicy(),
})
```

### Monitoring

```go
// Get cluster metrics
metrics := cluster.Metrics()
fmt.Printf("Primary queries: %d\n", metrics.PrimaryQueries.Load())
fmt.Printf("Replica queries: %d\n", metrics.ReplicaQueries.Load())

// Get health status
health := cluster.ReplicaHealth()
for i, isHealthy := range health {
    fmt.Printf("Replica %d: %v\n", i, isHealthy)
}

// Get connection pool stats
stats := cluster.Stats()
fmt.Printf("Primary connections: %d\n", stats.Primary.OpenConnections)
```

## Production Considerations

1. **Replication Lag**: Be aware of replication lag between primary and replicas
2. **Connection Pooling**: Configure appropriate `MaxOpenConns` and `MaxIdleConns`
3. **Health Checks**: Tune health check intervals based on your requirements
4. **Failover**: Enable `FallbackToPrimary` for high availability
5. **Monitoring**: Integrate with your monitoring system using cluster metrics

## Troubleshooting

### Connection Errors

If you see connection errors, verify:
- MySQL is running and accessible
- DSN credentials are correct
- Firewall allows connections on port 3306

### All Queries Going to Primary

This can happen if:
- No replicas are configured
- All replicas are marked unhealthy
- `FallbackToPrimary` is enabled and replicas are down

Check replica health:
```go
health := cluster.ReplicaHealth()
fmt.Println(health) // [true, true] means both healthy
```

## Related Documentation

- [Database Sharding Design](../../store/db/SHARDING_DESIGN.md)
- [API Reference](../../store/db/SHARDING_API.md)
- [Implementation Roadmap](../../store/db/ROADMAP.md)
