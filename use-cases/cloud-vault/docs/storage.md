# Storage Configuration Guide

This guide covers storage configuration options for Markdown Cloud Vault, including local filesystem and Qiniu cloud storage.

## Overview

Cloud Vault supports two storage backends:
- **Local Storage**: Files stored on the local filesystem (default)
- **Qiniu Cloud Storage**: Files stored in Qiniu cloud buckets

## Local Storage

### Configuration

Local storage is the default and simplest option:

```toml
[storage]
provider = "local"
root = "./data/objects"
```

### Directory Structure

```
data/objects/
├── documents/
│   ├── 2026/
│   │   ├── 01/
│   │   │   ├── abc123.md
│   │   │   └── def456.md
│   │   └── 02/
│   └── ...
├── attachments/
│   ├── images/
│   └── files/
└── temp/
```

### Advantages

- **Fast**: Direct filesystem access
- **Simple**: No external dependencies
- **Private**: Data stays on your machine
- **Backup-friendly**: Easy to backup entire directory

### Limitations

- **Single server**: Not suitable for multi-server deployments
- **Disk space**: Limited by available disk space
- **No redundancy**: Requires manual backup strategy

### Best Practices

1. **Use dedicated storage volume**:
```bash
# Create dedicated partition
sudo mkfs.ext4 /dev/sdb1
sudo mount /dev/sdb1 /var/lib/cloud-vault/data/objects
```

2. **Monitor disk space**:
```bash
df -h /var/lib/cloud-vault/data/objects
```

3. **Regular backups**:
```bash
# Daily backup
tar -czf /backups/objects-$(date +%Y%m%d).tar.gz /var/lib/cloud-vault/data/objects
```

4. **Filesystem optimization**:
```bash
# Use ext4 with optimized settings
sudo tune2fs -o journal_data_writeback /dev/sdb1
```

## Qiniu Cloud Storage

### Overview

Qiniu provides object storage with CDN delivery, suitable for:
- Multi-server deployments
- Large document collections
- Global content delivery
- Automatic redundancy

### Prerequisites

1. **Qiniu account**: Sign up at [qiniu.com](https://www.qiniu.com/)
2. **Create bucket**: Create a storage bucket in Qiniu console
3. **Get credentials**: Obtain Access Key and Secret Key
4. **Configure domain**: Set up custom domain or use Qiniu CDN

### Configuration

```toml
[storage]
provider = "qiniu"

[storage.qiniu]
access_key = "your-access-key"
secret_key = "your-secret-key"
bucket = "your-bucket-name"
domain = "https://cdn.yourdomain.com"
region = "z0"  # z0=华东, z1=华北, z2=华南, na0=北美, as0=东南亚
```

### Qiniu Regions

| Region Code | Location | Use Case |
|-------------|----------|----------|
| `z0` | 华东 (East China) | Default, best for China mainland |
| `z1` | 华北 (North China) | Northern China users |
| `z2` | 华南 (South China) | Southern China users |
| `na0` | 北美 (North America) | US/Canada users |
| `as0` | 东南亚 (Southeast Asia) | Singapore/SEA users |

### Bucket Configuration

**In Qiniu Console**:

1. **Create bucket**:
   - Name: `cloud-vault-docs`
   - Region: Choose closest to your users
   - Access: Private (recommended)

2. **Configure CORS**:
   - Allowed origins: `https://vault.yourdomain.com`
   - Allowed methods: `GET, POST, PUT, DELETE`
   - Allowed headers: `*`

3. **Set up CDN domain**:
   - Add custom domain: `cdn.yourdomain.com`
   - Configure SSL certificate
   - Enable HTTPS

### Security Best Practices

1. **Use private bucket**:
   - Set bucket access to "Private"
   - Use signed URLs for access
   - Never expose Access Key/Secret Key

2. **Environment variables**:
```bash
export QINIU_ACCESS_KEY="your-access-key"
export QINIU_SECRET_KEY="your-secret-key"
```

Update `config.toml`:
```toml
[storage.qiniu]
access_key = "${QINIU_ACCESS_KEY}"
secret_key = "${QINIU_SECRET_KEY}"
```

3. **IP whitelist**:
   - Restrict bucket access to server IPs
   - Configure in Qiniu console

4. **Referer whitelist**:
   - Restrict CDN access to your domain
   - Prevent hotlinking

### URL Signing

Cloud Vault automatically signs URLs for Qiniu resources:

```go
// Signed URL format
https://cdn.yourdomain.com/documents/2026/01/abc123.md?e=1234567890&token=xxx

// Token expires after configured TTL
```

Configure token expiration:
```toml
[storage.qiniu]
token_expires_seconds = 3600  # 1 hour
```

### Migration from Local to Qiniu

To migrate existing documents from local storage to Qiniu:

1. **Stop Cloud Vault**:
```bash
sudo systemctl stop cloud-vault
```

2. **Update configuration**:
```toml
[storage]
provider = "qiniu"
# ... qiniu config ...
```

3. **Run migration tool**:
```bash
./cloud-vault migrate-storage \
  --from local \
  --to qiniu \
  --source /var/lib/cloud-vault/data/objects \
  --dry-run  # Test first
```

4. **Execute migration**:
```bash
./cloud-vault migrate-storage \
  --from local \
  --to qiniu \
  --source /var/lib/cloud-vault/data/objects
```

5. **Restart Cloud Vault**:
```bash
sudo systemctl start cloud-vault
```

### Migration from Qiniu to Local

```bash
./cloud-vault migrate-storage \
  --from qiniu \
  --to local \
  --dest /var/lib/cloud-vault/data/objects
```

## Storage Comparison

| Feature | Local | Qiniu |
|---------|-------|-------|
| **Setup complexity** | Simple | Moderate |
| **Cost** | Free (disk space) | Pay per GB/transfer |
| **Performance** | Fast (local) | Fast (CDN) |
| **Scalability** | Limited | Unlimited |
| **Redundancy** | Manual | Automatic |
| **Backup** | Manual | Built-in |
| **Multi-server** | No | Yes |
| **Global access** | No | Yes (CDN) |

## Advanced Configuration

### Storage Quotas

Limit storage usage per user or collection:

```toml
[storage]
max_total_size = "100GB"
max_document_size = "50MB"
max_attachment_size = "20MB"
```

### File Deduplication

Enable content-based deduplication:

```toml
[storage]
enable_deduplication = true
deduplication_algorithm = "sha256"
```

### Storage Tiers

Use different storage for different content types:

```toml
[storage.tiers]
documents = "local"      # Fast access
attachments = "qiniu"    # CDN delivery
backups = "qiniu"        # Long-term storage
```

### Compression

Enable automatic compression for large files:

```toml
[storage]
enable_compression = true
compression_algorithm = "gzip"
compression_threshold = "1MB"
```

## Backup Strategies

### Local Storage Backup

**Full backup**:
```bash
# Daily full backup
tar -czf /backups/storage-full-$(date +%Y%m%d).tar.gz \
  /var/lib/cloud-vault/data/objects

# Keep last 7 days
find /backups -name "storage-full-*.tar.gz" -mtime +7 -delete
```

**Incremental backup**:
```bash
# Use rsync for incremental
rsync -av --delete \
  /var/lib/cloud-vault/data/objects/ \
  /backups/storage-mirror/
```

**Offsite backup**:
```bash
# Sync to remote server
rsync -avz -e ssh \
  /var/lib/cloud-vault/data/objects/ \
  user@backup-server:/backups/cloud-vault/
```

### Qiniu Backup

**Cross-region replication**:
- Enable in Qiniu console
- Automatically replicates to another region

**Download backup**:
```bash
# Download all objects
./cloud-vault backup-qiniu \
  --bucket your-bucket \
  --output /backups/qiniu-backup/
```

## Monitoring Storage

### Disk Usage

```bash
# Check disk usage
du -sh /var/lib/cloud-vault/data/objects

# Check by subdirectory
du -sh /var/lib/cloud-vault/data/objects/documents
du -sh /var/lib/cloud-vault/data/objects/attachments
```

### Qiniu Usage

**API query**:
```bash
curl -X GET \
  "https://api.qiniu.com/v2/buckets/your-bucket/stat" \
  -H "Authorization: Bearer $QINIU_TOKEN"
```

**Dashboard**:
- Login to Qiniu console
- View storage usage, bandwidth, requests

### Alerts

Set up storage alerts:

```toml
[monitoring]
storage_alert_threshold = 80  # Alert at 80% capacity
storage_alert_email = "admin@example.com"
```

## Performance Optimization

### Local Storage

**Use SSD**:
```bash
# Check disk type
lsblk -d -o name,rota
# rota=0 means SSD
```

**Optimize filesystem**:
```bash
# Mount with noatime
/dev/sdb1 /var/lib/cloud-vault/data/objects ext4 noatime,nodiratime 0 0
```

**Increase cache**:
```bash
# Increase page cache
echo 3 > /proc/sys/vm/drop_caches
```

### Qiniu Storage

**Enable CDN**:
- Configure CDN in Qiniu console
- Use custom domain with CDN

**Optimize image delivery**:
```toml
[storage.qiniu]
image_processing = true
image_quality = 85
image_format = "webp"
```

**Prefetch popular content**:
```toml
[storage.qiniu]
prefetch_enabled = true
prefetch_threshold = 100  # Views per day
```

## Troubleshooting

### Local Storage Issues

**Permission denied**:
```bash
sudo chown -R cloud-vault:cloud-vault /var/lib/cloud-vault/data/objects
sudo chmod -R 755 /var/lib/cloud-vault/data/objects
```

**Disk full**:
```bash
# Check usage
df -h /var/lib/cloud-vault/data/objects

# Clean up temp files
rm -rf /var/lib/cloud-vault/data/objects/temp/*

# Expand disk or move to larger volume
```

**File corruption**:
```bash
# Check filesystem
sudo fsck /dev/sdb1

# Restore from backup
tar -xzf /backups/storage-full-20260530.tar.gz -C /
```

### Qiniu Storage Issues

**Authentication failed**:
- Verify Access Key and Secret Key
- Check if keys are active in Qiniu console
- Ensure no IP restrictions blocking server

**Upload failed**:
```bash
# Check network connectivity
curl -I https://up.qiniup.com

# Check bucket permissions
# Ensure bucket allows uploads from your IP
```

**CDN not working**:
- Verify domain configuration in Qiniu console
- Check DNS records point to Qiniu CDN
- Test with direct bucket URL

**Slow downloads**:
- Enable CDN if not already enabled
- Choose region closer to users
- Enable image processing for optimization

## Migration Checklist

When migrating storage providers:

- [ ] Backup current storage
- [ ] Update configuration file
- [ ] Test new storage connection
- [ ] Run migration tool with `--dry-run`
- [ ] Execute migration
- [ ] Verify all documents accessible
- [ ] Update backup scripts
- [ ] Monitor for errors
- [ ] Remove old storage after verification

## Next Steps

- [Qiniu Setup Guide](./qiniu.md) - Detailed Qiniu configuration
- [Backup & Restore](./backup-restore.md) - Backup strategies
- [Security](./security.md) - Storage security best practices

---

**Last updated**: 2026-05-30
