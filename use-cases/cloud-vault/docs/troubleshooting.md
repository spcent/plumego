# Troubleshooting Guide

This guide helps you diagnose and resolve common issues with Markdown Cloud Vault.

## Quick Diagnostics

Before troubleshooting, run health checks:

```bash
./cloud-vault doctor check
./cloud-vault logs tail --level error
```

## Installation Issues

### Desktop App Won't Start

**Windows**:

**Symptom**: App doesn't launch or crashes immediately

**Solutions**:
1. Check Windows Event Viewer for error details
2. Run as administrator (right-click → Run as administrator)
3. Install Visual C++ Redistributable 2015-2022
4. Reinstall application

**macOS**:

**Symptom**: "App is damaged" or won't open

**Solutions**:
```bash
# Remove quarantine attribute
sudo xattr -rd com.apple.quarantine /Applications/Cloud\ Vault.app

# Reset preferences
defaults delete com.cloudvault.app
```

**Linux**:

**Symptom**: AppImage won't run

**Solutions**:
```bash
# Make executable
chmod +x cloud-vault-v1.0.0-linux-amd64.AppImage

# Extract if FUSE error
./cloud-vault-v1.0.0-linux-amd64.AppImage --appimage-extract
./squashfs-root/AppRun
```

### Server Won't Start

**Error**: `Port 8080 already in use`

**Solutions**:
```bash
# Find process using port
lsof -i :8080

# Kill process
kill -9 <PID>

# Or use different port
./cloud-vault --addr :8081
```

**Error**: `Permission denied`

**Solutions**:
```bash
# Check file permissions
ls -la /opt/cloud-vault/

# Fix permissions
sudo chown -R cloud-vault:cloud-vault /opt/cloud-vault
sudo chmod -R 755 /opt/cloud-vault
```

**Error**: `Database locked`

**Solutions**:
```bash
# Stop all Cloud Vault instances
sudo systemctl stop cloud-vault

# Check for processes
lsof /var/lib/cloud-vault/data/app.db

# Restart
sudo systemctl start cloud-vault
```

## Database Issues

### Database Corruption

**Symptoms**:
- Unexpected errors
- Missing documents
- Search not working

**Solutions**:
```bash
# Check database integrity
sqlite3 /var/lib/cloud-vault/data/app.db "PRAGMA integrity_check;"

# If corrupted, stop the server and restore from backup
./cloud-vault restore --file backup.zip --data-dir /var/lib/cloud-vault/data

# Or rebuild database
./cloud-vault database rebuild
```

### Database Performance

**Symptoms**:
- Slow queries
- High CPU usage
- Long response times

**Solutions**:
```bash
# Optimize database
sqlite3 /var/lib/cloud-vault/data/app.db "VACUUM;"

# Analyze query performance
sqlite3 /var/lib/cloud-vault/data/app.db "ANALYZE;"

# Rebuild indexes
./cloud-vault database reindex
```

### Database Migration Errors

**Error**: `Migration failed`

**Solutions**:
```bash
# Check current version
./cloud-vault database version

# Backup database
cp /var/lib/cloud-vault/data/app.db /var/lib/cloud-vault/data/app.db.backup

# Force migration
./cloud-vault database migrate --force

# If fails, restore backup
cp /var/lib/cloud-vault/data/app.db.backup /var/lib/cloud-vault/data/app.db
```

## Storage Issues

### Local Storage Full

**Symptoms**:
- Upload failures
- Import failures
- "Disk full" errors

**Solutions**:
```bash
# Check disk space
df -h /var/lib/cloud-vault/data

# Clean old backups through the UI or DELETE /api/v1/system/backups/:name

# Remove orphaned files
./cloud-vault storage cleanup

# Expand disk or move to larger volume
```

### Qiniu Storage Errors

**Error**: `401 Unauthorized`

**Solutions**:
1. Verify Access Key and Secret Key
2. Check if keys are active in Qiniu console
3. Ensure no IP restrictions blocking server
4. Regenerate keys if compromised

**Error**: `Upload failed: connection timeout`

**Solutions**:
```bash
# Check network connectivity
ping up.qiniup.com
curl -I https://up.qiniup.com

# Increase timeout
[storage.qiniu]
upload_timeout_seconds = 300

# Check firewall
sudo ufw allow out 443/tcp
```

**Error**: `403 Forbidden`

**Solutions**:
1. Verify bucket permissions
2. Check Referer whitelist
3. Ensure SSL certificate is valid
4. Test with direct bucket URL

## Search Issues

### Search Not Working

**Symptoms**:
- No search results
- Search returns wrong documents
- Search is slow

**Solutions**:
```bash
# Check search index status
./cloud-vault search status

# Rebuild search index
./cloud-vault search reindex

# Optimize index
./cloud-vault search optimize

# Check logs
./cloud-vault logs tail --level error | grep search
```

### Index Out of Sync

**Symptoms**:
- Documents not appearing in search
- Deleted documents still searchable

**Solutions**:
```bash
# Full reindex
./cloud-vault search reindex --full

# Verify index
./cloud-vault search verify

# Check for errors
./cloud-vault logs tail --level error | grep index
```

### Slow Search

**Symptoms**:
- Search takes > 1 second
- High CPU during search

**Solutions**:
```toml
[search]
cache_enabled = true
cache_ttl = 300
batch_size = 200
```

```bash
# Optimize index
./cloud-vault search optimize

# Check database performance
./cloud-vault metrics database
```

## Import Issues

### Import Fails

**Error**: `Import failed: permission denied`

**Solutions**:
```bash
# Check file permissions
ls -l /path/to/files

# Fix permissions
chmod -R 644 /path/to/files
chown -R cloud-vault:cloud-vault /path/to/files
```

**Error**: `Out of memory during import`

**Solutions**:
```toml
[import]
batch_size = 50
max_file_size = 5242880  # 5 MB
workers = 2
```

```bash
# Import in smaller batches
./cloud-vault import \
  --source /path/to/files \
  --batch-size 50
```

**Error**: `Invalid UTF-8 encoding`

**Solutions**:
```toml
[import.transform]
fix_encoding = true
fallback_encoding = "latin1"
```

### Duplicate Detection Issues

**Symptoms**:
- Duplicates not detected
- False positives

**Solutions**:
```toml
[import]
duplicate_detection = true
duplicate_method = "content_hash"
duplicate_threshold = 0.95
```

```bash
# Manual duplicate check
./cloud-vault organize detect-duplicates
```

## AI Issues

### AI Features Not Working

**Error**: `401 Unauthorized`

**Solutions**:
1. Verify API key is valid
2. Check API key has required permissions
3. Verify API endpoint URL
4. Check account balance/credits

**Error**: `429 Too Many Requests`

**Solutions**:
```toml
[ai]
rate_limit = 30  # Reduce rate
cache_enabled = true
cache_ttl = 3600
```

```bash
# Check usage
./cloud-vault ai stats
```

**Error**: `Timeout`

**Solutions**:
```toml
[ai]
timeout = 120
max_tokens = 2048
```

### Poor AI Quality

**Symptoms**:
- Summaries missing key points
- Inaccurate Q&A answers
- Irrelevant tags

**Solutions**:
1. Use better model (gpt-4 instead of gpt-3.5)
2. Provide more context
3. Adjust temperature (lower = more focused)
4. Break into smaller tasks

## Performance Issues

### High Memory Usage

**Symptoms**:
- Memory usage > 2 GB
- Out of memory errors

**Solutions**:
```toml
[search]
cache_max_size = 500  # Reduce cache

[ai]
cache_enabled = false  # Disable AI cache

[import]
batch_size = 50  # Smaller batches
```

```bash
# Restart service
sudo systemctl restart cloud-vault

# Check for memory leaks
./cloud-vault metrics
```

### High CPU Usage

**Symptoms**:
- CPU usage > 80%
- Slow response times

**Solutions**:
```toml
[import]
workers = 2  # Reduce workers

[ai]
task_workers = 1  # Reduce AI workers

[search]
index_on_save = false  # Disable auto-index
```

```bash
# Check what's using CPU
./cloud-vault metrics
./cloud-vault logs tail --level info | grep -E "(import|search|ai)"
```

### Slow Response Times

**Symptoms**:
- API responses > 1 second
- UI feels sluggish

**Solutions**:
```bash
# Check API performance
./cloud-vault metrics api

# Optimize database
./cloud-vault database optimize

# Enable caching
[search]
cache_enabled = true

[ai]
cache_enabled = true
```

## Network Issues

### Connection Refused

**Error**: `Connection refused`

**Solutions**:
```bash
# Check if service is running
sudo systemctl status cloud-vault

# Check if port is listening
netstat -tlnp | grep 8080

# Check firewall
sudo ufw status
sudo ufw allow 8080/tcp
```

### SSL/TLS Errors

**Error**: `Certificate verify failed`

**Solutions**:
```bash
# Check certificate
openssl x509 -in /path/to/cert.pem -text -noout

# Verify certificate chain
openssl verify -CAfile /path/to/ca.pem /path/to/cert.pem

# Update CA certificates
sudo update-ca-certificates
```

## Authentication Issues

### Login Fails

**Error**: `Invalid credentials`

**Solutions**:
1. Check username and password
2. Verify account is not locked
3. Reset password:
```bash
./cloud-vault user reset-password --username admin
```

**Error**: `Account locked`

**Solutions**:
```bash
# Wait for lockout period (15 minutes)
# Or unlock immediately
./cloud-vault user unlock --username admin
```

### Session Issues

**Error**: `Session expired`

**Solutions**:
1. Login again
2. Check session TTL:
```toml
[auth]
session_ttl = 86400  # 24 hours
```

3. Verify session secret hasn't changed

## Recovery Procedures

### Complete System Recovery

If all else fails:

1. **Stop service**:
```bash
sudo systemctl stop cloud-vault
```

2. **Backup current state**:
```bash
tar -czf recovery-backup-$(date +%Y%m%d).tar.gz \
  /var/lib/cloud-vault
```

3. **Restore from backup**:
```bash
./cloud-vault restore --file backup.zip --data-dir /var/lib/cloud-vault/data
```

4. **Rebuild indexes**:
```bash
./cloud-vault search reindex --full
./cloud-vault database reindex
```

5. **Start service**:
```bash
sudo systemctl start cloud-vault
```

6. **Verify**:
```bash
./cloud-vault doctor check
```

### Emergency Contact

If you cannot resolve the issue:

1. Generate diagnostic bundle:
```bash
./cloud-vault diagnostics export --output emergency-diagnostic.zip
```

2. Collect logs:
```bash
tar -czf logs-$(date +%Y%m%d).tar.gz \
  /var/log/cloud-vault/
```

3. Document steps to reproduce

4. Contact support with diagnostic bundle

## Next Steps

- [Diagnostics](./diagnostics.md) - Export diagnostic data
- [Security](./security.md) - Security best practices
- [Backup & Restore](./backup-restore.md) - Backup your data

---

**Last updated**: 2026-05-30
