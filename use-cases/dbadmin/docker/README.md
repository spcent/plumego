# Docker Development Environment

This directory contains configuration for running local development instances of all supported databases.

## Quick Start

From the project root (`use-cases/dbadmin/`):

```bash
# Start all databases
docker compose -f docker-compose.dev.yml up -d

# Check status
docker compose -f docker-compose.dev.yml ps

# View logs
docker compose -f docker-compose.dev.yml logs -f

# Stop all databases
docker compose -f docker-compose.dev.yml down

# Stop and delete all data (fresh start)
docker compose -f docker-compose.dev.yml down -v
```

## Database Connections

Use these credentials when creating connections in the dbadmin UI:

### MySQL
- **Host:** `localhost`
- **Port:** `13306`
- **Username:** `dbadmin`
- **Password:** `dbadminpass`
- **Database:** `testdb`
- **Root password:** `rootpassword`

### Redis
- **Host:** `localhost`
- **Port:** `16379`
- **Password:** `redispass`

### MongoDB
- **URI:** `mongodb://dbadmin:dbadminpass@localhost:37017/testdb?authSource=admin`
- **Or individual fields:**
  - **Host:** `localhost`
  - **Port:** `37017`
  - **Username:** `dbadmin`
  - **Password:** `dbadminpass`
  - **Database:** `testdb`
  - **Auth DB:** `admin`

### Elasticsearch
- **Host:** `localhost`
- **Port:** `19200`
- **Protocol:** `http` (no SSL in dev)

### SQLite
SQLite doesn't need Docker. Just point to a local `.db` file. For testing, create one:

```bash
sqlite3 ./data/test.db
# Then run some CREATE TABLE / INSERT statements
```

## Seed Data

The MySQL and MongoDB containers automatically load seed data on first startup:

- **MySQL:** `docker/mysql/init/01-seed.sql` creates `users`, `products`, `orders`, `order_items` tables with sample data, plus a view and stored procedure
- **MongoDB:** `docker/mongodb/init/01-seed.js` creates `users`, `products`, `orders`, `datatypes` collections with sample documents

Redis and Elasticsearch are intentionally started empty. To seed quick demo
records for those services, use the commands in
[`docs/demo-playbook.md`](../docs/demo-playbook.md#seed-optional-redis-and-elasticsearch-data).

To re-seed:
```bash
# Stop and remove volumes
docker compose -f docker-compose.dev.yml down -v

# Start again (seed scripts run on first init)
docker compose -f docker-compose.dev.yml up -d
```

## Health Checks

All services have health checks configured. Wait for them to pass before connecting:

```bash
# Wait for all services to be healthy
docker compose -f docker-compose.dev.yml up -d --wait

# Or check individual service health
docker inspect --format='{{.State.Health.Status}}' dbadmin-mysql
docker inspect --format='{{.State.Health.Status}}' dbadmin-redis
docker inspect --format='{{.State.Health.Status}}' dbadmin-mongodb
docker inspect --format='{{.State.Health.Status}}' dbadmin-elasticsearch
```

## Troubleshooting

### Port conflicts
If any port is already in use, edit `docker-compose.dev.yml` and change the host port (left side of the mapping):
```yaml
ports:
  - "YOUR_PORT:3306"  # Change YOUR_PORT
```

### Elasticsearch won't start
Elasticsearch requires sufficient memory. If it fails:
1. Increase Docker memory allocation (Docker Desktop → Settings → Resources)
2. Reduce JVM heap in `docker-compose.dev.yml`: `ES_JAVA_OPTS=-Xms256m -Xmx256m`

### MongoDB auth fails
Ensure you're using `authSource=admin` in the connection URI, or use the individual field form.

### Data persistence
Data is stored in Docker volumes:
- `mysql_data`
- `redis_data`
- `mongodb_data`
- `elasticsearch_data`

To list volumes:
```bash
docker volume ls | grep dbadmin
```

To delete a specific volume:
```bash
docker volume rm dbadmin_mysql_data
```

## Resource Usage

Approximate memory usage when all services are running:
- MySQL: ~400MB
- Redis: ~50MB
- MongoDB: ~200MB
- Elasticsearch: ~1GB (configured for 512MB heap)

**Total:** ~1.7GB RAM

If this is too much for your machine, start only the databases you need:
```bash
# Start only MySQL and Redis
docker compose -f docker-compose.dev.yml up -d mysql redis

# Start only MongoDB
docker compose -f docker-compose.dev.yml up -d mongodb
```

## Custom Configuration

To customize database settings, create override files:

**docker/mysql/custom.cnf:**
```ini
[mysqld]
max_connections=500
innodb_buffer_pool_size=256M
```

Then mount it in `docker-compose.dev.yml`:
```yaml
volumes:
  - ./docker/mysql/custom.cnf:/etc/mysql/conf.d/custom.cnf
```

## CI/CD Usage

For CI pipelines, use the same compose file with a different project name to avoid conflicts:

```bash
docker compose -f docker-compose.dev.yml -p dbadmin-ci up -d --wait
# Run tests
docker compose -f docker-compose.dev.yml -p dbadmin-ci down -v
```

For release readiness, prefer the scripted smoke test from `use-cases/dbadmin/`:

```bash
make release-smoke
```

The smoke test starts the datasource stack, builds the frontend, runs dbadmin
with an explicit password and encryption key, logs in, creates MySQL, Redis,
MongoDB, and Elasticsearch connections, and verifies each connection.
