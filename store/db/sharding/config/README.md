# Database Sharding Configuration

This package provides configuration management for the database sharding system, including JSON/YAML file loading, environment variable support, and hot reload functionality.

## Features

- **Multiple Format Support**: Load configuration from JSON or YAML files
- **Environment Variables**: Override configuration with environment variables
- **Hot Reload**: Watch configuration files for changes and reload automatically
- **Validation**: Comprehensive validation of all configuration fields
- **DSN Building**: Automatic DSN construction for MySQL, PostgreSQL, and SQLite

## Configuration File Formats

### YAML Format (Recommended)

```yaml
# Database shards
shards:
  - name: shard0
    primary:
      driver: mysql
      host: db0.example.com
      port: 3306
      database: app_shard0
      username: app_user
      password: secret
      max_open_conns: 100
      max_idle_conns: 10
      conn_max_lifetime: 30m
      conn_max_idle_time: 5m
    replicas:
      - driver: mysql
        host: db0-replica.example.com
        port: 3306
        database: app_shard0
        username: app_user
        password: secret
    replica_weights: [1]
    fallback_to_primary: true
    health_check:
      enabled: true
      interval: 30s
      timeout: 5s
      failure_threshold: 3
      recovery_threshold: 2

  - name: shard1
    primary:
      driver: mysql
      host: db1.example.com
      port: 3306
      database: app_shard1
      username: app_user
      password: secret

# Sharding rules
sharding_rules:
  - table_name: users
    shard_key_column: user_id
    strategy: mod
    actual_table_names:
      0: users_0
      1: users_1
    default_shard: -1

  - table_name: orders
    shard_key_column: user_id
    strategy: hash
    actual_table_names:
      0: orders_0
      1: orders_1

  - table_name: events
    shard_key_column: created_at
    strategy: range
    strategy_config:
      ranges:
        - start: 0
          end: 1000000
          shard: 0
        - start: 1000000
          end: 2000000
          shard: 1

  - table_name: regions
    shard_key_column: region_code
    strategy: list
    strategy_config:
      mapping:
        US: 0
        EU: 1
      default_shard: 0

# Global settings
cross_shard_policy: deny  # deny, first, or all
default_shard_index: -1   # -1 to disable
enable_metrics: true
enable_tracing: false
log_level: info  # debug, info, warn, error
```

### JSON Format

```json
{
  "shards": [
    {
      "name": "shard0",
      "primary": {
        "driver": "mysql",
        "host": "db0.example.com",
        "port": 3306,
        "database": "app_shard0",
        "username": "app_user",
        "password": "secret"
      }
    }
  ],
  "sharding_rules": [
    {
      "table_name": "users",
      "shard_key_column": "user_id",
      "strategy": "mod",
      "actual_table_names": {
        "0": "users_0",
        "1": "users_1"
      }
    }
  ],
  "cross_shard_policy": "deny",
  "log_level": "info"
}
```

## Configuration Reference

### Top-Level Configuration

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `shards` | array | Yes | - | List of database shards |
| `sharding_rules` | array | Yes | - | List of sharding rules for tables |
| `cross_shard_policy` | string | No | `deny` | Policy for cross-shard queries: `deny`, `first`, or `all` |
| `default_shard_index` | int | No | `-1` | Default shard when routing fails (-1 to disable) |
| `enable_metrics` | bool | No | `true` | Enable Prometheus metrics collection |
| `enable_tracing` | bool | No | `false` | Enable OpenTelemetry distributed tracing |
| `log_level` | string | No | `info` | Logging level: `debug`, `info`, `warn`, or `error` |

### Shard Configuration

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `name` | string | Yes | - | Unique shard identifier |
| `primary` | object | Yes | - | Primary database configuration |
| `replicas` | array | No | `[]` | Read replica configurations |
| `replica_weights` | array | No | `[]` | Load balancing weights for replicas |
| `fallback_to_primary` | bool | No | `false` | Use primary when all replicas are down |
| `health_check` | object | No | - | Health check configuration |

### Database Configuration

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `driver` | string | Yes* | - | Database driver: `mysql`, `postgres`, or `sqlite3` |
| `host` | string | Yes* | - | Database host |
| `port` | int | No | Driver default | Database port (3306 for MySQL, 5432 for PostgreSQL) |
| `database` | string | Yes | - | Database name or path (for SQLite) |
| `username` | string | No | - | Database username |
| `password` | string | No | - | Database password |
| `dsn` | string | No | - | Full DSN (overrides other fields if set) |
| `max_open_conns` | int | No | `0` | Maximum number of open connections (0 = unlimited) |
| `max_idle_conns` | int | No | `2` | Maximum number of idle connections |
| `conn_max_lifetime` | string | No | `0` | Maximum connection lifetime (e.g., `30m`) |
| `conn_max_idle_time` | string | No | `0` | Maximum connection idle time (e.g., `5m`) |

*Not required if `dsn` is provided

### Health Check Configuration

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `enabled` | bool | No | `false` | Enable health checking |
| `interval` | string | No | `30s` | Health check interval |
| `timeout` | string | No | `5s` | Health check timeout |
| `failure_threshold` | int | No | `3` | Failures before marking unhealthy |
| `recovery_threshold` | int | No | `2` | Successes before marking healthy |

### Sharding Rule Configuration

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `table_name` | string | Yes | - | Logical table name |
| `shard_key_column` | string | Yes | - | Column name for sharding |
| `strategy` | string | Yes | - | Sharding strategy: `hash`, `mod`, `range`, or `list` |
| `strategy_config` | object | No | - | Strategy-specific configuration |
| `actual_table_names` | map | No | - | Mapping of shard index to physical table name |
| `default_shard` | int | No | `-1` | Default shard for this table (-1 to disable) |

### Strategy Configuration

For **range** strategy:
```yaml
strategy_config:
  ranges:
    - start: 0
      end: 1000
      shard: 0
    - start: 1000
      end: 2000
      shard: 1
```

For **list** strategy:
```yaml
strategy_config:
  mapping:
    US: 0
    EU: 1
    ASIA: 2
  default_shard: 0
```

## Environment Variables

Configuration values can be overridden with environment variables:

| Variable | Type | Description |
|----------|------|-------------|
| `DB_SHARD_CROSS_SHARD_POLICY` | string | Cross-shard query policy |
| `DB_SHARD_DEFAULT_INDEX` | int | Default shard index |
| `DB_SHARD_ENABLE_METRICS` | bool | Enable metrics (true/false, 1/0, yes/no) |
| `DB_SHARD_ENABLE_TRACING` | bool | Enable tracing |
| `DB_SHARD_LOG_LEVEL` | string | Log level |

## Usage Examples

### Basic Usage

```go
package main

import (
    "github.com/spcent/plumego/store/db/sharding/config"
)

func main() {
    // Load configuration from file
    cfg, err := config.LoadFromFile("sharding.yaml")
    if err != nil {
        panic(err)
    }

    // Merge with environment variables
    if err := cfg.MergeWithEnv(); err != nil {
        panic(err)
    }

    // Use configuration...
}
```

### Hot Reload

```go
package main

import (
    "context"
    "log"
    "time"

    "github.com/spcent/plumego/store/db/sharding/config"
)

func main() {
    // Create config watcher
    watcher, err := config.NewConfigWatcher(
        "sharding.yaml",
        config.WithWatchInterval(5*time.Second),
        config.WithOnChange(func(cfg *config.ShardingConfig) {
            log.Printf("Configuration reloaded: %d shards", len(cfg.Shards))
        }),
    )
    if err != nil {
        panic(err)
    }

    // Start watching
    ctx := context.Background()
    go watcher.Start(ctx)
    defer watcher.Stop()

    // Get current configuration
    cfg := watcher.Get()
    log.Printf("Initial config: %d shards", len(cfg.Shards))

    // Keep running...
    select {}
}
```

### Using ConfigReloader

```go
package main

import (
    "context"
    "log"
    "time"

    "github.com/spcent/plumego/store/db/sharding/config"
)

func main() {
    // Create reloader
    reloader, err := config.NewConfigReloader(
        "sharding.yaml",
        config.WithWatchInterval(10*time.Second),
    )
    if err != nil {
        panic(err)
    }

    // Register change callback
    reloader.OnChange(func(cfg *config.ShardingConfig) {
        log.Printf("Configuration changed: %d shards", len(cfg.Shards))
        // Update your sharding cluster...
    })

    // Start reloader
    ctx := context.Background()
    go reloader.Start(ctx)
    defer reloader.Stop()

    // Get configuration
    cfg := reloader.Get()
    log.Printf("Loaded: %d shards, %d rules",
        len(cfg.Shards), len(cfg.ShardingRules))

    // Manually reload if needed
    if err := reloader.Reload(); err != nil {
        log.Printf("Manual reload failed: %v", err)
    }

    // Keep running...
    select {}
}
```

### Loading Different Formats

```go
// Auto-detect format (JSON or YAML)
cfg, err := config.LoadFromFile("config.yaml")

// Explicitly load JSON
cfg, err := config.LoadFromJSONFile("config.json")

// Explicitly load YAML
cfg, err := config.LoadFromYAMLFile("config.yaml")

// Load from bytes
data := []byte(`{"shards": [...]}`)
cfg, err := config.LoadFromJSON(data)
```

## DSN Building

The package automatically builds DSNs for different database drivers:

### MySQL
```yaml
driver: mysql
host: localhost
port: 3306
database: mydb
username: user
password: pass
```
Generates: `user:pass@tcp(localhost:3306)/mydb`

### PostgreSQL
```yaml
driver: postgres
host: localhost
port: 5432
database: mydb
username: user
password: pass
```
Generates: `host=localhost port=5432 dbname=mydb user=user password=pass sslmode=disable`

### SQLite
```yaml
driver: sqlite3
database: /path/to/db.sqlite
```
Generates: `/path/to/db.sqlite`

### Custom DSN
```yaml
dsn: "custom://connection/string"
```
Uses the provided DSN directly, ignoring other fields.

## Validation

All configuration is validated on load:

- At least one shard required
- At least one sharding rule required
- Valid cross-shard policy (`deny`, `first`, `all`)
- Valid log level (`debug`, `info`, `warn`, `error`)
- Valid database driver
- Required fields present
- Replica weights match replica count
- Range strategy has range definitions

Validation errors are returned with context about which field failed.

## Best Practices

1. **Use YAML for readability**: YAML is easier to read and maintain than JSON
2. **Version control your config**: Keep configuration files in version control
3. **Environment-specific configs**: Use different files for dev/staging/prod
4. **Environment variables for secrets**: Override passwords and secrets via env vars
5. **Enable health checks**: Always enable health checks in production
6. **Set connection limits**: Configure `max_open_conns` and `max_idle_conns`
7. **Use hot reload in production**: Enable automatic configuration reloading
8. **Monitor metrics**: Enable metrics and integrate with Prometheus

## Examples

See the `examples/` directory for complete configuration examples:
- `cluster.yaml` - Full production configuration
- `cluster-simple.yaml` - Minimal configuration
- `cluster.json` - JSON format example
