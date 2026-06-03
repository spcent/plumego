# Privacy Guide

This guide explains how Markdown Cloud Vault handles your data and protects your privacy.

## Overview

Cloud Vault is designed with privacy as a core principle:
- **Local-first**: Data stays on your device by default
- **User control**: You decide what data is shared
- **Transparency**: Clear documentation of data handling
- **No telemetry**: No tracking or analytics by default

## Data Storage

### Local Storage (Default)

When using local storage:
- All documents stored on your device
- Database stored locally (`data/app.db`)
- No data sent to external servers
- Complete control over your data

**Data locations**:
```
Desktop App:
  Windows: %APPDATA%\Cloud Vault\
  macOS: ~/Library/Application Support/Cloud Vault/
  Linux: ~/.config/cloud-vault/

Server:
  /var/lib/cloud-vault/data/
```

### Cloud Storage (Qiniu)

When using Qiniu cloud storage:
- Documents uploaded to your Qiniu bucket
- You control the bucket and access
- Data stored in Qiniu data centers
- Subject to Qiniu privacy policy

**What you control**:
- Bucket location (region)
- Access permissions
- Encryption settings
- Data retention

## Data Collection

### What We Collect

**By default, Cloud Vault collects**:
- Nothing (no telemetry)
- No usage statistics
- No personal information
- No device information

**Optional data collection**:
- Error reports (if enabled)
- Update checks (if enabled)
- AI service requests (if AI enabled)

### What We Don't Collect

- Document content
- Search queries
- User behavior
- Device fingerprints
- IP addresses (unless server logs enabled)

## AI Services

### Data Sent to AI Providers

When AI features are enabled:
- Document content (for summarization, Q&A)
- Document metadata (for tagging)
- User prompts

**Data NOT sent**:
- API keys
- Authentication tokens
- Database content
- File paths
- User information

### AI Provider Privacy

**OpenAI**:
- Data processed according to OpenAI privacy policy
- API data not used for training (as of 2026)
- Data retention: 30 days (unless enterprise plan)

**Azure OpenAI**:
- Data processed in your Azure region
- Subject to Azure privacy policy
- Enterprise-grade privacy controls

**Local LLM (Ollama)**:
- No data leaves your device
- Complete privacy
- No third-party access

### Disable AI Features

To prevent any data from being sent to AI providers:
```toml
[ai]
enabled = false
```

## Diagnostic Bundles

### What's Included

Diagnostic bundles contain:
- System information (OS, memory, disk)
- Configuration (sensitive values redacted)
- Database statistics (not content)
- Recent logs (sensitive data redacted)
- Health check results

### What's Redacted

Automatically redacted from diagnostics:
- API keys
- Passwords
- Tokens
- Secret keys
- Database credentials
- Session tokens

### What's Excluded

Never included in diagnostics:
- Document content
- Search queries
- User data
- Authentication tokens

### User Control

- Diagnostics are user-initiated only
- Never automatically generated
- Never automatically uploaded
- You choose who to share with

## Logs

### Application Logs

**What's logged**:
- Application events
- Error messages
- Performance metrics
- Authentication events

**What's redacted**:
- Passwords
- API keys
- Tokens
- Sensitive data

### Log Retention

```toml
[logging]
max_files = 10
max_size_mb = 100
compress = true
```

### Disable Logging

```toml
[logging]
level = "error"  # Only log errors
# or
enabled = false  # Disable completely
```

## Cookies and Sessions

### Session Cookies

Cloud Vault uses session cookies for authentication:
- **Name**: Configurable (default: `cv_session`)
- **Purpose**: Maintain login state
- **Lifetime**: Configurable (default: 24 hours)
- **Security**: HttpOnly, Secure (if HTTPS)

### Third-Party Cookies

Cloud Vault does not use:
- Tracking cookies
- Analytics cookies
- Advertising cookies
- Social media cookies

## Data Sharing

### No Automatic Sharing

Cloud Vault never automatically shares:
- Document content
- User data
- Usage statistics
- Device information

### Manual Sharing

You control sharing:
- Export documents manually
- Share diagnostic bundles
- Collaborate via collections (future feature)

## Data Deletion

### Delete Documents

Documents are permanently deleted:
- No soft delete
- No recovery after deletion
- Search index updated immediately

### Delete User Data

```bash
# Delete user
./cloud-vault user delete --username user

# Delete all user data
./cloud-vault user delete --username user --purge-data
```

### Delete Everything

```bash
# Stop service
sudo systemctl stop cloud-vault

# Delete all data
rm -rf /var/lib/cloud-vault/data

# Delete configuration
rm -rf /var/lib/cloud-vault/config

# Delete logs
rm -rf /var/log/cloud-vault
```

## GDPR Compliance

### Your Rights

Under GDPR, you have the right to:
- **Access**: Export your data
- **Rectification**: Correct your data
- **Erasure**: Delete your data
- **Portability**: Export in standard formats
- **Restriction**: Limit processing
- **Objection**: Object to processing

### Data Export

Export your data:
```bash
# Export all documents
./cloud-vault export --format markdown --output export.zip

# Export database
./cloud-vault database export --output database.sql

# Export configuration
cp /var/lib/cloud-vault/config/config.toml export/
```

### Data Deletion

Request data deletion:
```bash
# Delete all data
./cloud-vault purge --all
```

## Data Security

### Encryption

**In transit** (if HTTPS enabled):
- TLS 1.2 or higher
- Strong cipher suites
- Certificate validation

**At rest** (if enabled):
- Database encryption
- Backup encryption
- File system encryption

### Access Control

- Role-based permissions
- Session management
- Rate limiting
- Audit logging

## Third-Party Services

### Qiniu Cloud

If using Qiniu storage:
- Data stored in Qiniu infrastructure
- Subject to Qiniu privacy policy
- You control bucket permissions
- Encryption available

### AI Providers

If using AI features:
- Data sent to selected provider
- Subject to provider privacy policy
- Optional: use local LLM for privacy

### Update Server

If update checking enabled:
- Version information sent
- No personal data sent
- Can be disabled

## Privacy Settings

### Recommended Privacy Configuration

```toml
[ai]
enabled = false  # Disable AI to prevent data sharing

[update]
enabled = false  # Disable update checks

[logging]
level = "error"  # Minimal logging
redact_sensitive = true

[diagnostics]
include_logs = false  # Exclude logs from diagnostics
include_config = false  # Exclude config from diagnostics
```

### Maximum Privacy

```toml
[ai]
enabled = false

[update]
enabled = false

[logging]
enabled = false

[diagnostics]
enabled = false

[storage]
provider = "local"  # Use local storage only
```

## Privacy Best Practices

### 1. Use Local Storage

Keep data on your device:
```toml
[storage]
provider = "local"
```

### 2. Disable AI Features

Prevent data from being sent to AI providers:
```toml
[ai]
enabled = false
```

### 3. Review Diagnostic Bundles

Before sharing diagnostics:
1. Extract ZIP file
2. Review contents
3. Verify sensitive data is redacted
4. Share only with trusted parties

### 4. Regular Data Review

Periodically review stored data:
```bash
# List all documents
./cloud-vault documents list

# Check storage usage
./cloud-vault storage stats
```

### 5. Secure Backups

Encrypt backup archives before storing them offsite:

```bash
gpg -c backups/cloud-vault-backup-20260603-120000.zip
```

## Transparency

### Open Source

Cloud Vault is open source:
- Review source code
- Verify privacy claims
- Contribute improvements
- Fork for custom needs

### Privacy Policy Updates

This privacy guide may be updated:
- Check release notes for changes
- Review documentation updates
- Subscribe to release notifications

## Contact

For privacy questions or concerns:
- Review documentation
- Check source code
- Open GitHub issue
- Contact maintainers

## Legal

### Data Controller

You are the data controller:
- You control your data
- You decide how it's processed
- You can delete it anytime

### Data Processor

Cloud Vault is a data processor:
- Processes data on your behalf
- Follows your instructions
- No independent data use

### Jurisdiction

Cloud Vault is developed in compliance with:
- GDPR (European Union)
- CCPA (California)
- PIPEDA (Canada)
- Other applicable privacy laws

## Next Steps

- [Security](./security.md) - Security features
- [Configuration](./configuration.md) - Privacy settings
- [Diagnostics](./diagnostics.md) - Diagnostic privacy

---

**Last updated**: 2026-05-30
