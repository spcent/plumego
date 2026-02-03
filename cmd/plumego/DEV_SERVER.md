# Plumego Dev Server with Dashboard

This document describes the enhanced `plumego dev` command with integrated web-based dashboard.

## Overview

The `plumego dev` command features a **dual-server architecture** with a development dashboard that provides real-time monitoring, route inspection, and application management. The dashboard is **built with plumego itself**, demonstrating the framework's capabilities.

## Architecture

```
┌─────────────────────────────────────────────┐
│  plumego dev                                │
└────────────────┬────────────────────────────┘
                 │
        ┌────────┴────────┐
        │                 │
   ┌────▼─────┐    ┌─────▼────────┐
   │ User App │    │ Dev Dashboard│
   │  :8080   │    │   :9999      │
   └──────────┘    └──────┬───────┘
        │                 │
        │         ┌───────┴────────┐
        │         │ Event Bus      │
        │         │ (PubSub)       │
        │         │ - File changes │
        │         │ - Build events │
        │         │ - App logs     │
        │         └────────────────┘
        │
   Hot Reload (< 5s)
```

## Features

### Core Features
- **Hot Reload**: Automatic rebuild and restart on file changes (< 5 seconds)
- **Dual Server Mode**: User app + Dashboard run simultaneously
- **Event-Driven**: PubSub architecture for loose coupling
- **WebSocket Streaming**: Real-time log and event streaming
- **Dogfooding**: Dashboard built with plumego framework itself

### Dashboard Features (Default)
- **Real-time Logs**: Capture and filter stdout/stderr
- **Route Browser**: Discover and display all HTTP routes
- **Metrics Dashboard**: Performance and health monitoring
- **Dependency Graph**: Visualize module relationships and hot spots
- **Profiling**: Capture CPU/heap/trace profiles and open flamegraphs
- **API Testing**: Send ad-hoc requests to your app from the dashboard
- **Build Management**: Manual build triggers and output
- **App Control**: Restart/build/stop controls
- **Event Stream**: All development events in one place

## Usage

### Quick Start
```bash
plumego dev
# Dashboard: http://localhost:9999
# Your app:  http://localhost:8080
```

### Custom Ports
```bash
# Custom application port
plumego dev --addr :3000

# Custom dashboard port
plumego dev --dashboard-addr :8888

# Both custom
plumego dev --addr :3000 --dashboard-addr :7777
```

### Full Options
```bash
plumego dev \
  --dir . \                         # Project directory (default: .)
  --addr :8080 \                    # User app address (default: :8080)
  --dashboard-addr :9999 \          # Dashboard address (default: :9999)
  --watch "**/*.go,**/*.yaml" \     # Watch patterns
  --exclude "**/vendor/**" \        # Exclude patterns
  --debounce 1s \                   # File change debounce (default: 500ms)
  --no-reload \                     # Disable file watcher / hot reload
  --build-cmd "go build -o .dev-server ./cmd/api" \  # Custom build
  --run-cmd "./.dev-server"         # Custom run command
```

## Dashboard UI

### Tabs

#### Logs
- Real-time application logs (stdout/stderr)
- Filter by level (info/warn/error)
- Log statistics (total, error count)
- Clear logs button

#### Routes
- Lists all HTTP routes from your application
- Fetches from `/_debug/routes.json` endpoint
- Color-coded HTTP methods
- Refresh button

#### Metrics
- **Dashboard**: Uptime, start time
- **Application**: Status, PID, health
- **Requests**: Aggregated request latency (p50/p95/p99), error rate, and top slow routes
- **Alerts**: Threshold-based latency and error-rate warnings
- **Database**: Query latency stats, slow query list, and redaction-aware drilldown (requires instrumented DB)
- Auto-refreshes every 5 seconds
- Health check integration
- Alert thresholds can be overridden in the dashboard (stored per-project in browser state)
- Alert thresholds can be exported/imported as JSON

#### Dependencies
- Module dependency graph based on `go list` output
- Summary of direct vs indirect modules and edge counts
- Top edges list for hot dependency paths

#### Profiling
- CPU/heap/allocs/goroutine/block/mutex/threadcreate/trace profiles
- Download raw pprof profiles
- Open flamegraphs in Speedscope (external viewer)

#### API Test
- Send ad-hoc requests to the running app
- Edit method/path/headers/body
- View status, headers, body, and timing
- Request history and saved presets
- Inline JSON viewer with collapsible formatting
- Export/import saved requests as JSON

#### Build Output
- Build status (success/failure)
- Compilation output and errors
- Build duration

#### Events
- All development events
- File changes, builds, restarts
- Timestamped event log

### Controls

- **Restart**: Rebuild and restart the application
- **Build**: Trigger a manual build
- **Stop**: Stop the running application
- **Clear Logs**: Clear the log display

## API Endpoints

The dashboard exposes the following API endpoints:

### Application Status
```bash
GET /api/status
```
Returns dashboard URL, app URL, running status, PID, and project directory.

### Health Check
```bash
GET /api/health
```
Returns application health status.

### Routes
```bash
GET /api/routes
```
Returns all HTTP routes discovered from the application.

**Response:**
```json
{
  "count": 9,
  "routes": [
    {"method": "GET", "path": "/"},
    {"method": "GET", "path": "/health"},
    {"method": "POST", "path": "/api/users"}
  ]
}
```

### Metrics
```bash
GET /api/metrics
POST /api/metrics/clear
```
Returns dashboard and application metrics.

**Response:**
```json
{
  "dashboard": {
    "uptime": 123.45,
    "startTime": "2026-02-02T11:06:29Z"
  },
  "app": {
    "running": true,
    "pid": 12345,
    "healthy": true,
    "healthDetails": {...},
    "requests": {
      "window_seconds": 300,
      "started_at": "2026-02-02T11:06:29Z",
      "total": {
        "count": 1280,
        "error_count": 12,
        "duration_ms": {"p50": 12.2, "p95": 45.1, "p99": 80.3}
      },
      "routes": [
        {"method": "GET", "path": "/api/users", "count": 420, "duration_ms": {"p95": 32.4}}
      ]
    },
    "db": {
      "total": {
        "count": 88,
        "error_count": 2,
        "duration_ms": {"p95": 30.4}
      },
      "slow": [
        {"operation": "query", "driver": "mysql", "duration_ms": 520.2}
      ]
    }
  }
}
```
Includes `alerts` and `thresholds` fields for performance bottleneck detection.

### Dependencies
```bash
GET /api/deps?include_std=1&max_nodes=20&refresh=1
```
Returns module dependency graph data derived from `go list`.

Query parameters:
- `include_std` (default `1`): include standard library packages
- `max_nodes` (default `0` for all): limit node count for visualization
- `refresh` (default `0`): force graph refresh (otherwise cached briefly)

Response includes `summary`, `nodes`, and `edges`.

### Profiling
```bash
GET /api/pprof/types
GET /api/pprof/raw?type=cpu&seconds=10
GET /api/pprof/raw?type=cpu&seconds=10&download=0
```
Returns supported profile types and raw pprof data (for Speedscope or download).

**Notes**
- CPU/trace profiles accept a `seconds` parameter (defaults applied).
- Raw pprof responses include CORS headers for Speedscope integration.
- `download=0` returns a JSON preview (size, content type, hex sample) for inline display.

### Configuration
```bash
GET /api/config
```
Returns application configuration from `/_debug/config`.

### Build Control
```bash
POST /api/build
POST /api/restart
POST /api/stop
```
Control application lifecycle.

### API Testing
```bash
POST /api/test
```
Send an API request to the running application and return status, headers, and body.

**Request:**
```json
{
  "method": "POST",
  "path": "/api/users",
  "query": "debug=true",
  "headers": {"Content-Type": "application/json"},
  "body": "{\"name\":\"demo\"}"
}
```

## WebSocket

Real-time events are streamed over WebSocket:

```javascript
const ws = new WebSocket('ws://localhost:9999/ws');

ws.onmessage = (event) => {
  const message = JSON.parse(event.data);
  // message.type: event type
  // message.data: event payload
};
```

### Event Types
- `file.change` - File modification detected
- `build.start` - Build started
- `build.success` - Build completed successfully
- `build.fail` - Build failed
- `app.start` - Application starting/started
- `app.stop` - Application stopped
- `app.log` - Log message from application
- `app.error` - Error from application

## Implementation Details

### Components

#### Dashboard Server (`dashboard.go`)
- Built with plumego framework itself (dogfooding)
- Manages WebSocket hub for real-time communication
- Coordinates PubSub events
- Serves embedded UI

#### Application Runner (`runner.go`)
- Manages user application lifecycle
- Captures stdout/stderr streams
- Publishes lifecycle events
- Graceful shutdown (SIGTERM → SIGKILL)

#### Builder (`builder.go`)
- Handles Go compilation
- Publishes build events
- Custom build command support

#### Analyzer (`analyzer.go`)
- Discovers routes from `/_debug/routes.json`
- Probes common endpoints (fallback)
- Health check integration
- Configuration fetching
- DB metrics are available when your app uses instrumented DB helpers in `store/db`

#### Database Monitoring
To collect query metrics, wrap your `*sql.DB` with `store/db` instrumentation and pass a metrics collector:

```go
collector := metrics.NewBaseMetricsCollector()
dbConn, _ := sql.Open("mysql", dsn)
instrumented := db.NewInstrumentedDB(dbConn, collector, "mysql")
// Use instrumented in your data layer
```
Slow query drilldown displays redacted SQL and highlights the redaction rules used.

### Embedded UI

UI resources (HTML/CSS/JS) are embedded into the binary using Go's `embed` package:

```go
//go:embed ui/*
var uiFS embed.FS
```

Fallback to disk-based serving for development.

## Performance

| Metric | Value |
|--------|-------|
| Binary Size | 13MB |
| Dashboard Overhead | < 50MB RAM |
| Hot Reload Time | < 5 seconds |
| API Response Time | < 100ms |
| Build Time (simple app) | ~2 seconds |

## Requirements

- Go 1.24+
- Plumego application with debug endpoints enabled (`core.WithDebug` or honoring `APP_DEBUG`; dev server sets `APP_DEBUG=true`)
- Available ports for app and dashboard

## Positioning & Production Guidance

- `core.WithDebug` exposes app-level `/_debug` endpoints. Use only in local/dev or protect them in production.
- `plumego dev` dashboard is a local developer tool that runs a separate dashboard server; do not expose it publicly in production.
- The dashboard may query app `/_debug` endpoints for routes/config, so keep debug endpoints gated outside local/dev usage.
- Profiling uses the app `/_debug/pprof` endpoints, so keep them gated outside local/dev usage.

## Backward Compatibility

The dashboard is always enabled. Use `--dashboard-addr` to change the port, or `--no-reload` to disable file watching:

```bash
# Default behavior
plumego dev

# Disable auto reload (no watcher)
plumego dev --no-reload
```

## Troubleshooting

### Dashboard not loading
- Ensure the port is not in use
- Check that the UI files are embedded or available on disk

### Routes not showing
- Verify your app enables debug endpoints (`core.WithDebug` or `APP_DEBUG`)
- Check if `/_debug/routes.json` endpoint is accessible
- Dashboard will fallback to probing common paths

### Request metrics not showing
- Ensure logging middleware is enabled (`core.WithLogging` or `core.WithRecommendedMiddleware`)
- Verify `/_debug/metrics` endpoint is accessible (debug must be enabled)

### Database monitoring not showing
- Ensure your DB is instrumented (see `store/db` helpers)
- Confirm the app is passing a metrics collector to the instrumented DB

### Hot reload not working
- Check file watch patterns (`--watch`)
- Verify files are not in excluded directories
- Check debounce setting (`--debounce`)

### Profiling not working
- Ensure debug endpoints are enabled (`core.WithDebug` or `APP_DEBUG`)
- Verify `/_debug/pprof` endpoints are reachable
- For Speedscope, confirm the dashboard endpoint is reachable from the browser

## Examples

### Simple App
```bash
cd my-plumego-app
plumego dev --dashboard-addr :9999
# Visit http://localhost:9999 for dashboard
# Visit http://localhost:8080 for your app
```

### Custom Ports
```bash
plumego dev --addr :3000 --dashboard-addr :9000
```

### Specific Watch Patterns
```bash
plumego dev --dashboard-addr :9999 --watch "**/*.go,**/*.yaml"
```

## Future Enhancements

Potential improvements for future releases:
- [x] Request profiling and flamegraphs
- [x] Database query monitoring
- [x] API endpoint testing UI
- [x] Performance bottleneck detection
- [x] Dependency graph visualization
- [ ] Live configuration editing

## Contributing

When modifying the dev server:
1. Keep backward compatibility
2. Test both legacy and dashboard modes
3. Update UI when adding new features
4. Document new API endpoints
5. Add event types to the list above
