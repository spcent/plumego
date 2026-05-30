# Qiniu Cloud Storage Setup Guide

This guide walks you through setting up Qiniu cloud storage for Markdown Cloud Vault.

## Overview

Qiniu Cloud provides object storage with CDN delivery, ideal for:
- Multi-server deployments
- Large document collections
- Global content delivery
- Automatic redundancy and backup

## Prerequisites

- Qiniu Cloud account
- Domain name (optional, for custom CDN)
- SSL certificate (optional, for HTTPS)

## Step 1: Create Qiniu Account

1. Visit [qiniu.com](https://www.qiniu.com/)
2. Click "Register" (注册)
3. Complete registration with email or phone
4. Verify your account
5. Login to Qiniu Console

## Step 2: Create Storage Bucket

1. Navigate to **Object Storage** → **Bucket Management**
2. Click **Create Bucket**
3. Fill in details:
   - **Bucket Name**: `cloud-vault-docs` (must be unique)
   - **Region**: Choose closest to your users
     - `z0` - 华东 (East China) - Default
     - `z1` - 华北 (North China)
     - `z2` - 华南 (South China)
     - `na0` - 北美 (North America)
     - `as0` - 东南亚 (Southeast Asia)
   - **Access Control**: **Private** (recommended)
   - **Storage Type**: Standard Storage
4. Click **Create**

## Step 3: Get Access Credentials

1. Go to **Personal Center** → **Key Management**
2. Click **Create Key**
3. Note down:
   - **Access Key (AK)**: Public identifier
   - **Secret Key (SK)**: Private key (keep secure!)
4. Download or copy the keys

**Security Note**: Never commit keys to version control or expose them publicly.

## Step 4: Configure Bucket Permissions

### CORS Configuration

1. Select your bucket
2. Go to **Bucket Settings** → **CORS**
3. Click **Add Rule**:
   - **Allowed Origins**: `https://vault.yourdomain.com`
   - **Allowed Methods**: `GET, POST, PUT, DELETE, HEAD`
   - **Allowed Headers**: `*`
   - **Exposed Headers**: `ETag, x-qn-meta-*`
   - **Max Age**: `3600`
4. Click **Save**

### Referer Whitelist

1. Go to **Bucket Settings** → **Hotlink Protection**
2. Enable **Referer Whitelist**
3. Add allowed domains:
   - `*.yourdomain.com`
   - `vault.yourdomain.com`
4. Enable **Allow Empty Referer**: No
5. Click **Save**

### IP Whitelist (Optional)

1. Go to **Bucket Settings** → **IP Access Control**
2. Enable **IP Whitelist**
3. Add your server IPs:
   - `123.45.67.89` (your server IP)
4. Click **Save**

## Step 5: Configure CDN Domain

### Option A: Use Qiniu Default Domain

Qiniu provides a default domain:
```
http://<bucket>.qbox.me
```

**Note**: Default domain is for testing only, not production.

### Option B: Custom CDN Domain (Recommended)

1. Go to **CDN** → **Domain Management**
2. Click **Add Domain**
3. Fill in:
   - **Domain**: `cdn.yourdomain.com`
   - **Origin**: Your bucket name
   - **Origin Protocol**: Follow
   - **HTTPS**: Enable (if you have SSL cert)
4. Click **Create**

### DNS Configuration

Add CNAME record in your DNS:

```
Type: CNAME
Name: cdn
Value: cdn.yourdomain.com.qiniudns.com
TTL: 600
```

Verify DNS propagation:
```bash
nslookup cdn.yourdomain.com
# Should return Qiniu CDN servers
```

### SSL Certificate

**Option 1: Upload existing certificate**
1. Go to **CDN** → **Certificate Management**
2. Click **Upload Certificate**
3. Upload:
   - Certificate file (PEM format)
   - Private key file
4. Bind to your CDN domain

**Option 2: Use Qiniu free certificate**
1. Click **Apply for Free Certificate**
2. Enter domain: `cdn.yourdomain.com`
3. Complete DNS verification
4. Wait for approval (usually 1-2 hours)

## Step 6: Configure Cloud Vault

### Update config.toml

```toml
[storage]
provider = "qiniu"

[storage.qiniu]
access_key = "your-access-key"
secret_key = "your-secret-key"
bucket = "cloud-vault-docs"
domain = "https://cdn.yourdomain.com"
region = "z0"
token_expires_seconds = 3600
```

### Use Environment Variables (Recommended)

Create `.env` file:
```bash
QINIU_ACCESS_KEY=your-access-key
QINIU_SECRET_KEY=your-secret-key
QINIU_BUCKET=cloud-vault-docs
QINIU_DOMAIN=https://cdn.yourdomain.com
QINIU_REGION=z0
```

Update `config.toml`:
```toml
[storage]
provider = "qiniu"

[storage.qiniu]
access_key = "${QINIU_ACCESS_KEY}"
secret_key = "${QINIU_SECRET_KEY}"
bucket = "${QINIU_BUCKET}"
domain = "${QINIU_DOMAIN}"
region = "${QINIU_REGION}"
```

## Step 7: Test Configuration

1. **Restart Cloud Vault**:
```bash
sudo systemctl restart cloud-vault
```

2. **Check logs**:
```bash
sudo journalctl -u cloud-vault -f
```

3. **Upload test document**:
- Create a new document in Cloud Vault
- Check if it appears in Qiniu bucket
- Verify CDN URL is accessible

4. **Test download**:
```bash
curl -I https://cdn.yourdomain.com/documents/test.md
# Should return 200 OK
```

## Step 8: Migrate Existing Documents

If you have documents in local storage:

```bash
# Dry run (test first)
./cloud-vault migrate-storage \
  --from local \
  --to qiniu \
  --source /var/lib/cloud-vault/data/objects \
  --dry-run

# Execute migration
./cloud-vault migrate-storage \
  --from local \
  --to qiniu \
  --source /var/lib/cloud-vault/data/objects

# Verify migration
./cloud-vault verify-storage --provider qiniu
```

## Qiniu Pricing

### Storage Costs

| Region | Standard Storage | Infrequent Access | Archive |
|--------|------------------|-------------------|---------|
| 华东/华北/华南 | ¥0.118/GB/month | ¥0.06/GB/month | ¥0.02/GB/month |
| 北美/东南亚 | ¥0.14/GB/month | ¥0.08/GB/month | ¥0.03/GB/month |

### Traffic Costs

| Type | Price |
|------|-------|
| CDN Traffic (China) | ¥0.21/GB |
| CDN Traffic (Overseas) | ¥0.49/GB |
| Origin Pull | ¥0.15/GB |

### API Requests

| Operation | Price |
|-----------|-------|
| PUT/POST/DELETE | ¥0.01/10,000 requests |
| GET/HEAD | ¥0.01/100,000 requests |

**Cost Estimation Example**:
- 100 GB storage: ¥11.8/month
- 1 TB CDN traffic: ¥210/month
- 1M API requests: ¥0.10/month
- **Total**: ~¥222/month (~$30 USD)

## Security Best Practices

### 1. Protect Access Keys

**Never expose keys**:
```bash
# ❌ Don't commit to git
git add config.toml  # Contains keys

# ✅ Use environment variables
git add .env.example  # Template only
```

**Rotate keys regularly**:
1. Create new key in Qiniu console
2. Update Cloud Vault configuration
3. Restart service
4. Delete old key after verification

### 2. Use Private Bucket

- Set bucket to **Private** access
- Cloud Vault generates signed URLs automatically
- URLs expire after configured TTL

### 3. Enable Hotlink Protection

- Configure Referer whitelist
- Prevent unauthorized embedding
- Reduce bandwidth theft

### 4. Monitor Access Logs

Enable access logging:
1. Go to **Bucket Settings** → **Log Management**
2. Enable **Access Log**
3. Logs stored in separate bucket
4. Analyze with Qiniu log tools

### 5. Set Up Alerts

Configure alerts in Qiniu console:
- Storage usage > 80%
- Traffic spike > 200%
- Error rate > 5%

## Performance Optimization

### 1. Enable CDN Cache

```toml
[storage.qiniu]
cache_enabled = true
cache_max_age = 86400  # 24 hours
```

### 2. Image Processing

Optimize images on-the-fly:
```toml
[storage.qiniu]
image_processing = true
image_quality = 85
image_format = "webp"
image_resize_width = 1920
```

### 3. Prefetch Popular Content

```toml
[storage.qiniu]
prefetch_enabled = true
prefetch_threshold = 100  # Views per day
```

### 4. Use Persistent URLs

For frequently accessed documents:
```toml
[storage.qiniu]
use_persistent_urls = true
persistent_url_ttl = 604800  # 7 days
```

## Monitoring and Analytics

### Qiniu Dashboard

Monitor in Qiniu console:
- **Storage**: Total size, file count
- **Traffic**: CDN bandwidth, origin pull
- **Requests**: API calls by type
- **Cost**: Daily/monthly spending

### Cloud Vault Metrics

Check storage stats:
```bash
curl http://localhost:8080/api/v1/system/stats
```

Response:
```json
{
  "storage": {
    "provider": "qiniu",
    "total_size": 1073741824,
    "file_count": 1234,
    "bandwidth_today": 5368709120
  }
}
```

### Third-Party Monitoring

Integrate with monitoring tools:
- **Prometheus**: Scrape `/metrics` endpoint
- **Grafana**: Visualize storage metrics
- **Alertmanager**: Set up alerts

## Backup Strategy

### Qiniu Cross-Region Replication

1. Create backup bucket in different region
2. Enable **Cross-Region Replication**
3. Configure replication rules
4. Automatic backup of new objects

### Manual Backup

```bash
# Download all objects
./cloud-vault backup-qiniu \
  --bucket cloud-vault-docs \
  --output /backups/qiniu-backup-$(date +%Y%m%d)/

# Verify backup
./cloud-vault verify-backup \
  --source qiniu \
  --backup /backups/qiniu-backup-20260530/
```

### Automated Backup Script

```bash
#!/bin/bash
# /usr/local/bin/backup-qiniu.sh

BACKUP_DIR="/backups/qiniu"
DATE=$(date +%Y%m%d)
KEEP_DAYS=30

# Create backup
./cloud-vault backup-qiniu \
  --bucket cloud-vault-docs \
  --output "$BACKUP_DIR/backup-$DATE/"

# Compress backup
tar -czf "$BACKUP_DIR/backup-$DATE.tar.gz" \
  -C "$BACKUP_DIR" "backup-$DATE"

# Remove old backups
find "$BACKUP_DIR" -name "backup-*.tar.gz" -mtime +$KEEP_DAYS -delete

# Log success
logger "Qiniu backup completed: backup-$DATE.tar.gz"
```

Add to crontab:
```bash
# Daily backup at 3 AM
0 3 * * * /usr/local/bin/backup-qiniu.sh
```

## Troubleshooting

### Authentication Errors

**Error**: `401 Unauthorized`

**Solutions**:
1. Verify Access Key and Secret Key
2. Check if keys are active in Qiniu console
3. Ensure no IP restrictions blocking server
4. Regenerate keys if compromised

### Upload Failures

**Error**: `Upload failed: connection timeout`

**Solutions**:
```bash
# Check network connectivity
ping up.qiniup.com
curl -I https://up.qiniup.com

# Increase timeout in config
[storage.qiniu]
upload_timeout_seconds = 300

# Check firewall rules
sudo ufw allow out 443/tcp
```

### CDN Not Accessible

**Error**: `403 Forbidden` or `404 Not Found`

**Solutions**:
1. Verify domain DNS points to Qiniu CDN
2. Check Referer whitelist configuration
3. Ensure SSL certificate is valid
4. Test with direct bucket URL

### Slow Downloads

**Solutions**:
1. Enable CDN if not already enabled
2. Choose region closer to users
3. Enable image processing for optimization
4. Check Qiniu status page for outages

### High Costs

**Solutions**:
1. Review traffic sources in dashboard
2. Enable hotlink protection
3. Optimize image sizes
4. Use Infrequent Access tier for old documents
5. Set up budget alerts

## Advanced Features

### Lifecycle Management

Automatically move old documents to cheaper storage:

```toml
[storage.qiniu.lifecycle]
rules = [
  { days = 90, storage = "infrequent" },
  { days = 365, storage = "archive" },
  { days = 1825, action = "delete" }  # 5 years
]
```

### Version Control

Keep document versions:

```toml
[storage.qiniu]
versioning = true
max_versions = 10
keep_old_versions_days = 30
```

### Event Notifications

Get notified on storage events:

```toml
[storage.qiniu.events]
upload_notification = true
delete_notification = true
webhook_url = "https://vault.yourdomain.com/api/webhooks/qiniu"
```

## Next Steps

- [Storage Configuration](./storage.md) - General storage guide
- [Backup & Restore](./backup-restore.md) - Backup strategies
- [Security](./security.md) - Security best practices
- [Monitoring](./monitoring.md) - Set up monitoring

---

**Last updated**: 2026-05-30
