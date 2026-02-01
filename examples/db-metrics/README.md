# Database Metrics Example

This example demonstrates comprehensive database metrics tracking using plumego's instrumented database wrapper.

## Features

- ✅ Automatic query tracking (SELECT, INSERT, UPDATE, DELETE)
- ✅ Transaction monitoring
- ✅ Connection pool health metrics
- ✅ Query duration tracking with percentiles
- ✅ Error rate monitoring
- ✅ Prometheus metrics exposition
- ✅ Real-time stats dashboard

## Quick Start

```bash
# Run the example
cd examples/db-metrics
go run main.go
```

The server will start on http://localhost:8080

## Endpoints

- **http://localhost:8080/** - Web interface with documentation
- **http://localhost:8080/metrics** - Prometheus metrics endpoint
- **http://localhost:8080/stats** - Real-time connection pool statistics

## Simulated Workload

The application continuously performs:

1. **User Creation** - Creates new users every 2-5 seconds
2. **Post Creation** - Creates 2-3 posts for each user
3. **Query Operations** - Performs various SELECT queries
4. **Transactions** - Runs UPDATE transactions every 10 users
5. **Pool Monitoring** - Records pool stats every 10 seconds

## Metrics Available

### Operation Metrics

```
db_query_total{operation="query",driver="sqlite3"}
db_query_total{operation="exec",driver="sqlite3"}
db_query_total{operation="transaction",driver="sqlite3"}

db_query_duration_milliseconds{operation="query",driver="sqlite3"}
db_query_errors_total{operation="query",driver="sqlite3"}
```

### Pool Metrics

```
db_pool_open_connections{driver="sqlite3"}
db_pool_in_use{driver="sqlite3"}
db_pool_idle{driver="sqlite3"}
db_pool_wait_count{driver="sqlite3"}
db_pool_wait_duration{driver="sqlite3"}
db_pool_max_idle_closed{driver="sqlite3"}
db_pool_max_lifetime_closed{driver="sqlite3"}
```

## Prometheus Integration

### Using Prometheus

1. Create `prometheus.yml`:

```yaml
global:
  scrape_interval: 15s

scrape_configs:
  - job_name: 'dbmetrics_example'
    static_configs:
      - targets: ['localhost:8080']
```

2. Run Prometheus:

```bash
prometheus --config.file=prometheus.yml
```

3. Access Prometheus UI: http://localhost:9090

### Sample Queries

```promql
# Queries per second
rate(db_query_total[1m])

# Average query duration (ms)
rate(db_query_duration_milliseconds_sum[5m]) /
  rate(db_query_duration_milliseconds_count[5m])

# 95th percentile query latency
histogram_quantile(0.95,
  rate(db_query_duration_milliseconds_bucket[5m]))

# Error rate percentage
(rate(db_query_errors_total[5m]) /
  rate(db_query_total[5m])) * 100

# Connection pool utilization
(db_pool_in_use / db_pool_open_connections) * 100

# Connection wait rate
rate(db_pool_wait_count[5m])
```

## Grafana Dashboard

Create a Grafana dashboard with panels for:

1. **Query Rate** - Total queries per second by operation type
2. **Query Latency** - p50, p95, p99 latencies over time
3. **Error Rate** - Error percentage over time
4. **Pool Health** - Open connections, in-use, idle
5. **Wait Times** - Connection wait count and duration

Example panel configuration:

```json
{
  "title": "Query Rate",
  "targets": [
    {
      "expr": "rate(db_query_total[1m])",
      "legendFormat": "{{operation}}"
    }
  ]
}
```

## Alerting

Set up alerts for:

```promql
# High error rate (> 1%)
(rate(db_query_errors_total[5m]) / rate(db_query_total[5m])) > 0.01

# Slow queries (p95 > 1 second)
histogram_quantile(0.95, rate(db_query_duration_milliseconds_bucket[5m])) > 1000

# Pool exhaustion (> 90% utilization)
(db_pool_in_use / db_pool_open_connections) > 0.9

# Connection waits occurring
rate(db_pool_wait_count[5m]) > 0

# High connection churn
rate(db_pool_max_lifetime_closed[5m]) > 10
```

## Production Configuration

For production use, configure the connection pool:

```go
db.SetMaxOpenConns(25)           // Maximum open connections
db.SetMaxIdleConns(5)            // Maximum idle connections
db.SetConnMaxLifetime(5*time.Minute)  // Connection max lifetime
db.SetConnMaxIdleTime(2*time.Minute)  // Idle connection timeout
```

## Testing

Run with different workloads:

```bash
# Simulate high load
wrk -t12 -c400 -d30s http://localhost:8080/stats

# Simulate slow queries (add SLEEP in queries)
# Monitor metrics to see impact
```

## Learn More

- [Plumego Metrics Package](../../metrics/README.md)
- [Database Metrics Documentation](../../store/db/README_METRICS.md)
- [Prometheus Documentation](https://prometheus.io/docs/)
- [Grafana Dashboards](https://grafana.com/docs/)
