# dbadmin

A local-first, security-focused database management workbench for developers. Manage MySQL, PostgreSQL, SQLite, Redis, MongoDB, and Elasticsearch from a single modern web interface.

## Overview

dbadmin is a self-hosted database workbench designed for local development and internal tools. Unlike cloud-based solutions, dbadmin runs entirely on your machine, keeping your data and credentials private. It provides a unified interface for multiple database types with built-in security features to prevent accidental data loss.

As a Plumego use-case app, dbadmin demonstrates how to build a production-scale internal tool with explicit routing, session authentication, KV-backed state, connection pooling, embedded frontend assets, health checks, and backend-enforced safety controls.

**Who is this for?**
- Developers who need to inspect and modify data across multiple databases
- Teams building internal tools that need database access
- Anyone who wants a lightweight, local alternative to heavy database clients

## Supported Data Sources

| Database | Features | Status |
|----------|----------|--------|
| **MySQL** | Browse tables, execute queries, view/edit rows, DDL operations (incl. views), bulk row update/delete, import/export (CSV/JSON/SQL/XLSX) | ✅ Stable |
| **PostgreSQL** | Browse schemas/tables, execute queries, view/edit rows, DDL (CREATE/ALTER/DROP TABLE, views), bulk row update/delete, import/export (CSV/JSON/SQL/XLSX) | ✅ Stable |
| **SQLite** | All MySQL features + file upload/download, schema inspection | ✅ Stable |
| **Redis** | Browse keys by type, execute commands, command history, TTL management, batch operations, standalone/Cluster/Sentinel modes | ✅ Stable |
| **MongoDB** | Browse collections, query with filters, aggregation pipelines, explain plans, collection and index management | ✅ Stable |
| **Elasticsearch** | Browse indices, DSL queries, document CRUD, import/export, cluster health, mappings, index and index-template management | ✅ Stable |

## Features

### Core Capabilities
- **Connection Management**: Save, test, and organize connections with optional credential encryption
- **Connection Config Import/Export**: Export all saved connections as a JSON document (with secrets redacted) and re-import them on another instance
- **Resource Browsing**: Tree view of databases, tables, collections, indices, and keys
- **Query Execution**: Syntax-highlighted editors with server-side SQL, Redis, MongoDB, and Elasticsearch history plus result formatting
- **Data Manipulation**: View, edit, insert, delete with confirmation dialogs for dangerous operations
- **Import/Export**: CSV, JSON, SQL dump, and XLSX formats for data migration
- **Bulk Row Operations**: Update or delete multiple SQL rows by primary key in a single confirmed request
- **SQL View Management**: Create and drop views (`CREATE VIEW`/`DROP VIEW`) per connection
- **Schema Inspection**: View table structure, indexes, foreign keys, mappings
- **MongoDB Collection/Index Management**: Create and drop collections and indexes (compound, unique, sparse) per connection
- **Elasticsearch Index/Template Management**: Create and delete indices and index templates, with configurable settings and mappings
- **Redis Cluster & Sentinel**: Configure a connection for standalone, Cluster, or Sentinel-backed failover topologies
- **Operational Diagnostics**: Health, readiness, runtime, and connection-pool endpoints for local troubleshooting
- **Prometheus Metrics**: `/metrics` endpoint exposing HTTP request counts, latency histograms, and uptime in Prometheus text exposition format
- **Admin Controls**: Active operation cancellation, connection runtime closing, audit events, and multi-user `admin`/`readonly` roles

### Security & Safety
- **Read-only Mode**: Block all write operations per connection
- **Dangerous Operation Detection**: Intercept DROP, DELETE, FLUSH, and other destructive commands
- **Confirmation Dialogs**: Require explicit confirmation before executing dangerous operations
- **Credential Redaction**: Passwords and API keys redacted in all logs and API responses
- **Session Management**: HttpOnly cookie-based authentication with automatic timeout, login rate limiting, same-origin checks for unsafe requests, and self-service bulk session revocation
- **Multi-user Accounts**: Configure multiple named users with independent `admin`/`readonly` roles via `DBADMIN_USERS`
- **IP Allowlisting**: Restrict the login endpoint and all protected routes to a configured set of CIDRs/IPs via `DBADMIN_ALLOWED_IPS`
- **Input Validation**: SQL injection prevention and command injection protection
- **Bounded Execution**: Configurable SQL, Redis, MongoDB, Elasticsearch, and resource-list timeouts
- **SQL Query Cancellation**: Cancel active SQL queries from the UI when cancellation is enabled

### User Experience
- **Modern UI**: Responsive design with Tailwind CSS
- **Dark Mode**: Automatic theme switching based on system preferences
- **Internationalization**: English and Chinese (简体中文) support
- **Query History**: Searchable history with execution time and result counts, filterable by query text, error status, and database
- **Keyboard Shortcuts**: Efficient navigation and execution (Ctrl/Cmd+Enter)

## Quick Start

### Prerequisites
- Go 1.26 or later
- Node.js 24 or later
- npm

### Installation

```bash
# Clone the repository
git clone https://github.com/spcent/plumego.git
cd plumego/use-cases/dbadmin

# Backend setup
go mod download

# Frontend setup
cd web
npm install
npm run build
cd ..

# Start the server
DBADMIN_PASSWORD=your-secure-password go run main.go
```

Visit `http://localhost:8080`

**Default user:** `admin`

⚠️ **Security Notice**: dbadmin refuses to start without `DBADMIN_PASSWORD`. The demo-only `admin/admin` credential requires `DBADMIN_ALLOW_DEFAULT_PASSWORD=true` and should never be used for real deployments.

## Development

### Running in Development Mode

Start both backend and frontend with hot reload:

```bash
# Terminal 1: Backend
go run main.go

# Terminal 2: Frontend (with hot reload)
cd web
npm run dev
```

The frontend dev server runs on `http://localhost:5173` and proxies API requests to the backend.

### Building for Production

Build a single binary with embedded frontend:

```bash
cd web
npm run build
cd ..

go build -o dbadmin main.go
```

Run the binary:
```bash
./dbadmin
```

The binary includes all frontend assets and can be deployed without Node.js.

### Using Docker

```bash
# From the repository root. The dbadmin module uses a local replace directive.
docker build -f use-cases/dbadmin/Dockerfile -t dbadmin .

# Run container
docker run -p 8080:8080 -v ./data:/app/data \
  -e DBADMIN_PASSWORD=your-secure-password \
  -e DBADMIN_ENCRYPTION_KEY=$(openssl rand -hex 32) \
  dbadmin
```

For development with hot reload, use Docker Compose:

```bash
docker-compose up
```

For a guided multi-database demo, see [docs/demo-playbook.md](docs/demo-playbook.md).

## Configuration

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `APP_ADDR` | `127.0.0.1:8080` | Server listen address |
| `DBADMIN_DATA_DIR` | `./data` | Directory for sessions, history, and uploads |
| `DBADMIN_USER` | `admin` | Admin username (single-user mode; ignored when `DBADMIN_USERS` is set) |
| `DBADMIN_PASSWORD` | required | Admin password (single-user mode; ignored when `DBADMIN_USERS` is set) |
| `DBADMIN_ROLE` | `admin` | Single-user role: `admin` or `readonly` |
| `DBADMIN_USERS` | unset | JSON array for multi-user mode: `[{"username":"alice","password":"...","role":"admin"}]`. Overrides `DBADMIN_USER`/`DBADMIN_PASSWORD`/`DBADMIN_ROLE` when set |
| `DBADMIN_ALLOWED_IPS` | unset | Comma-separated CIDRs/IPs allowed to reach the login endpoint and all protected routes, e.g. `10.0.0.0/8,192.168.1.50`. Empty allows all |
| `DBADMIN_ALLOW_DEFAULT_PASSWORD` | `false` | Demo-only override that permits `admin/admin` |
| `DBADMIN_SESSION_TTL` | `24h` | Session timeout duration |
| `DBADMIN_ENCRYPTION_KEY` | unset | 32-byte hex key for credential encryption; required before saving passwords, API keys, or Mongo URIs with embedded credentials |
| `DBADMIN_QUERY_TIMEOUT_SECONDS` | `30` | Maximum SQL query execution time in seconds |
| `DBADMIN_QUERY_CANCEL_ENABLED` | `true` | Enable SQL query and non-SQL operation cancellation endpoints and UI controls |
| `DBADMIN_REDIS_COMMAND_TIMEOUT_SECONDS` | `30` | Maximum Redis command/listing time in seconds |
| `DBADMIN_MONGO_QUERY_TIMEOUT_SECONDS` | `30` | Maximum MongoDB operation time in seconds |
| `DBADMIN_ES_QUERY_TIMEOUT_SECONDS` | `30` | Maximum Elasticsearch operation time in seconds |
| `DBADMIN_RESOURCE_LIST_TIMEOUT_SECONDS` | `30` | Maximum unified resource-tree listing time in seconds |
| `DBADMIN_AUDIT_RETENTION_DAYS` | `90` | Local audit event retention period in days |
| `DBADMIN_AUDIT_MAX_EVENTS` | `10000` | Maximum local audit events retained |

### Configuration File

Create a `.env` file in the project root:

```bash
APP_ADDR=127.0.0.1:8080
DBADMIN_DATA_DIR=./data
DBADMIN_USER=admin
DBADMIN_PASSWORD=your-secure-password
DBADMIN_ROLE=admin
DBADMIN_ALLOW_DEFAULT_PASSWORD=false
DBADMIN_SESSION_TTL=24h
DBADMIN_ENCRYPTION_KEY=your-32-byte-hex-key
DBADMIN_QUERY_TIMEOUT_SECONDS=30
DBADMIN_QUERY_CANCEL_ENABLED=true
DBADMIN_REDIS_COMMAND_TIMEOUT_SECONDS=30
DBADMIN_MONGO_QUERY_TIMEOUT_SECONDS=30
DBADMIN_ES_QUERY_TIMEOUT_SECONDS=30
DBADMIN_RESOURCE_LIST_TIMEOUT_SECONDS=30
DBADMIN_AUDIT_RETENTION_DAYS=90
DBADMIN_AUDIT_MAX_EVENTS=10000
# Optional: multi-user mode (overrides DBADMIN_USER/PASSWORD/ROLE above)
# DBADMIN_USERS=[{"username":"alice","password":"alice-secret","role":"admin"},{"username":"bob","password":"bob-secret","role":"readonly"}]
# Optional: restrict access to a set of CIDRs/IPs
# DBADMIN_ALLOWED_IPS=10.0.0.0/8,192.168.1.50
```

Generate a secure encryption key:
```bash
openssl rand -hex 32
```

### Command-line Flags

```bash
./dbadmin --addr=:9090 --data-dir=/var/dbadmin
```

## Architecture

```
use-cases/dbadmin/
├── main.go             # Application entry point
├── internal/
│   ├── app/            # HTTP routes and middleware
│   ├── config/         # Configuration management
│   ├── domain/         # Business logic
│   │   ├── connection/ # Connection CRUD and encryption
│   │   ├── session/    # Session management
│   │   ├── history/    # Query history (SQL)
│   │   ├── redishistory/ # Command history (Redis)
│   │   ├── mongohistory/ # Query history (MongoDB)
│   │   └── eshistory/  # Query history (Elasticsearch)
│   ├── handler/        # HTTP handlers
│   │   ├── sql.go      # MySQL/SQLite operations
│   │   ├── redis.go    # Redis operations
│   │   ├── mongodb.go  # MongoDB operations
│   │   └── elasticsearch.go # Elasticsearch operations
│   ├── dbmanager/      # SQL connection pooling
│   ├── redismanager/   # Redis connection management
│   ├── mongomanager/   # MongoDB connection management
│   ├── esmanager/      # Elasticsearch connection management
│   ├── datasource/     # Unified resource tree adapters
│   └── retry/          # Retry helpers
├── web/                # React frontend
│   ├── src/
│   │   ├── components/ # Reusable UI components
│   │   ├── pages/      # Page components
│   │   ├── api.ts      # API client
│   │   └── hooks/      # Custom React hooks
│   └── public/         # Static assets
├── docker/             # Local datasource seed data and docs
└── docs/               # Documentation
```

## Testing

### Backend Tests

Run all backend tests:
```bash
go test ./...
```

Run with coverage:
```bash
go test -cover ./...
```

Run specific test category:
```bash
go test ./internal/handler -run TestSQLQuote
go test ./internal/handler -run TestRedisCommand
go test ./internal/handler -run TestMongoFilter
go test ./internal/handler -run TestESValidation
```

### Frontend Tests

Run all frontend tests:
```bash
cd web
npm test
```

Run with coverage:
```bash
npm run test:coverage
```

Run in watch mode:
```bash
npm run test:watch
```

### Manual Testing

See [docs/manual-test.md](docs/manual-test.md) for comprehensive manual testing procedures covering:
- Connection management
- Data source operations
- Security features
- UI/UX validation
- Performance testing

## Security Considerations

### Local Development
dbadmin is designed for local development and internal use. By default, it binds to `localhost:8080` and should not be exposed to the public internet.

### Production Deployment
If deploying to production:
1. **Set credentials explicitly** - Set strong `DBADMIN_USER` and `DBADMIN_PASSWORD`
2. **Use HTTPS** - Configure TLS certificates or use a reverse proxy with SSL
3. **Set encryption key** - Generate a secure `DBADMIN_ENCRYPTION_KEY` before saving connection credentials
4. **Restrict network access** - Use firewall rules or VPN to limit access
5. **Enable read-only mode** - For connections that don't need write access
6. **Review audit events** - Export JSON or NDJSON audit events from Settings before pruning old data
7. **Regular backups** - Backup the `data` directory containing sessions, history, and audit data

### Security Features
- **Credential Encryption**: Connection passwords, Elasticsearch secrets, and MongoDB URIs with embedded credentials require `DBADMIN_ENCRYPTION_KEY` and are encrypted at rest using AES-256-GCM
- **Credential Redaction**: Passwords and API keys never logged or exposed in API responses
- **Dangerous Operation Detection**: Destructive commands require explicit confirmation
- **Read-only Mode**: Block all write operations per connection
- **Audit Export**: Local audit events include role, request ID, source address, RBAC denial reason, and JSON/NDJSON export support
- **Session Timeout**: Automatic logout after configurable inactivity period
- **Input Validation**: SQL injection and command injection prevention
- **Secure Cookies**: HttpOnly and SameSite attributes on session cookies; Secure is set for HTTPS and trusted HTTPS reverse-proxy requests

## Troubleshooting

### Frontend not loading
- Ensure you've run `npm run build` in the `web` directory
- Check that the backend is running on the expected port
- Verify no firewall is blocking the connection

### Connection test fails
- Verify database credentials are correct
- Check that the database server is running and accessible
- For remote databases, ensure firewall allows connections from your IP
- For MongoDB/Redis, verify authentication method matches your server config

### Query execution errors
- Check database user has appropriate permissions
- For read-only connections, write operations will be blocked
- Verify SQL/DSL syntax is correct for your database version
- Check dbadmin logs for detailed error messages

### Performance issues
- For large datasets, use pagination (default 50 rows per page)
- Add indexes to frequently queried columns
- Use filters to reduce result set size
- For Elasticsearch, use `size` parameter to limit results

### Data directory errors
- Ensure `DBADMIN_DATA_DIR` path exists and is writable
- Check disk space is available
- Verify file permissions allow read/write access

## Contributing

Contributions are welcome! Please:

1. Fork the repository
2. Create a feature branch
3. Add tests for new functionality
4. Ensure all tests pass
5. Submit a pull request

See [docs/manual-test.md](docs/manual-test.md) for testing guidelines.

## License

MIT

## Acknowledgments

Built with:
- [Go](https://golang.org/) - Backend runtime
- [React](https://react.dev/) - Frontend framework
- [Tailwind CSS](https://tailwindcss.com/) - Styling
- [Monaco Editor](https://microsoft.github.io/monaco-editor/) - Code editor
- plumego framework - HTTP server and utilities
