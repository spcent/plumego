# Cloud Vault Desktop V0.9 Verification Checklist

## Overview
This checklist verifies the V0.9 desktop wrapper implementation, including:
- Wails v2 integration
- Local HTTP server with token authentication
- System tray functionality
- Desktop service APIs
- Configuration management
- Platform-specific features

---

## Build Verification

### Go Build
- [ ] Desktop package compiles successfully
  ```bash
  go build ./internal/desktop/...
  ```
- [ ] Desktop command builds successfully
  ```bash
  go build -o bin/cloud-vault-desktop ./cmd/desktop
  ```
- [ ] All dependencies resolve correctly
  ```bash
  go mod tidy
  go mod verify
  ```

### Frontend Build
- [ ] Web frontend builds successfully
  ```bash
  cd web && npm install && npm run build
  ```
- [ ] Frontend assets embedded correctly in `cmd/desktop/frontend/dist/`
- [ ] `index.html` loads in browser

### Wails Build
- [ ] Wails development mode starts
  ```bash
  make desktop-dev
  ```
- [ ] Wails production build succeeds
  ```bash
  make desktop-build
  ```

---

## Configuration Verification

### Default Configuration
- [ ] `DefaultConfig()` returns valid configuration
- [ ] App name defaults to "Cloud Vault"
- [ ] Data directory path is platform-appropriate:
  - macOS: `~/Library/Application Support/CloudVault`
  - Windows: `%APPDATA%\CloudVault`
  - Linux: `~/.local/share/cloudvault` or `$XDG_DATA_HOME/cloudvault`
- [ ] CloseToTray defaults to `true`
- [ ] NativeNotifications defaults to `true`

### Configuration Loading
- [ ] Config loads from `config.toml` when present
- [ ] Environment variables override config file values
- [ ] Missing config file uses defaults without error
- [ ] Invalid config values produce clear error messages

### Data Directory Setup
- [ ] `EnsureDataDir()` creates all required subdirectories:
  - `data/`
  - `objects/`
  - `backups/`
  - `logs/`
  - `cache/`
- [ ] Directories have correct permissions (0755)
- [ ] Function is idempotent (safe to call multiple times)

---

## Local HTTP Server Verification

### Server Initialization
- [ ] Server binds to `127.0.0.1:0` (localhost only, random port)
- [ ] Server never binds to `0.0.0.0` or public interfaces
- [ ] Random 32-byte hex token generated on startup
- [ ] Token is cryptographically secure (`crypto/rand`)
- [ ] Port selection finds available port automatically

### Token Authentication
- [ ] Requests without token are rejected (401)
- [ ] Token accepted via query parameter: `?token=...`
- [ ] Token accepted via header: `X-Auth-Token: ...`
- [ ] Invalid tokens are rejected (401)
- [ ] Token validation uses constant-time comparison
- [ ] All API routes require token authentication

### Server Lifecycle
- [ ] `Start()` begins listening without blocking
- [ ] `Port()` returns correct assigned port
- [ ] `BaseURL()` returns `http://127.0.0.1:PORT`
- [ ] `Stop()` gracefully shuts down server
- [ ] Server handles concurrent requests correctly

---

## Desktop Service Verification

### Runtime Info API
- [ ] `GetRuntimeInfo()` returns complete information:
  - Mode: "desktop"
  - Version: matches build version
  - Data directory path
  - Storage path
  - Platform (darwin/windows/linux)
  - Process ID
  - Start time
- [ ] When server is running:
  - Port field populated
  - BaseURL field populated
  - TokenEnabled: true
- [ ] When server is not running:
  - Port: 0
  - BaseURL: ""
  - TokenEnabled: false

### Scan Preview API
- [ ] Scans directory recursively when `Recursive: true`
- [ ] Scans only top-level when `Recursive: false`
- [ ] Correctly counts:
  - Total files
  - Markdown files (.md, .markdown)
  - Empty files (0 bytes)
  - Large files (> MaxFileSizeMB)
  - Skipped files (hidden, empty, large)
- [ ] Respects `IgnoreHidden` flag
- [ ] Respects `IgnoreEmpty` flag
- [ ] Respects `MaxFileSizeMB` limit
- [ ] Returns sample of up to 10 markdown file paths
- [ ] Collects errors without stopping scan
- [ ] Fails gracefully for invalid paths

### Data Directory Info API
- [ ] `GetDataDirInfo()` returns directory information:
  - Path matches config
  - Exists flag accurate
  - IsWritable flag accurate
- [ ] Creates directory if it doesn't exist
- [ ] Handles permission errors gracefully

---

## System Tray Verification

### Tray Initialization
- [ ] Tray icon appears in system tray
- [ ] Tooltip shows "Cloud Vault - Running"
- [ ] Icon loads from embedded resources
- [ ] Fallback to empty icon if resource missing

### Tray Menu
- [ ] "Show Window" menu item present
- [ ] "Open Data Directory" menu item present
- [ ] "Scan Import Directory" menu item present
- [ ] Separator between menu sections
- [ ] "Runtime Info" menu item present
- [ ] "Close to Tray" checkbox present and reflects config
- [ ] "Quit" menu item present

### Tray Menu Actions
- [ ] "Open Data Directory" opens file manager:
  - macOS: Finder
  - Windows: Explorer
  - Linux: xdg-open
- [ ] "Runtime Info" logs runtime information
- [ ] "Close to Tray" toggles checkbox state
- [ ] "Quit" initiates graceful shutdown

### Window Visibility
- [ ] `ShowWindow()` sets visible state to true
- [ ] `HideWindow()` sets visible state to false
- [ ] `IsVisible()` returns current state
- [ ] Thread-safe with mutex protection

### Close-to-Tray Behavior
- [ ] When `CloseToTray` enabled:
  - Window close hides instead of quitting
  - Tray icon remains visible
  - Can restore from tray menu
- [ ] When `CloseToTray` disabled:
  - Window close triggers full quit
  - Application exits

### Notifications
- [ ] `ShowNotification()` displays notification when enabled
- [ ] `ShowNotification()` logs only when disabled
- [ ] Notification includes title and message
- [ ] Wails native notifications preferred
- [ ] Fallback to logging if Wails unavailable

---

## Wails Bindings Verification

### Binding Initialization
- [ ] `WailsBindings` struct created successfully
- [ ] Context set by Wails runtime
- [ ] All dependencies injected (config, server, service, tray, logger)

### Exposed Methods
- [ ] `GetRuntimeInfo()` callable from frontend
- [ ] `SelectDirectory()` opens native dialog
- [ ] `ScanPreview()` delegates to service
- [ ] `OpenDataDirectory()` opens file manager
- [ ] `ShowNotification()` displays notification
- [ ] `StartImport()` emits progress events
- [ ] `GetConfig()` returns current config
- [ ] `UpdateConfig()` updates config fields
- [ ] `Quit()` initiates shutdown
- [ ] `MinimizeToTray()` hides window
- [ ] `RestoreFromTray()` shows window

### Directory Selection
- [ ] `SelectDirectory()` uses `runtime.OpenDirectoryDialog`
- [ ] Default path falls back to data directory
- [ ] Returns selected path or empty string if cancelled
- [ ] Absolute path resolution works correctly

### Import Progress
- [ ] `StartImport()` emits initial progress event
- [ ] Event name format: `import:progress:{jobID}`
- [ ] Progress includes job ID and status
- [ ] Events skipped if context is nil

---

## Integration with App Layer

### HTTP Handler
- [ ] `App.HTTPHandler()` returns valid handler
- [ ] Handler calls `Core.Prepare()`
- [ ] Handler implements `http.Handler` interface
- [ ] Can be embedded in desktop server

### Background Tasks
- [ ] `App.StartBackgroundTasks()` starts indexer
- [ ] `App.StartBackgroundTasks()` starts AI workers
- [ ] Tasks respect context cancellation
- [ ] Graceful shutdown on context done

### Desktop Main Entry
- [ ] `cmd/desktop/main.go` loads config
- [ ] Overrides server address to `127.0.0.1:0`
- [ ] Disables TLS for desktop mode
- [ ] Overrides DB and storage paths to data directory
- [ ] Creates app instance
- [ ] Registers routes
- [ ] Starts local server
- [ ] Initializes desktop service
- [ ] Initializes tray manager
- [ ] Initializes Wails bindings
- [ ] Configures Wails app options
- [ ] Handles graceful shutdown on signal

---

## Platform-Specific Verification

### macOS
- [ ] Data directory: `~/Library/Application Support/CloudVault`
- [ ] File manager opens with `open` command
- [ ] System tray icon visible
- [ ] Native notifications work

### Windows
- [ ] Data directory: `%APPDATA%\CloudVault`
- [ ] File manager opens with `explorer` command
- [ ] System tray icon visible
- [ ] Native notifications work

### Linux
- [ ] Data directory: `~/.local/share/cloudvault` or `$XDG_DATA_HOME/cloudvault`
- [ ] File manager opens with `xdg-open` command
- [ ] System tray icon visible (if supported)
- [ ] Notifications fall back to logging

---

## Security Verification

### Token Security
- [ ] Token generated with `crypto/rand`
- [ ] Token length: 32 bytes (64 hex characters)
- [ ] Token validation uses `hmac.Equal` for timing-safe comparison
- [ ] Token never logged or exposed in error messages
- [ ] Token unique per application startup

### Network Security
- [ ] Server binds only to `127.0.0.1`
- [ ] Server never binds to `0.0.0.0`
- [ ] No public network exposure
- [ ] Port selection avoids privileged ports (<1024)

### Configuration Security
- [ ] Sensitive config values (API keys) not logged
- [ ] Config file permissions restricted (0600 recommended)
- [ ] Data directory permissions restricted (0700 recommended)

---

## Testing Verification

### Unit Tests
- [ ] All tests in `internal/desktop/desktop_test.go` pass
  ```bash
  go test ./internal/desktop/... -v
  ```
- [ ] Test coverage for configuration
- [ ] Test coverage for service methods
- [ ] Test coverage for tray manager
- [ ] Test coverage for helper functions

### Test Scenarios
- [ ] Default configuration validation
- [ ] Data directory path resolution
- [ ] Service initialization
- [ ] Runtime info retrieval
- [ ] Scan preview (recursive and non-recursive)
- [ ] Invalid path handling
- [ ] Data directory info
- [ ] Tray manager creation
- [ ] Window visibility toggling
- [ ] Close-to-tray behavior
- [ ] Hidden file detection
- [ ] Platform detection

---

## Performance Verification

### Startup Time
- [ ] Desktop app starts in < 3 seconds
- [ ] Tray icon appears immediately
- [ ] HTTP server ready within 1 second

### Memory Usage
- [ ] Idle memory usage reasonable (< 100MB)
- [ ] No memory leaks during operation
- [ ] Large directory scans don't consume excessive memory

### Responsiveness
- [ ] Tray menu responds instantly
- [ ] Directory selection dialog opens quickly
- [ ] Scan preview completes in reasonable time
- [ ] UI remains responsive during background tasks

---

## Error Handling Verification

### Configuration Errors
- [ ] Invalid config file produces clear error
- [ ] Missing data directory creates it automatically
- [ ] Permission errors reported clearly

### Network Errors
- [ ] Port conflict handled gracefully
- [ ] Server startup failure logged clearly
- [ ] Token generation failure reported

### File System Errors
- [ ] Invalid scan path returns error
- [ ] Permission denied errors reported
- [ ] Disk full errors handled
- [ ] Symlink loops don't cause infinite recursion

### Runtime Errors
- [ ] Tray initialization failure doesn't crash app
- [ ] Missing icon resource uses fallback
- [ ] Wails binding errors logged but don't crash

---

## Documentation Verification

### Code Documentation
- [ ] All public functions have GoDoc comments
- [ ] Complex logic explained inline
- [ ] Configuration options documented
- [ ] Error cases documented

### User Documentation
- [ ] README.md updated with V0.9 section
- [ ] Desktop mode usage documented
- [ ] Configuration options explained
- [ ] Platform-specific notes included

### Developer Documentation
- [ ] Architecture overview provided
- [ ] Integration points documented
- [ ] Testing instructions clear
- [ ] Build process documented

---

## Acceptance Criteria

All of the following must be true for V0.9 to be considered complete:

1. [ ] Desktop binary builds successfully
2. [ ] All unit tests pass
3. [ ] Local HTTP server binds to 127.0.0.1 only
4. [ ] Token authentication protects all API routes
5. [ ] System tray icon appears and menu works
6. [ ] Close-to-tray behavior functions correctly
7. [ ] Directory selection dialog works
8. [ ] Scan preview returns accurate results
9. [ ] Runtime info API returns complete data
10. [ ] Configuration loads from file and environment
11. [ ] Data directory created with all subdirectories
12. [ ] Platform-specific paths used correctly
13. [ ] File manager opens data directory
14. [ ] Notifications display when enabled
15. [ ] Graceful shutdown works correctly
16. [ ] No security vulnerabilities (token, network, config)
17. [ ] Performance meets requirements
18. [ ] Error handling is robust
19. [ ] Documentation is complete
20. [ ] All verification checklist items pass

---

## Test Commands

```bash
# Build verification
go build ./internal/desktop/...
go build -o bin/cloud-vault-desktop ./cmd/desktop

# Test verification
go test ./internal/desktop/... -v
go test ./internal/desktop/... -cover

# Run desktop app
./bin/cloud-vault-desktop

# Development mode
make desktop-dev

# Production build
make desktop-build

# Clean artifacts
make desktop-clean
```

---

## Sign-Off

- [ ] All checklist items verified
- [ ] No critical issues found
- [ ] Code review completed
- [ ] Security review completed
- [ ] Documentation review completed
- [ ] Ready for V0.9 release

**Verified by:** _________________
**Date:** _________________
**Notes:** _________________
