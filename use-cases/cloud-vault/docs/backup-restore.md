# Backup and Restore Guide

This guide covers backing up and restoring your Markdown Cloud Vault data.

## Overview

Cloud Vault provides:
- **Automatic backups**: Scheduled backups with retention
- **Manual backups**: On-demand backup creation
- **Restore**: Restore from backup files
- **Export**: Export data in standard formats

## Automatic Backups

### Configuration

```toml
[backup]
enabled = true
schedule = "0 2 * * *"  # Daily at 2 AM
retention_days = 30
max_backups = 10
compression = true
```

### Backup Location

Backups are stored in:
```
data/backups/
├── backup-2026-05-30-020000.zip
├── backup-2026-05-29-020000.zip
└── ...
```

### Disable Automatic Backups

```toml
[backup]
enabled = false
```

## Manual Backups

### Create Backup

**Via Web Interface**:
1. Navigate to **Settings** → **Backup**
2. Click **Create Backup**
3. Wait for backup to complete
4. Download backup file

**Via CLI**:
```bash
./cloud-vault backup create \
  --output /path/to/backup.zip \
  --compress
```

**Via API**:
```bash
curl -X POST http://localhost:8080/api/v1/backup \
  -H "Authorization: Bearer $TOKEN"
```

### Backup Contents

Each backup includes:
- `database.db` - SQLite database
- `config.toml` - Configuration file
- `metadata.json` - Backup metadata
- `documents/` - Document files (if local storage)

### Exclude from Backup

```toml
[backup]
exclude_patterns = [
  "*.tmp",
  "*.log",
  "cache/*"
]
```

## List Backups

**Via Web Interface**:
1. Navigate to **Settings** → **Backup**
2. View backup list with dates and sizes

**Via CLI**:
```bash
./cloud-vault backup list
```

**Via API**:
```bash
curl http://localhost:8080/api/v1/backup \
  -H "Authorization: Bearer $TOKEN"
```

## Restore

### Restore from Backup

**Via CLI**:
```bash
./cloud-vault backup restore \
  --backup /path/to/backup.zip \
  --data-dir ./data
```

**Via API**:
```bash
curl -X POST http://localhost:8080/api/v1/backup/restore \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "backup_file": "backup-2026-05-30-020000.zip"
  }'
```

### Restore Options

**Partial restore**:
```bash
./cloud-vault backup restore \
  --backup backup.zip \
  --only-database
```

**Dry run** (preview):
```bash
./cloud-vault backup restore \
  --backup backup.zip \
  --dry-run
```

## Export Data

### Export Documents

**Via Web Interface**:
1. Navigate to **Documents** page
2. Select documents
3. Click **Export** → **Markdown**
4. Download ZIP file

**Via CLI**:
```bash
./cloud-vault export \
  --format markdown \
  --output /path/to/export.zip
```

### Export Formats

- **Markdown**: Raw Markdown files
- **JSON**: Structured data with metadata
- **CSV**: Spreadsheet format

## Backup Best Practices

### 1. Regular Backups

Enable automatic backups:
```toml
[backup]
enabled = true
schedule = "0 2 * * *"  # Daily
```

### 2. Offsite Storage

Store backups offsite:
```bash
# Sync to remote server
rsync -av data/backups/ user@backup-server:/backups/

# Upload to cloud storage
aws s3 sync data/backups/ s3://my-backups/cloud-vault/
```

### 3. Test Restores

Regularly test restore process:
```bash
# Create test environment
mkdir test-restore
cd test-restore

# Restore backup
cloud-vault backup restore \
  --backup ../backup.zip

# Verify data
cloud-vault documents list
```

### 4. Monitor Backup Size

Check backup sizes:
```bash
du -sh data/backups/
```

If backups grow too large:
- Reduce retention days
- Exclude large attachments
- Use incremental backups

### 5. Encrypt Backups

Encrypt sensitive backups:
```bash
# Encrypt backup
gpg -c backup.zip

# Decrypt backup
gpg backup.zip.gpg
```

## Backup Verification

### Verify Backup Integrity

```bash
./cloud-vault backup verify \
  --backup backup.zip
```

### Check Database Integrity

```bash
./cloud-vault doctor check-database
```

## Troubleshooting

### Backup Fails

**Error**: `Permission denied`
```bash
# Check permissions
ls -la data/

# Fix permissions
chmod 755 data/
```

**Error**: `Disk full`
```bash
# Check disk space
df -h

# Clean old backups
./cloud-vault backup prune --keep 5
```

### Restore Fails

**Error**: `Database locked`
```bash
# Stop Cloud Vault
sudo systemctl stop cloud-vault

# Restore
./cloud-vault backup restore --backup backup.zip

# Start Cloud Vault
sudo systemctl start cloud-vault
```

**Error**: `Backup corrupted`
```bash
# Verify backup
./cloud-vault backup verify --backup backup.zip

# Try previous backup
./cloud-vault backup restore --backup backup-previous.zip
```

## Next Steps

- [Security](./security.md) - Secure your backups
- [Troubleshooting](./troubleshooting.md) - Common issues
- [Diagnostics](./diagnostics.md) - Export diagnostic data

---

**Last updated**: 2026-05-30
