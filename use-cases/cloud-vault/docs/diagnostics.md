# Diagnostics Guide

This guide covers diagnostic tools for troubleshooting Markdown Cloud Vault issues.

## Overview

Cloud Vault provides diagnostic tools to help identify and resolve issues:
- **Diagnostic bundles**: Export system information for troubleshooting
- **Health checks**: Verify system components
- **Logs**: Application and access logs
- **Metrics**: Performance and usage metrics

## Diagnostic Bundles

### Generate Diagnostic Bundle

**Via Web Interface**:
1. Navigate to **Settings** → **Diagnostics**
2. Click **Generate Diagnostic Bundle**
3. Wait for bundle to be created
4. Download the ZIP file

**Via CLI**:
```bash
./cloud-vault diagnostics export \
  --output /path/to/diagnostic.zip
```

**Via API**:
```bash
curl -X POST http://localhost:8080/api/v1/diagnostics/generate \
  -H "Authorization: Bearer $TOKEN"
```

### Bundle Contents

Each diagnostic bundle includes:
- `system-info.json` - System information (OS, memory, disk)
- `config-redacted.json` - Configuration (sensitive values redacted)
- `database-stats.json` - Database statistics
- `logs.txt` - Recent application logs
- `health-check.json` - Health check results
- `build-info.json` - Version and build information

**Sensitive data is automatically redacted**:
- API keys
- Passwords
- Tokens
- Secret keys

### View Diagnostic Bundles

**Via Web Interface**:
1. Navigate to **Settings** → **Diagnostics**
2. View list of generated bundles
3. Click to download or delete

**Via CLI**:
```bash
./cloud-vault diagnostics list
```

**Via API**:
```bash
curl http://localhost:8080/api/v1/diagnostics \
  -H "Authorization: Bearer $TOKEN"
```

### Download Diagnostic Bundle

**Via CLI**:
```bash
./cloud-vault diagnostics download \
  --id 123 \
  --output /path/to/diagnostic.zip
```

**Via API**:
```bash
curl http://localhost:8080/api/v1/diagnostics/123/download \
  -H "Authorization: Bearer $TOKEN" \
  --output diagnostic.zip
```

## Health Checks

### Run Health Check

**Via Web Interface**:
1. Navigate to **System** page
2. Click **Run Health Check**
3. View results

**Via CLI**:
```bash
./cloud-vault doctor check
```

Output:
```
Cloud Vault Health Check
========================

✓ Database: Connected (SQLite)
✓ Storage: Accessible (local)
✓ Search Index: Healthy (1234 documents)
✓ Configuration: Valid
✓ Permissions: OK
✓ Disk Space: 45.2 GB available

Status: Healthy
```

**Via API**:
```bash
curl http://localhost:8080/api/v1/health \
  -H "Authorization: Bearer $TOKEN"
```

Response:
```json
{
  "status": "healthy",
  "components": {
    "database": {
      "status": "healthy",
      "message": "Connected"
    },
    "storage": {
      "status": "healthy",
      "message": "Accessible"
    },
    "search": {
      "status": "healthy",
      "message": "1234 documents indexed"
    }
  }
}
```

### Specific Health Checks

**Database**:
```bash
./cloud-vault doctor check-database
```

**Storage**:
```bash
./cloud-vault doctor check-storage
```

**Search Index**:
```bash
./cloud-vault doctor check-search
```

**Configuration**:
```bash
./cloud-vault doctor check-config
```

## Logs

### View Logs

**Via Web Interface**:
1. Navigate to **Settings** → **Diagnostics**
2. Click **View Logs**
3. Filter by level (error, warn, info)

**Via CLI**:
```bash
# View recent logs
./cloud-vault logs tail --lines 100

# View logs with filter
./cloud-vault logs tail --level error

# View logs from file
./cloud-vault logs view /var/log/cloud-vault/app.log
```

**Via API**:
```bash
curl http://localhost:8080/api/v1/logs?level=error&limit=100 \
  -H "Authorization: Bearer $TOKEN"
```

### Log Levels

- **error**: Critical errors requiring attention
- **warn**: Warnings that may indicate issues
- **info**: General information
- **debug**: Detailed debugging information

### Log Rotation

Logs are automatically rotated:
```toml
[logging]
max_size_mb = 100
max_files = 10
compress = true
```

## Metrics

### View Metrics

**Via CLI**:
```bash
./cloud-vault metrics
```

Output:
```
Cloud Vault Metrics
===================

Documents: 1234
Collections: 45
Tags: 234
Search Index: 1234 documents

Database Size: 45.2 MB
Storage Used: 1.2 GB

API Requests (24h): 12345
API Errors (24h): 23
Average Response Time: 45ms
```

**Via API**:
```bash
curl http://localhost:8080/api/v1/metrics \
  -H "Authorization: Bearer $TOKEN"
```

### Performance Metrics

**Database queries**:
```bash
./cloud-vault metrics database
```

**Search performance**:
```bash
./cloud-vault metrics search
```

**API performance**:
```bash
./cloud-vault metrics api
```

## Troubleshooting with Diagnostics

### Common Issues

**Database connection issues**:
```bash
./cloud-vault doctor check-database
./cloud-vault logs tail --level error | grep database
```

**Storage access issues**:
```bash
./cloud-vault doctor check-storage
./cloud-vault logs tail --level error | grep storage
```

**Search index issues**:
```bash
./cloud-vault doctor check-search
./cloud-vault search reindex
```

**Performance issues**:
```bash
./cloud-vault metrics
./cloud-vault logs tail --level warn | grep slow
```

### Collecting Diagnostic Information

When reporting issues, provide:
1. Diagnostic bundle
2. Steps to reproduce
3. Expected vs actual behavior
4. Screenshots (if applicable)

## Security and Privacy

### Data Protection

Diagnostic bundles automatically redact:
- API keys
- Passwords
- Tokens
- Secret keys
- Database credentials

### Sharing Diagnostics

Before sharing diagnostic bundles:
1. Review bundle contents
2. Verify sensitive data is redacted
3. Share only with trusted parties

### Local Storage

Diagnostic bundles are stored locally:
```
data/diagnostics/
├── diagnostic-2026-05-30-120000.zip
├── diagnostic-2026-05-29-120000.zip
└── ...
```

## Best Practices

### 1. Regular Health Checks

Run health checks periodically:
```bash
# Add to cron
0 */6 * * * /usr/local/bin/cloud-vault doctor check > /var/log/cloud-vault/health.log
```

### 2. Monitor Logs

Set up log monitoring:
```bash
# Watch for errors
tail -f /var/log/cloud-vault/app.log | grep ERROR
```

### 3. Export Before Updates

Generate diagnostic bundle before updates:
```bash
./cloud-vault diagnostics export \
  --output pre-update-diagnostic.zip
```

### 4. Review Metrics

Check metrics regularly:
```bash
./cloud-vault metrics
```

Look for:
- Increasing error rates
- Slow response times
- Resource usage trends

### 5. Automate Diagnostics

Create diagnostic script:
```bash
#!/bin/bash
# daily-diagnostic.sh

DATE=$(date +%Y-%m-%d)
OUTPUT="/backups/diagnostics/diagnostic-$DATE.zip"

cloud-vault diagnostics export --output "$OUTPUT"
cloud-vault doctor check > "/backups/diagnostics/health-$DATE.log"
cloud-vault metrics > "/backups/diagnostics/metrics-$DATE.log"

# Keep last 30 days
find /backups/diagnostics -mtime +30 -delete
```

## Next Steps

- [Troubleshooting](./troubleshooting.md) - Common issues and solutions
- [Security](./security.md) - Security best practices
- [Backup & Restore](./backup-restore.md) - Backup your data

---

**Last updated**: 2026-05-30
