# Plumego Dev Server with Dashboard

This document describes the enhanced `plumego dev` command with an optional web-based dashboard.

## Overview

The `plumego dev` command now supports a **dual-server architecture** with an optional development dashboard that provides real-time monitoring, route inspection, and application management.

## Architecture

```
┌─────────────────────────────────────────────┐
│  plumego dev --dashboard :9999              │
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
- ✅ **Hot Reload**: Automatic rebuild and restart on file changes
- ✅ **Dual Server Mode**: User app + Dashboard run simultaneously
- ✅ **Event-Driven**: PubSub architecture for loose coupling
- ✅ **WebSocket Streaming**: Real-time log and event streaming
- ✅ **Backward Compatible**: Legacy mode works without changes

### Dashboard Features
- 🚀 **Real-time Logs**: Capture and filter stdout/stderr
- 🛣️ **Route Browser**: Discover and display all HTTP routes
- 📊 **Metrics Dashboard**: Performance and health monitoring
- 🔨 **Build Management**: Manual build triggers and output
- 🔄 **App Control**: Start, stop, restart buttons
- 📋 **Event Stream**: All development events in one place

## Usage

### Basic Usage (Legacy Mode)
```bash
plumego dev
# Runs your app with hot reload (no dashboard)
```

### Dashboard Mode
```bash
plumego dev --dashboard :9999
# Runs your app at :8080 with dashboard at :9999
```

### Full Options
```bash
plumego dev \
  --addr :8080 \              # User app address
  --dashboard :9999 \         # Dashboard address (enables dashboard)
  --watch "**/*.go" \         # Watch patterns
  --exclude "**/vendor/**" \  # Exclude patterns
  --debounce 500ms            # File change debounce
```

## Dashboard UI

### Tabs

#### 📄 Logs
- Real-time application logs (stdout/stderr)
- Filter by level (info/warn/error)
- Log statistics (total, error count)
- Clear logs button

#### 🛣️ Routes
- Lists all HTTP routes from your application
- Fetches from `/_debug/routes.json` endpoint
- Color-coded HTTP methods
- Refresh button

#### 📊 Metrics
- **Dashboard**: Uptime, start time
- **Application**: Status, PID, health
- Auto-refreshes every 5 seconds
- Health check integration

#### 🔨 Build Output
- Build status (success/failure)
- Compilation output and errors
- Build duration

#### 🔔 Events
- All development events
- File changes, builds, restarts
- Timestamped event log

### Controls

- **🔄 Restart**: Rebuild and restart the application
- **🔨 Build**: Trigger a manual build
- **⏹️ Stop**: Stop the running application
- **🗑️ Clear Logs**: Clear the log display

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
    "healthDetails": {...}
  }
}
```

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
- `app.restart` - Application restarting
- `app.log` - Log message from application
- `app.error` - Error from application
- `app.health` - Health check result

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
- Plumego application with debug mode enabled
- Available ports for app and dashboard

## Backward Compatibility

The dashboard is **opt-in**. Without the `--dashboard` flag, the command behaves exactly as before:

```bash
# Old behavior (still works)
plumego dev

# New behavior (opt-in)
plumego dev --dashboard :9999
```

## Troubleshooting

### Dashboard not loading
- Ensure the port is not in use
- Check that the UI files are embedded or available on disk

### Routes not showing
- Verify your app runs with `--debug` flag
- Check if `/_debug/routes.json` endpoint is accessible
- Dashboard will fallback to probing common paths

### Hot reload not working
- Check file watch patterns (`--watch`)
- Verify files are not in excluded directories
- Check debounce setting (`--debounce`)

## Examples

### Simple App
```bash
cd my-plumego-app
plumego dev --dashboard :9999
# Visit http://localhost:9999 for dashboard
# Visit http://localhost:8080 for your app
```

### Custom Ports
```bash
plumego dev --addr :3000 --dashboard :9000
```

### Specific Watch Patterns
```bash
plumego dev --dashboard :9999 --watch "**/*.go,**/*.yaml"
```

## Future Enhancements

Potential improvements for future releases:
- [ ] Request profiling and flamegraphs
- [ ] Database query monitoring
- [ ] API endpoint testing UI
- [ ] Performance bottleneck detection
- [ ] Dependency graph visualization
- [ ] Live configuration editing

## Contributing

When modifying the dev server:
1. Keep backward compatibility
2. Test both legacy and dashboard modes
3. Update UI when adding new features
4. Document new API endpoints
5. Add event types to the list above
