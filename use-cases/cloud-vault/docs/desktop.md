# Desktop Application Guide

This guide covers desktop-specific features and workflows for Markdown Cloud Vault on Windows, macOS, and Linux.

## Overview

The desktop application provides a native experience with:
- System tray integration for background operation
- Native file dialogs for import/export
- Automatic updates (optional)
- Desktop notifications
- Keyboard shortcuts

## Installation

See [Installation Guide](./installation.md) for platform-specific installation instructions.

## First Launch

### Setup Wizard

On first launch, the setup wizard guides you through:

1. **Welcome Screen**
   - Introduction to Cloud Vault
   - Click "Get Started" to continue

2. **Admin Account Creation**
   - Username (3-50 characters, alphanumeric and underscore)
   - Email address (for account recovery)
   - Password (minimum 12 characters)
   - Click "Create Account"

3. **Storage Configuration**
   - Choose storage location:
     - **Default**: `~/Cloud Vault` (recommended)
     - **Custom**: Select any directory
   - Storage type:
     - **Local**: Documents stored in selected directory
     - **Qiniu**: Cloud storage (advanced)
   - Click "Next"

4. **Import Existing Documents** (Optional)
   - Select folder containing Markdown files
   - Review scan preview
   - Click "Import" or "Skip"

5. **Complete Setup**
   - Review configuration summary
   - Click "Finish" to launch the vault

### Main Window

After setup, the main window opens with:
- **Sidebar**: Navigation (Vault, Search, Import, etc.)
- **Content Area**: Document list, editor, or tools
- **Status Bar**: Connection status, sync indicator
- **System Tray**: Background operation (if enabled)

## System Tray

The system tray icon provides quick access to Cloud Vault:

**Left-click**: Show/hide main window
**Right-click**: Context menu with:
- Show/Hide Window
- Open Vault Folder
- Check for Updates
- Export Diagnostics
- Quit Application

### Close to Tray

By default, closing the window minimizes to system tray instead of quitting:
- **Windows**: Tray icon in system tray area
- **macOS**: Menu bar icon
- **Linux**: System tray or app indicator

To quit completely:
- Right-click tray icon → Quit
- Or use File → Quit menu
- Or press `Ctrl+Q` (Windows/Linux) or `Cmd+Q` (macOS)

### Disable Close to Tray

To close window instead of minimizing to tray:
1. Go to **Settings** → **Desktop**
2. Uncheck "Close to tray"
3. Restart application

## Native File Dialogs

The desktop app uses native file dialogs for better integration:

**Import Documents:**
- Click "Select Directory" in Import page
- Native folder picker opens
- Select folder containing Markdown files
- Click "Choose" or "Open"

**Export Documents:**
- Select document(s) in Vault
- Right-click → Export
- Native save dialog opens
- Choose location and filename
- Click "Save"

**Backup/Restore:**
- Settings → Backup → Create Backup
- Native save dialog for backup location
- Choose where to save `.zip` backup file

## Keyboard Shortcuts

### Global Shortcuts

| Shortcut | Action |
|----------|--------|
| `Ctrl/Cmd + N` | New document |
| `Ctrl/Cmd + O` | Open file |
| `Ctrl/Cmd + S` | Save document |
| `Ctrl/Cmd + F` | Search |
| `Ctrl/Cmd + ,` | Settings |
| `Ctrl/Cmd + Q` | Quit application |
| `F1` | Help / Documentation |

### Navigation

| Shortcut | Action |
|----------|--------|
| `Ctrl/Cmd + 1` | Go to Vault |
| `Ctrl/Cmd + 2` | Go to Search |
| `Ctrl/Cmd + 3` | Go to Import |
| `Ctrl/Cmd + 4` | Go to Collections |
| `Ctrl/Cmd + 5` | Go to Tags |

### Document Editing

| Shortcut | Action |
|----------|--------|
| `Ctrl/Cmd + B` | Bold text |
| `Ctrl/Cmd + I` | Italic text |
| `Ctrl/Cmd + K` | Insert link |
| `Ctrl/Cmd + Shift + K` | Insert code block |
| `Tab` | Indent list item |
| `Shift + Tab` | Outdent list item |

## Desktop Notifications

Cloud Vault can send desktop notifications for:
- Import job completion
- Sync errors
- Update availability
- Backup completion

### Enable/Disable Notifications

1. Go to **Settings** → **Desktop**
2. Toggle "Show desktop notifications"
3. Restart application

### Notification Types

**Import Complete**: When batch import finishes
**Sync Error**: When cloud sync fails (Qiniu)
**Update Available**: When new version is detected
**Backup Complete**: When backup job finishes

## Automatic Updates

The desktop app can check for updates automatically:

### Check for Updates Manually

1. Go to **Settings** → **About**
2. Click "Check for Updates"
3. If update available, click "Download"
4. Installer downloads in background
5. Click "Install and Restart" when ready

### Automatic Update Checks

By default, Cloud Vault checks for updates:
- On application startup
- Every 24 hours while running

To change update behavior:
1. Go to **Settings** → **About**
2. Adjust "Update check interval"
3. Toggle "Check for updates automatically"

### Update Process

When an update is available:
1. Notification appears in system tray
2. Click notification or go to Settings → About
3. Review release notes
4. Click "Download Update"
5. Wait for download to complete
6. Click "Install and Restart"
7. Application closes and installs update
8. Application restarts with new version

**Note**: Updates are installed to the same location as the original installation.

## Data Directory

The desktop app stores data in a user-specific directory:

**Windows**: `%APPDATA%\Cloud Vault\`
**macOS**: `~/Library/Application Support/Cloud Vault/`
**Linux**: `~/.config/cloud-vault/`

### Directory Structure

```
Cloud Vault/
├── config.toml          # Configuration file
├── data/
│   ├── app.db          # SQLite database
│   └── objects/        # Document storage
├── logs/
│   └── app.log         # Application logs
├── backups/            # Backup files
└── diagnostics/        # Diagnostic bundles
```

### Change Data Directory

To move your vault to a different location:

1. Quit Cloud Vault completely
2. Copy the entire data directory to new location
3. Edit `config.toml`:
   ```toml
   [storage]
   root = "/path/to/new/location/data/objects"
   ```
4. Restart Cloud Vault

Or use the Setup Wizard:
1. Quit Cloud Vault
2. Delete or rename the data directory
3. Restart Cloud Vault
4. Setup wizard will prompt for new location

### Open Data Directory

To quickly access the data directory:
- Right-click system tray icon → "Open Vault Folder"
- Or go to **Settings** → **Desktop** → "Open Data Directory"

## Desktop-Specific Settings

### Appearance

**Theme**:
- System (follows OS theme)
- Light
- Dark

**Font Size**:
- Small (12px)
- Medium (14px, default)
- Large (16px)

**Window Size**:
- Remembers last window size and position
- Reset to default: Settings → Desktop → "Reset Window"

### Performance

**Hardware Acceleration**:
- Enabled by default for smooth rendering
- Disable if experiencing display issues:
  - Settings → Desktop → Advanced
  - Uncheck "Enable hardware acceleration"
  - Restart application

**Background Sync**:
- Syncs with cloud storage when app is minimized
- Disable to save resources:
  - Settings → Desktop
  - Uncheck "Sync in background"

### Startup

**Launch at Login**:
- Start Cloud Vault automatically on system boot
- Settings → Desktop → "Launch at login"

**Start Minimized**:
- Launch directly to system tray
- Settings → Desktop → "Start minimized"

## Troubleshooting

### App Won't Start

**Windows**:
- Check Windows Event Viewer for errors
- Run as administrator (right-click → Run as administrator)
- Reinstall application

**macOS**:
- Check Console app for crash logs
- Reset preferences: `defaults delete com.cloudvault.app`
- Reinstall application

**Linux**:
- Run from terminal to see error messages
- Check `~/.config/cloud-vault/logs/app.log`
- Verify dependencies: `ldd ./cloud-vault`

### System Tray Icon Missing

**Windows**:
- Check hidden icons area (arrow in system tray)
- Right-click taskbar → Taskbar settings → Notification area
- Enable Cloud Vault icon

**Linux**:
- Install system tray extension for your desktop environment
- GNOME: Install "AppIndicator" extension
- KDE: Should work out of the box

### Notifications Not Showing

**Windows**:
- Check Windows notification settings
- Settings → System → Notifications → Cloud Vault → On

**macOS**:
- System Settings → Notifications → Cloud Vault
- Enable "Allow Notifications"

**Linux**:
- Install notification daemon: `sudo apt install notification-daemon`
- Check desktop environment notification settings

### High CPU/Memory Usage

1. Check Import page for stuck jobs
2. Disable background sync: Settings → Desktop
3. Reduce indexed documents: Settings → Search → "Rebuild Index"
4. Restart application

## Platform-Specific Features

### Windows

- **Jump List**: Right-click taskbar icon for quick access
- **Live Tiles**: Pin to Start menu with live preview
- **Windows Search**: Vault documents appear in Windows Search (optional)

### macOS

- **Spotlight Integration**: Vault documents appear in Spotlight (optional)
- **Handoff**: Continue editing on iOS device (future feature)
- **Touch Bar**: Quick formatting options (MacBook Pro)

### Linux

- **AppImage**: Portable, no installation required
- **Flatpak**: Sandboxed installation via Flathub
- **Snap**: Automatic updates via Snap Store

## Next Steps

- [Configuration Guide](./configuration.md) - Customize settings
- [Import Guide](./import.md) - Add documents to your vault
- [Keyboard Shortcuts](#keyboard-shortcuts) - Speed up your workflow

---

**Last updated**: 2026-05-30
