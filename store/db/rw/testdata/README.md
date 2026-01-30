# MySQL Primary-Replica Test Environment

This directory contains a Docker Compose configuration for testing read/write splitting with real MySQL replication.

## Architecture

```
┌─────────────────┐
│ mysql-primary   │  Port: 3306
│  (Read/Write)   │
└────────┬────────┘
         │
    ┌────┴────┐
    │         │
    ▼         ▼
┌─────────┐ ┌─────────┐
│ replica1│ │ replica2│
│  (Read) │ │  (Read) │
│  :3307  │ │  :3308  │
└─────────┘ └─────────┘
```

## Quick Start

### 1. Start the Environment

```bash
cd store/db/rw/testdata
docker-compose up -d
```

This will:
- Start MySQL primary on port **3306**
- Start MySQL replica1 on port **3307**
- Start MySQL replica2 on port **3308**
- Automatically configure GTID-based replication
- Create `testdb` database with sample data

### 2. Verify Replication

```bash
# Check all containers are running
docker-compose ps

# Check replication status on replica1
docker exec plumego-mysql-replica1 mysql -uroot -prootpassword \
  -e "SHOW SLAVE STATUS\G" | grep "Slave_.*_Running"

# Check replication status on replica2
docker exec plumego-mysql-replica2 mysql -uroot -prootpassword \
  -e "SHOW SLAVE STATUS\G" | grep "Slave_.*_Running"

# Both should show:
# Slave_IO_Running: Yes
# Slave_SQL_Running: Yes
```

### 3. Test Replication

```bash
# Insert data on primary
docker exec plumego-mysql-primary mysql -utestuser -ptestpass testdb \
  -e "INSERT INTO users (name, email) VALUES ('Test', 'test@example.com')"

# Verify data replicated to replica1
docker exec plumego-mysql-replica1 mysql -utestuser -ptestpass testdb \
  -e "SELECT * FROM users WHERE email='test@example.com'"

# Verify data replicated to replica2
docker exec plumego-mysql-replica2 mysql -utestuser -ptestpass testdb \
  -e "SELECT * FROM users WHERE email='test@example.com'"
```

## Connection Strings

### Primary (Read/Write)
```
testuser:testpass@tcp(localhost:3306)/testdb
root:rootpassword@tcp(localhost:3306)/testdb
```

### Replica 1 (Read Only)
```
testuser:testpass@tcp(localhost:3307)/testdb
root:rootpassword@tcp(localhost:3307)/testdb
```

### Replica 2 (Read Only)
```
testuser:testpass@tcp(localhost:3308)/testdb
root:rootpassword@tcp(localhost:3308)/testdb
```

## Running Integration Tests

```bash
# From plumego root directory
cd store/db/rw

# Run integration tests (requires docker environment)
go test -v -tags=integration -timeout 60s

# Or run specific integration test
go test -v -run TestIntegrationRealMySQL -tags=integration
```

## Monitoring Replication

### Check Replication Lag

```bash
# On replica1
docker exec plumego-mysql-replica1 mysql -uroot -prootpassword \
  -e "SHOW SLAVE STATUS\G" | grep "Seconds_Behind_Master"

# On replica2
docker exec plumego-mysql-replica2 mysql -uroot -prootpassword \
  -e "SHOW SLAVE STATUS\G" | grep "Seconds_Behind_Master"

# Should show: Seconds_Behind_Master: 0
```

### View Binary Log Position

```bash
# Primary position
docker exec plumego-mysql-primary mysql -uroot -prootpassword \
  -e "SHOW MASTER STATUS"

# Replica positions
docker exec plumego-mysql-replica1 mysql -uroot -prootpassword \
  -e "SHOW SLAVE STATUS\G" | grep "Exec_Master_Log_Pos"
```

## Troubleshooting

### Replication Not Working

1. Check container logs:
```bash
docker-compose logs mysql-primary
docker-compose logs mysql-replica1
docker-compose logs replication-setup
```

2. Check replication errors:
```bash
docker exec plumego-mysql-replica1 mysql -uroot -prootpassword \
  -e "SHOW SLAVE STATUS\G" | grep "Last_.*_Error"
```

3. Restart replication:
```bash
docker exec plumego-mysql-replica1 mysql -uroot -prootpassword \
  -e "STOP SLAVE; START SLAVE;"
```

### Reset Environment

```bash
# Stop and remove containers + volumes
docker-compose down -v

# Start fresh
docker-compose up -d
```

### Port Already in Use

If ports 3306, 3307, or 3308 are already in use, you can change them in `docker-compose.yml`:

```yaml
ports:
  - "13306:3306"  # Use different host port
```

## Cleanup

```bash
# Stop containers
docker-compose stop

# Remove containers
docker-compose down

# Remove containers and volumes (deletes all data!)
docker-compose down -v
```

## Configuration Details

### MySQL Configuration

- **GTID Mode**: Enabled for automatic failover support
- **Binary Logging**: ROW format for complete replication
- **Server IDs**: Primary=1, Replica1=2, Replica2=3
- **Character Set**: utf8mb4
- **Storage Engine**: InnoDB

### Replication User

- Username: `repl`
- Password: `replpass`
- Privileges: `REPLICATION SLAVE`

### Test Database

- Database: `testdb`
- User: `testuser`
- Password: `testpass`
- Tables: `users` (with sample data)

## Performance Testing

This environment can be used for performance testing:

```bash
# Run concurrent reads from replicas
go test -bench=BenchmarkCluster -benchtime=10s

# Run load test
go test -run=TestConcurrentQueries -v
```

## Security Notes

⚠️ **WARNING**: This configuration is for testing only!

- Uses weak passwords
- All services accessible from host
- No SSL/TLS encryption
- Root user enabled with password

**Never use these settings in production!**

## Related Documentation

- [Read/Write Splitting Design](../SHARDING_DESIGN.md)
- [Example Application](../../examples/rw/)
- [Test Suite](../cluster_test.go)
