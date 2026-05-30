# Cloud Vault systemd Installation Guide

This guide explains how to install and run Cloud Vault as a systemd service on Linux systems.

## Prerequisites

1. Build the release binary:
   ```bash
   make release
   ```

2. Create system user:
   ```bash
   sudo useradd --system --shell /sbin/nologin cloud-vault
   ```

3. Install application:
   ```bash
   sudo mkdir -p /opt/cloud-vault
   sudo cp dist/cloud-vault /opt/cloud-vault/
   sudo cp config.example.toml /opt/cloud-vault/config.toml
   sudo mkdir -p /opt/cloud-vault/data /opt/cloud-vault/backups
   sudo chown -R cloud-vault:cloud-vault /opt/cloud-vault
   ```

4. Edit configuration:
   ```bash
   sudo nano /opt/cloud-vault/config.toml
   ```

## Installation

1. Copy service file:
   ```bash
   sudo cp deploy/systemd/cloud-vault.service /etc/systemd/system/
   ```

2. Reload systemd:
   ```bash
   sudo systemctl daemon-reload
   ```

3. Enable service:
   ```bash
   sudo systemctl enable cloud-vault
   ```

4. Start service:
   ```bash
   sudo systemctl start cloud-vault
   ```

## Management

Check status:
```bash
sudo systemctl status cloud-vault
```

View logs:
```bash
sudo journalctl -u cloud-vault -f
```

Restart service:
```bash
sudo systemctl restart cloud-vault
```

Stop service:
```bash
sudo systemctl stop cloud-vault
```

## Backup

Create backup via API:
```bash
curl -X POST http://localhost:8080/api/v1/system/backup \
  -H "Cookie: session=<your-session-token>"
```

Or use the restore CLI:
```bash
sudo -u cloud-vault /opt/cloud-vault/cloud-vault restore --file /path/to/backup.zip
```

## Security Notes

- The service runs as a dedicated `cloud-vault` user with minimal privileges
- `ProtectSystem=strict` prevents writes outside designated paths
- `PrivateTmp=true` isolates temporary files
- Data is stored in `/opt/cloud-vault/data` and `/opt/cloud-vault/backups`
- Ensure proper file permissions on config.toml (may contain secrets)

## Troubleshooting

Check if port 8080 is available:
```bash
sudo ss -tlnp | grep 8080
```

Verify file permissions:
```bash
ls -la /opt/cloud-vault/
```

Test configuration:
```bash
sudo -u cloud-vault /opt/cloud-vault/cloud-vault --config /opt/cloud-vault/config.toml --test
```
