# Installation Guide

This guide covers installation Markdown Cloud Vault on different platforms and deployment scenarios.

## Desktop Application

### Windows

**System Requirements:**
- Windows 10 or later (64-bit)
- 4 GB RAM minimum (8 GB recommended)
- 500 MB disk space

**Installation Steps:**
1. Download the Windows installer: `cloud-vault-v1.0.0-windows-amd64.exe`
2. Double-click the installer
3. Follow the setup wizard:
   - Accept the license agreement
   - Choose installation directory (default: `C:\Program Files\Cloud Vault`)
   - Select "Create desktop shortcut" (recommended)
4. Click "Install" and wait for completion
5. Launch from Start menu or desktop shortcut

**Uninstalling:**
- Go to Settings → Apps → Installed apps
- Find "Markdown Cloud Vault" and click "Uninstall"
- Or run the uninstaller from the installation directory

### macOS

**System Requirements:**
- macOS 12 (Monterey) or later
- Intel or Apple Silicon (M1/M2/M3)
- 4 GB RAM minimum (8 GB recommended)
- 500 MB disk space

**Installation Steps:**
1. Download the appropriate installer:
   - Apple Silicon: `cloud-vault-v1.0.0-macos-arm64.dmg`
   - Intel: `cloud-vault-v1.0.0-macos-amd64.dmg`
2. Double-click the `.dmg` file to mount it
3. Drag "Cloud Vault" to the Applications folder
4. Eject the mounted disk image
5. Launch from Applications folder or Spotlight

**First Launch Security Warning:**
macOS may show a security warning for unsigned apps. To bypass:
- Right-click the app in Applications
- Select "Open"
- Click "Open" in the dialog

Or enable via System Settings:
- Go to System Settings → Privacy & Security
- Scroll to "Security" section
- Click "Open Anyway" for Cloud Vault

**Uninstalling:**
- Drag the app from Applications to Trash
- Empty Trash to complete removal

### Linux

**System Requirements:**
- Linux with glibc 2.31+ (Ubuntu 20.04+, Debian 11+, Fedora 36+)
- 4 GB RAM minimum (8 GB recommended)
- 500 MB disk space

**Installation Methods:**

**AppImage (Recommended):**
1. Download: `cloud-vault-v1.0.0-linux-amd64.AppImage`
2. Make executable: `chmod +x cloud-vault-v1.0.0-linux-amd64.AppImage`
3. Run directly: `./cloud-vault-v1.0.0-linux-amd64.AppImage`
4. (Optional) Move to `/opt/` or `~/Applications/` for permanent installation

**Debian/Ubuntu Package:**
```bash
sudo dpkg -i cloud-vault-v1.0.0-linux-amd64.deb
sudo apt-get install -f  # Fix dependencies if needed
```

**Arch Linux (AUR):**
```bash
yay -S cloud-vault-bin
# or
paru -S cloud-vault-bin
```

**Uninstalling:**
- AppImage: Delete the `.AppImage` file
- Debian package: `sudo apt-get remove cloud-vault`
- Arch: `sudo pacman -R cloud-vault-bin`

## Server Deployment

### Binary Installation

**Download:**
```bash
# Linux amd64
wget https://releases.example.com/cloud-vault-v1.0.0-linux-amd64.tar.gz
tar -xzf cloud-vault-v1.0.0-linux-amd64.tar.gz
cd cloud-vault

# macOS
curl -L https://releases.example.com/cloud-vault-v1.0.0-macos-amd64.tar.gz | tar -xz
cd cloud-vault

# Windows
# Download and extract cloud-vault-v1.0.0-windows-amd64.zip
```

**Verify Installation:**
```bash
./cloud-vault --version
# Output: Cloud Vault v1.0.0 (commit: abc1234, built: 2026-05-30)
```

**Run Server:**
```bash
./cloud-vault --addr :8080
```

See [Server Deployment Guide](./server-deploy.md) for production setup with systemd, Docker, and reverse proxies.

### Docker Installation

**Pull Image:**
```bash
docker pull cloudvault/cloud-vault:1.0.0
```

**Run Container:**
```bash
docker run -d \
  --name cloud-vault \
  -p 8080:8080 \
  -v ./data:/app/data \
  -v ./config.toml:/app/config.toml \
  cloudvault/cloud-vault:1.0.0
```

**Docker Compose:**
```yaml
version: '3.8'
services:
  cloud-vault:
    image: cloudvault/cloud-vault:1.0.0
    ports:
      - "8080:8080"
    volumes:
      - ./data:/app/data
      - ./config.toml:/app/config.toml
    restart: unless-stopped
```

See [Server Deployment](./server-deploy.md) for production Docker setup with TLS and backups.

### Build from Source

**Prerequisites:**
- Go 1.22 or later
- Node.js 18+ and npm (for frontend)
- Make (optional)

**Clone and Build:**
```bash
git clone https://github.com/example/cloud-vault.git
cd cloud-vault

# Build server binary
make build-server

# Build frontend assets
cd web && npm install && npm run build
cd ..

# Run server
./bin/cloud-vault --addr :8080
```

**Build Desktop App:**
```bash
# Requires Wails CLI
go install github.com/wailsapp/wails/v2/cmd/wails@latest

# Build desktop app
make build-desktop
```

## Post-Installation

### Initial Setup

1. **Create Admin Account**
   - First launch shows setup wizard
   - Enter admin username, email, and password
   - Password must be at least 12 characters

2. **Configure Storage**
   - Local storage (default): Documents stored in `./data/objects`
   - Qiniu cloud: See [Qiniu Setup](./qiniu.md)

3. **Import Documents**
   - Use Import page to add Markdown files
   - See [Import Guide](./import.md) for batch import options

### Verify Installation

**Check Health:**
```bash
curl http://localhost:8080/health
# Response: {"status":"ok","version":"1.0.0"}
```

**Check Version:**
```bash
curl http://localhost:8080/api/v1/system/version
# Response: {"version":"1.0.0","commit":"abc1234","build_time":"2026-05-30T10:00:00Z"}
```

## Troubleshooting Installation

### Windows Issues

**Problem:** Installer blocked by Windows Defender
**Solution:**
- Right-click installer → Properties → Unblock
- Or add exception in Windows Security

**Problem:** "Application cannot be started"
**Solution:**
- Install Visual C++ Redistributable 2015-2022
- Check Windows Event Viewer for error details

### macOS Issues

**Problem:** "App is damaged and can't be opened"
**Solution:**
```bash
sudo xattr -rd com.apple.quarantine /Applications/Cloud\ Vault.app
```

**Problem:** Gatekeeper blocks unsigned app
**Solution:**
- System Settings → Privacy & Security → Open Anyway
- Or use `spctl --add` to whitelist

### Linux Issues

**Problem:** AppImage won't run
**Solution:**
```bash
chmod +x cloud-vault-v1.0.0-linux-amd64.AppImage
# If FUSE error:
./cloud-vault-v1.0.0-linux-amd64.AppImage --appimage-extract
./squashfs-root/AppRun
```

**Problem:** Missing dependencies
**Solution:**
```bash
# Ubuntu/Debian
sudo apt-get install libwebkit2gtk-4.0-dev libgtk-3-dev

# Fedora
sudo dnf install webkit2gtk3-devel gtk3-devel
```

### Server Issues

**Problem:** Port 8080 already in use
**Solution:**
```bash
# Use different port
./cloud-vault --addr :8081

# Or find and stop conflicting process
lsof -i :8080
kill -9 <PID>
```

**Problem:** Permission denied on data directory
**Solution:**
```bash
# Create data directory with correct permissions
mkdir -p ./data/objects
chmod 755 ./data
chown -R $USER:$USER ./data
```

## Next Steps

- [Configuration Guide](./configuration.md) - Customize settings
- [Server Deployment](./server-deploy.md) - Production setup
- [Getting Started](./getting-started.md) - First steps after installation

---

**Last updated**: 2026-05-30
